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
	cfg        config.Config
	log        *log.Logger
	host       windows.Handle
	tray       *tray.Tray
	overlay    *overlay.Overlay
	settings   *settings.Dialog
	ser        *serial.Manager
	list       []wins.Window
	sel        int
	connected  bool
	port       string
	dwell      *time.Timer
	mu         sync.Mutex
	msgs       []protocol.DeviceMsg
	conns      []connEv
	mutex      windows.Handle
}

type connEv struct {
	ok   bool
	info serial.PortInfo
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
				inst.overlay.Show(wins.PrimaryWorkArea(), inst.items(), inst.sel)
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
	out := make([]overlay.Item, len(a.list))
	for i, w := range a.list {
		out[i] = overlay.Item{Brand: string(w.Brand), Title: w.Title}
	}
	return out
}

func (a *App) refresh() {
	var old windows.Handle
	if a.sel >= 0 && a.sel < len(a.list) {
		old = a.list[a.sel].HWND
	}
	a.list = wins.Enumerate(a.cfg)
	if old != 0 {
		for i, w := range a.list {
			if w.HWND == old {
				a.sel = i
				return
			}
		}
	}
	if len(a.list) == 0 {
		a.sel = 0
		return
	}
	if a.sel >= len(a.list) {
		a.sel = 0
	}
}

func (a *App) onEnc(delta int) {
	if !a.connected {
		return
	}
	a.refresh()
	a.sel = overlay.Step(len(a.list), a.sel, delta)
	a.overlay.Show(wins.PrimaryWorkArea(), a.items(), a.sel)
	a.resetDwell()
	a.sendState()
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
	n := len(a.list)
	sel := a.sel
	brand, title := "", ""
	if n > 0 && sel >= 0 && sel < n {
		brand = string(a.list[sel].Brand)
		title = a.list[sel].Title
	}
	b, err := protocol.State(a.connected, n, sel, brand, title)
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
	a.cfg = cfg
	if err := config.Save(cfg); err != nil {
		a.log.Printf("save config: %v", err)
	}
	a.refresh()
	a.restartSerial()
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
