package app

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JeremyProffittOrg/cli-controller/internal/applog"
	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"github.com/JeremyProffittOrg/cli-controller/internal/motion"
	"github.com/JeremyProffittOrg/cli-controller/internal/overlay"
	"github.com/JeremyProffittOrg/cli-controller/internal/protocol"
	"github.com/JeremyProffittOrg/cli-controller/internal/serial"
	"github.com/JeremyProffittOrg/cli-controller/internal/settings"
	"github.com/JeremyProffittOrg/cli-controller/internal/tray"
	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"github.com/JeremyProffittOrg/cli-controller/internal/wins"
	"golang.org/x/sys/windows"
)

type App struct {
	cfg       config.Config
	log       *log.Logger
	host      windows.Handle
	tray      *tray.Tray
	overlay   *overlay.Overlay
	settings  *settings.Dialog
	ser       *serial.Manager
	list      []wins.Window
	sel       int
	connected bool
	port      string
	dwell     *time.Timer
	mu        sync.Mutex
	msgs      []protocol.DeviceMsg
	conns     []connEv
	mutex     windows.Handle
	motion    *motion.Engine
	sensorOK  [5]bool
	inventory []protocol.SensorStatus
	// refreshQueued is set by the serial goroutine when a WM_REFRESH is already
	// posted and cleared by drain, so a 130 msg/s sensor stream cannot flood
	// the UI thread's posted-message queue and starve input and paint.
	refreshQueued atomic.Bool
	// liveDirty marks that sensor samples arrived since the Settings live
	// readouts were last redrawn; the 125 ms timer flushes it.
	liveDirty bool
}

type connEv struct {
	ok   bool
	info serial.PortInfo
}

const SettingsItemTitle = "Settings"

func DialItems(list []wins.Window) []overlay.Item {
	out := make([]overlay.Item, 0, len(list)+1)
	for _, w := range list {
		out = append(out, overlay.Item{Brand: string(w.Brand), Title: w.Title})
	}
	return append(out, overlay.Item{Title: SettingsItemTitle})
}

func isSettingsSelection(sel, windowCount int) bool {
	return sel == windowCount
}

func selectionAfterRefresh(old []wins.Window, sel int, next []wins.Window) int {
	if isSettingsSelection(sel, len(old)) {
		return len(next)
	}
	if sel >= 0 && sel < len(old) {
		hwnd := old[sel].HWND
		for i, w := range next {
			if w.HWND == hwnd {
				return i
			}
		}
	}
	if sel < 0 || sel >= len(next) {
		return 0
	}
	return sel
}

func stateForSelection(list []wins.Window, sel int) (int, string, string) {
	n := len(list) + 1
	if isSettingsSelection(sel, len(list)) {
		return n, "", SettingsItemTitle
	}
	if sel < 0 || sel >= len(list) {
		return n, "", ""
	}
	brand := string(list[sel].Brand)
	if brand == string(wins.BrandUnknown) {
		brand = ""
	}
	return n, brand, list[sel].Title
}

var inst *App
var hostCB = windows.NewCallback(hostProc)

