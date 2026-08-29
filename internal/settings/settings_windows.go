package settings

import (
	"strings"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"github.com/JeremyProffittOrg/cli-controller/internal/serial"
	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

const (
	idCombo = 1010
	idSave  = 1
	idCancel = 2
	idBase  = 1001
)

type Dialog struct {
	hwnd   windows.Handle
	checks []windows.Handle
	combo  windows.Handle
	ports  []serial.PortInfo
	cfg    config.Config
	font   windows.Handle
	OnSave func(config.Config)
	OnClose func()
}

var inst *Dialog
var cb = windows.NewCallback(proc)

func New(parent windows.Handle) (*Dialog, error) {
	d := &Dialog{
		checks: make([]windows.Handle, len(config.BrandNames())),
		font:   win32.GetStockFont(),
	}
	if err := win32.RegisterClass("CLIDialSettings", cb, win32.CreateBrush(win32.RGB(255, 255, 255))); err != nil {
		return nil, err
	}
	h, err := win32.CreateWindow(
		win32.WS_EX_APPWINDOW,
		win32.WS_CAPTION|win32.WS_SYSMENU,
		"CLIDialSettings", "CLI Dial Settings",
		200, 200, 460, 520, parent, 0,
	)
	if err != nil {
		return nil, err
	}
	d.hwnd = h
	inst = d
	d.build()
	return d, nil
}

func proc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	h := windows.Handle(hwnd)
	if inst == nil {
		return win32.DefWindowProc(h, msg, wParam, lParam)
	}
	switch msg {
	case win32.WM_COMMAND:
		id := int(win32.LOWORD(wParam))
		if id == idSave {
			inst.save()
			return 0
		}
		if id == idCancel || id == win32.IDCANCEL {
			inst.hide()
			return 0
		}
	case win32.WM_CLOSE:
		inst.hide()
		return 0
	}
	return win32.DefWindowProc(h, msg, wParam, lParam)
}

func (d *Dialog) build() {
	names := config.BrandNames()
	labels := []string{"Cmd", "PowerShell", "Claude", "Grok", "Antigravity", "OpenCode", "Codex", "Unknown"}
	child := func(ex, style uint32, class, title string, x, y, w, h int32, id uintptr) windows.Handle {
		hwnd, err := win32.CreateWindow(ex, style|win32.WS_CHILD|win32.WS_VISIBLE, class, title, x, y, w, h, d.hwnd, windows.Handle(id))
		if err != nil {
			return 0
		}
		win32.Send(hwnd, win32.WM_SETFONT, uintptr(d.font), 1)
		return hwnd
	}
	child(0, 0, "STATIC", "Which CLIs this controls", 24, 16, 400, 22, 0)
	for i, lab := range labels {
		y := int32(44 + i*28)
		d.checks[i] = child(0, win32.BS_AUTOCHECKBOX|win32.WS_TABSTOP, "BUTTON", lab, 32, y, 380, 24, uintptr(idBase+i))
		_ = names
	}
	child(0, 0, "STATIC", "Dial connection", 24, 280, 400, 22, 0)
	d.combo = child(0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 32, 306, 390, 220, idCombo)
	child(0, win32.BS_DEFPUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Save", 230, 430, 90, 28, idSave)
	child(0, win32.BS_PUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Cancel", 332, 430, 90, 28, idCancel)
}

func (d *Dialog) Show(cfg config.Config, ports []serial.PortInfo) {
	d.cfg = cfg
	d.ports = ports
	names := config.BrandNames()
	for i, n := range names {
		if i < len(d.checks) && d.checks[i] != 0 {
			win32.SetCheck(d.checks[i], cfg.Enabled(n))
		}
	}
	win32.ComboReset(d.combo)
	win32.ComboAdd(d.combo, "Automatically find the dial")
	sel := 0
	for i, p := range ports {
		win32.ComboAdd(d.combo, serial.PortLabel(p))
		if cfg.PortMode == "manual" && cfg.Port == p.Name {
			sel = i + 1
		}
	}
	if cfg.PortMode == "auto" {
		sel = 0
	}
	win32.ComboSet(d.combo, sel)
	win32.ShowWindow(d.hwnd, win32.SW_SHOW)
	win32.SetForegroundWindow(d.hwnd)
}

func (d *Dialog) hide() {
	win32.ShowWindow(d.hwnd, win32.SW_HIDE)
	if d.OnClose != nil {
		d.OnClose()
	}
}

func (d *Dialog) save() {
	cfg := d.cfg
	cfg.Brands = config.DefaultBrands()
	names := config.BrandNames()
	for i, n := range names {
		if i < len(d.checks) {
			cfg.Brands[n] = win32.GetCheck(d.checks[i])
		}
	}
	idx := win32.ComboGet(d.combo)
	if idx <= 0 {
		cfg.PortMode = "auto"
		cfg.Port = ""
	} else {
		cfg.PortMode = "manual"
		text := win32.ComboText(d.combo, idx)
		cfg.Port = strings.Fields(strings.Split(text, "—")[0])[0]
		cfg.Port = strings.TrimSpace(cfg.Port)
	}
	if d.OnSave != nil {
		d.OnSave(cfg)
	}
	d.hide()
}

func (d *Dialog) Hwnd() windows.Handle { return d.hwnd }
