package app

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/JeremyProffittOrg/cli-controller/internal/applog"
	"github.com/JeremyProffittOrg/cli-controller/internal/config"
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
	a := &App{cfg: cfg, log: logger, mutex: mh}
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
		win32.Post(a.host, win32.WM_REFRESH, 0, 0)
	}
	a.ser.OnConn = func(ok bool, info serial.PortInfo) {
		a.mu.Lock()
		a.conns = append(a.conns, connEv{ok, info})
		a.mu.Unlock()
		win32.Post(a.host, win32.WM_REFRESH, 0, 0)
	}
	go a.ser.Run()
}

func (a *App) drain() {
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
	} else {
		a.log.Printf("disconnected")
		a.stopDwell()
		a.overlay.Hide()
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
	a.refresh()
	a.sel = overlay.Step(len(a.list)+1, a.sel, delta)
	a.showOverlay()
	a.resetDwell()
	a.sendState()
}

func (a *App) showOverlay() {
	a.overlay.SetView(a.cfg.OverlayView)
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
	a.settings.Show(a.cfg, ports)
}

func (a *App) saveSettings(cfg config.Config) {
	portChanged := cfg.PortMode != a.cfg.PortMode || cfg.Port != a.cfg.Port
	a.cfg = cfg
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