func Run() error {
	name, err := windows.UTF16PtrFromString(`Local\CLIDialController`)
	if err != nil {
		return err
	}
	mh, err := windows.CreateMutex(nil, true, name)
	if err == windows.ERROR_ALREADY_EXISTS {
		if mh != 0 {
			windows.CloseHandle(mh)
		}
		return fmt.Errorf("already running")
	}
	if err != nil {
		return err
	}
	logger, f, err := applog.Open()
	if err != nil {
		logger = log.Default()
	}
	if f != nil {
		defer f.Close()
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a := &App{cfg: cfg, log: logger, mutex: mh, motion: motion.New(cfg)}
	inst = a
	win32.FreeConsole()
	if err := win32.RegisterClass("CLIDialHost", hostCB, 0); err != nil {
		return err
	}
	h, err := win32.CreateWindow(0, win32.WS_POPUP, "CLIDialHost", "CLI Dial", 0, 0, 0, 0, 0, 0)
	if err != nil {
		return err
	}
	a.host = h
	ov, err := overlay.New()
	if err != nil {
		return err
	}
	a.overlay = ov
	a.tray = tray.New(h)
	dlg, err := settings.New(h)
	if err != nil {
		return err
	}
	a.settings = dlg
	dlg.OnSave = a.saveSettings
	dlg.OnScan = a.requestScan
	dlg.OnCalibrate = a.requestCalibrate
	a.restartSerial()
	win32.SetTimer(h, 1, 125)
	win32.SetTimer(h, 2, 1000)
	win32.SetTimer(h, 3, 500)
	a.log.Printf("started")
	var msg win32.MSG
	for {
		r := win32.GetMessage(&msg)
		if r <= 0 {
			break
		}
		win32.Translate(&msg)
		win32.Dispatch(&msg)
	}
	a.shutdown()
	return nil
}

func hostProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	h := windows.Handle(hwnd)
	if inst == nil {
		return win32.DefWindowProc(h, msg, wParam, lParam)
	}
	switch msg {
	case win32.WM_TRAY:
		switch lParam {
		case win32.WM_RBUTTONUP, win32.WM_CONTEXTMENU:
			cmd := inst.tray.Popup()
			if cmd == tray.IDSettings {
				inst.openSettings()
			}
			if cmd == tray.IDExit {
				win32.PostQuit(0)
			}
		case win32.WM_LBUTTONUP:
			inst.tray.Balloon()
		}
		return 0
	case win32.WM_TIMER:
		switch wParam {
		case 1:
			inst.tray.Tick()
			inst.applyMotion(inst.motion.Tick(time.Now()))
			if inst.liveDirty {
				inst.liveDirty = false
				inst.settings.SetLive(inst.inventory)
			}
		case 2:
			if inst.connected {
				inst.sendState()
			}
		case 3:
			if inst.overlay.Visible() {
				inst.refresh()
				inst.showOverlay()
			}
		}
		return 0
	case win32.WM_REFRESH:
		inst.drain()
		return 0
	case win32.WM_DWELL:
		inst.activate()
		return 0
	case win32.WM_OPEN_SETTINGS:
		inst.openSettings()
		return 0
	case win32.WM_DESTROY:
		win32.PostQuit(0)
		return 0
	}
	return win32.DefWindowProc(h, msg, wParam, lParam)
}

func (a *App) restartSerial() {
	if a.ser != nil {
		a.ser.Stop()
	}
	a.ser = serial.NewManager(func() config.Config { return a.cfg })
	a.ser.OnMsg = func(m protocol.DeviceMsg) {
		a.mu.Lock()
		a.msgs = append(a.msgs, m)
		a.mu.Unlock()
		if a.refreshQueued.CompareAndSwap(false, true) {
			win32.Post(a.host, win32.WM_REFRESH, 0, 0)
		}
	}
	a.ser.OnConn = func(ok bool, info serial.PortInfo) {
		a.mu.Lock()
		a.conns = append(a.conns, connEv{ok, info})
		a.mu.Unlock()
		if a.refreshQueued.CompareAndSwap(false, true) {
			win32.Post(a.host, win32.WM_REFRESH, 0, 0)
		}
	}
	go a.ser.Run()
}

func (a *App) drain() {
	a.refreshQueued.Store(false)
	a.mu.Lock()
	msgs := a.msgs
	a.msgs = nil
	conns := a.conns
	a.conns = nil
	a.mu.Unlock()
	for _, c := range conns {
		a.handleConn(c.ok, c.info)
	}
	for _, m := range msgs {
		a.handleMsg(m)
	}
}

func (a *App) handleConn(ok bool, info serial.PortInfo) {
	a.connected = ok
	a.port = info.Name
	a.tray.SetConnected(ok, info.Name)
	if ok {
		a.log.Printf("connected %s serial %s", info.Name, info.SerialNumber)
		if info.SerialNumber != "" && info.SerialNumber != a.cfg.LastSerial {
			a.cfg.LastSerial = info.SerialNumber
			_ = config.Save(a.cfg)
		}
		a.refresh()
		a.sendState()
		a.requestScan()
	} else {
		a.log.Printf("disconnected")
		a.stopDwell()
		a.overlay.Hide()
		for i := range a.inventory {
			a.inventory[i].OK = false
			a.applyMotion(a.motion.Connected(a.inventory[i].ID, false))
		}
		a.sensorOK = [5]bool{}
		a.settings.SetInventory(a.inventory)
		a.settings.SetSensorStatus(a.sensorOK)
	}
}

func (a *App) handleMsg(m protocol.DeviceMsg) {
	switch m.T {
	case "enc":
		a.onEnc(m.D)
	case "tap":
		a.onTap(m.ID)
	case "btn":
		if m.ID == "a" {
			a.activate()
		}
	case "sensor":
		a.upsertInventory(m)
		a.applyMotion(a.motion.Connected(m.SensorID(), m.OK))
		a.settings.SetInventory(a.inventory)
		a.settings.SetSensorStatus(a.sensorOK)
	case "tof":
		a.applySample(m)
		if m.St == 0 {
			a.applyMotion(a.motion.Distance(m.SensorID(), m.MM, time.Now()))
		}
		a.liveDirty = true
	case "accel":
		a.applySample(m)
		a.applyMotion(a.motion.Accel(m.SensorID(), m.X, m.Y, m.Z, time.Now()))
		a.liveDirty = true
	case "scan":
		a.settings.SetInventory(a.inventory)
	case "i2c":
		a.settings.SetI2CPort(m.Port, m.Sda, m.Scl)
	case "log":
		a.log.Printf("dial: %s", m.Msg)
	case "cal":
		a.log.Printf("dial cal %s ok=%v offset=%d avg=%d n=%d", m.SensorID(), m.OK, m.Offset, m.Avg, m.N)
		a.settings.SetCalResult(m.SensorID(), m.OK, m.Offset, m.Avg, m.N)
	}
}

func (a *App) upsertInventory(m protocol.DeviceMsg) {
	st := protocol.SensorStatus{
		ID:   m.SensorID(),
		Kind: m.Kind,
		Mux:  m.Mux,
		Ch:   m.Ch,
		Addr: m.Addr,
		Chip: m.Chip,
		Err:  m.Err,
		Init: m.Init,
		OK:   m.OK,
	}
	if st.Kind == "" {
		if m.T == "accel" {
			st.Kind = "accel"
		} else {
			st.Kind = "tof"
		}
	}
	for i := range a.inventory {
		if a.inventory[i].ID == st.ID {
			prev := a.inventory[i]
			st.Live = prev.Live
			st.MM = prev.MM
			st.St = prev.St
			st.Sig = prev.Sig
			st.Amb = prev.Amb
			st.Sg = prev.Sg
			if st.Addr == 0 {
				st.Addr = prev.Addr
			}
			if st.Chip == "" {
				st.Chip = prev.Chip
			}
			st.X = prev.X
			st.Y = prev.Y
			st.Z = prev.Z
			a.inventory[i] = st
			a.syncLegacyStatus()
			return
		}
	}
	a.log.Printf("sensor new %s kind=%s mux=%d ch=%d addr=%d chip=%s ok=%v", st.ID, st.Kind, st.Mux, st.Ch, st.Addr, st.Chip, st.OK)
	a.inventory = append(a.inventory, st)
	a.syncLegacyStatus()
}

func (a *App) applySample(m protocol.DeviceMsg) {
	id := m.SensorID()
	kind := m.Kind
	if kind == "" {
		if m.T == "accel" {
			kind = "accel"
		} else {
			kind = "tof"
		}
	}
	st := protocol.SensorStatus{ID: id, Kind: kind, Mux: m.Mux, Ch: m.Ch, OK: true, Live: true, MM: m.MM, St: m.St, Sig: m.Sig, Amb: m.Amb, Sg: m.Sg, X: m.X, Y: m.Y, Z: m.Z}
	for i := range a.inventory {
		if a.inventory[i].ID == id {
			cur := a.inventory[i]
			cur.OK = true
			cur.Live = true
			cur.Kind = kind
			if kind == "accel" {
				cur.X, cur.Y, cur.Z = m.X, m.Y, m.Z
			} else {
				cur.MM = m.MM
				cur.St = m.St
				cur.Sig = m.Sig
				cur.Amb = m.Amb
				cur.Sg = m.Sg
			}
			a.inventory[i] = cur
			a.syncLegacyStatus()
			return
		}
	}
	a.log.Printf("sensor new %s from %s sample mux=%d ch=%d", st.ID, m.T, st.Mux, st.Ch)
	a.inventory = append(a.inventory, st)
	a.syncLegacyStatus()
}

func (a *App) syncLegacyStatus() {
	var status [5]bool
	for _, s := range a.inventory {
		if !s.OK {
			continue
		}
		if s.Kind == "accel" && (s.ID == protocol.FormatSensorID(0x70, 4, "accel") || s.Ch == 4) {
			status[4] = true
			continue
		}
		if s.Kind == "tof" && s.Ch >= 0 && s.Ch < 4 && (s.Mux == 0x70 || s.Mux == 0) {
			status[s.Ch] = true
		}
	}
	a.sensorOK = status
}

func (a *App) requestCalibrate(id string, mm int) {
	if a.ser == nil {
		return
	}
	b, err := protocol.Calibrate(id, mm)
	if err != nil {
		return
	}
	a.log.Printf("calibrate %s at %d mm", id, mm)
	_ = a.ser.Send(b)
}

func (a *App) requestScan() {
	if a.ser == nil {
		return
	}
	b, err := protocol.Scan()
	if err != nil {
		return
	}
	_ = a.ser.Send(b)
}

func (a *App) applyMotion(events []motion.Event) {
	for _, event := range events {
		switch event.Kind {
		case motion.Show:
			a.refresh()
			a.showOverlay()
			a.sendState()
		case motion.Move:
			a.refresh()
			a.sel = overlay.Step(len(a.list)+1, a.sel, event.Delta)
			a.showOverlay()
			a.sendState()
		case motion.Activate:
			a.activate()
		case motion.Tile:
			a.onTap("tile")
		case motion.Stack:
			a.onTap("stack")
		}
	}
}

func (a *App) items() []overlay.Item {
	return DialItems(a.list)
}

func (a *App) refresh() {
	next := wins.Enumerate(a.cfg)
	a.sel = selectionAfterRefresh(a.list, a.sel, next)
	a.list = next
}

func (a *App) onEnc(delta int) {
	if !a.connected {
		return
	}
	a.sel = overlay.Step(len(a.list)+1, a.sel, delta)
	a.showOverlay()
	a.resetDwell()
	a.sendState()
}

func (a *App) showOverlay() {
	a.overlay.SetView(a.cfg.OverlayView)
	a.overlay.SetTheme(a.cfg.OverlayTheme)
	a.overlay.Show(wins.PrimaryWorkArea(), a.items(), a.sel)
}

func (a *App) onTap(id string) {
	a.stopDwell()
	a.overlay.Hide()
	a.refresh()
	switch id {
	case "tile":
		wins.Tile(a.list)
	case "stack":
		wins.Stack(a.list)
	}
}

func (a *App) activate() {
	a.stopDwell()
	a.refresh()
	a.overlay.Hide()
	if isSettingsSelection(a.sel, len(a.list)) {
		a.openSettings()
		return
	}
	if a.sel < 0 || a.sel >= len(a.list) {
		return
	}
	_ = wins.Focus(a.list[a.sel].HWND)
}

func (a *App) resetDwell() {
	a.stopDwell()
	ms := a.cfg.DwellMs
	if ms <= 0 {
		ms = 2000
	}
	h := a.host
	a.dwell = time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
		win32.Post(h, win32.WM_DWELL, 0, 0)
	})
}

func (a *App) stopDwell() {
	if a.dwell != nil {
		a.dwell.Stop()
		a.dwell = nil
	}
}

func (a *App) sendState() {
	if a.ser == nil {
		return
	}
	sel := a.sel
	n, brand, title := stateForSelection(a.list, sel)
	b, err := protocol.StateRot(a.connected, n, sel, brand, title, a.cfg.DisplayRotation)
	if err != nil {
		return
	}
	_ = a.ser.Send(b)
}

func (a *App) openSettings() {
	ports, err := serial.ListPresent(serial.DefaultLister)
	if err != nil {
		ports = nil
	}
	a.settings.SetInventory(a.inventory)
	a.settings.SetSensorStatus(a.sensorOK)
	a.settings.Show(a.cfg, ports)
	a.requestScan()
}

func (a *App) saveSettings(cfg config.Config) {
	portChanged := cfg.PortMode != a.cfg.PortMode || cfg.Port != a.cfg.Port
	a.cfg = cfg
	a.motion.Configure(cfg)
	if err := config.Save(cfg); err != nil {
		a.log.Printf("save config: %v", err)
	}
	a.refresh()
	if portChanged {
		a.restartSerial()
		return
	}
	a.sendState()
}

func (a *App) shutdown() {
	a.stopDwell()
	if a.ser != nil {
		a.ser.Stop()
	}
	if a.tray != nil {
		a.tray.Close()
	}
	if a.mutex != 0 {
		windows.CloseHandle(a.mutex)
	}
}
