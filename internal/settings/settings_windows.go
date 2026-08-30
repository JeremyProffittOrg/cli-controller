package settings

import (
	"fmt"
	"strings"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"github.com/JeremyProffittOrg/cli-controller/internal/serial"
	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

const (
	idCombo = 1010
	idView  = 1011
	idSave  = 1
	idCancel = 2
	idBase  = 1001
)

type Dialog struct {
	hwnd     windows.Handle
	checks   []windows.Handle
	combo    windows.Handle
	view     windows.Handle
	rotPad   windows.Handle
	rotLabel windows.Handle
	ports    []serial.PortInfo
	cfg      config.Config
	angle    int
	drag     bool
	font     windows.Handle
	bg       windows.Handle
	OnSave   func(config.Config)
	OnClose  func()
}

var inst *Dialog
var cb = windows.NewCallback(proc)
var rotCB = windows.NewCallback(rotProc)

func New(parent windows.Handle) (*Dialog, error) {
	d := &Dialog{
		checks: make([]windows.Handle, len(config.BrandNames())),
		font:   win32.GetStockFont(),
		bg:     win32.CreateBrush(win32.RGB(255, 255, 255)),
	}
	if err := win32.RegisterClass("CLIDialSettings", cb, d.bg); err != nil {
		return nil, err
	}
	if err := win32.RegisterClass("CLIDialAnglePad", rotCB, win32.CreateBrush(win32.RGB(15, 23, 42))); err != nil {
		return nil, err
	}
	h, err := win32.CreateWindow(
		win32.WS_EX_APPWINDOW,
		win32.WS_CAPTION|win32.WS_SYSMENU,
		"CLIDialSettings", "CLI Dial Settings",
		200, 80, 480, 760, parent, 0,
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

func rotProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	h := windows.Handle(hwnd)
	if inst == nil {
		return win32.DefWindowProc(h, msg, wParam, lParam)
	}
	switch msg {
	case win32.WM_PAINT:
		inst.paintDial(h)
		return 0
	case win32.WM_LBUTTONDOWN:
		win32.SetCapture(h)
		inst.drag = true
		inst.setAngleFromMouse(h, lParam)
		return 0
	case win32.WM_MOUSEMOVE:
		if inst.drag && wParam&win32.MK_LBUTTON != 0 {
			inst.setAngleFromMouse(h, lParam)
		}
		return 0
	case win32.WM_LBUTTONUP:
		if inst.drag {
			inst.setAngleFromMouse(h, lParam)
			inst.drag = false
			win32.ReleaseCapture()
		}
		return 0
	}
	return win32.DefWindowProc(h, msg, wParam, lParam)
}

func (d *Dialog) setAngleFromMouse(h windows.Handle, lParam uintptr) {
	rc := win32.GetClientRect(h)
	cx := int(rc.Right / 2)
	cy := int(rc.Bottom / 2)
	x, y := win32.MouseXY(lParam)
	d.angle = AngleFromPoint(cx, cy, int(x), int(y))
	d.updateRotLabel()
	win32.Invalidate(h)
}

func (d *Dialog) updateRotLabel() {
	if d.rotLabel != 0 {
		win32.SetWindowText(d.rotLabel, fmt.Sprintf("%d degrees", d.angle))
	}
}

func (d *Dialog) paintDial(h windows.Handle) {
	var ps win32.PAINTSTRUCT
	hdc := win32.BeginPaint(h, &ps)
	defer win32.EndPaint(h, &ps)
	rc := win32.GetClientRect(h)
	bg := win32.CreateBrush(win32.RGB(15, 23, 42))
	face := win32.CreateBrush(win32.RGB(30, 41, 59))
	win32.FillRect(hdc, &rc, bg)
	cx := rc.Right / 2
	cy := rc.Bottom / 2
	r := rc.Right/2 - 10
	if rc.Bottom/2-10 < r {
		r = rc.Bottom/2 - 10
	}
	ring := win32.CreatePen(3, win32.RGB(56, 189, 248))
	tick := win32.CreatePen(2, win32.RGB(148, 163, 184))
	needle := win32.CreatePen(4, win32.RGB(250, 204, 21))
	oldP := win32.SelectObject(hdc, ring)
	oldB := win32.SelectObject(hdc, face)
	win32.Ellipse(hdc, cx-r, cy-r, cx+r, cy+r)
	win32.SelectObject(hdc, tick)
	for i := 0; i < 12; i++ {
		deg := i * 30
		x0, y0 := NeedleEnd(int(cx), int(cy), int(r-4), deg)
		x1, y1 := NeedleEnd(int(cx), int(cy), int(r-16), deg)
		win32.MoveTo(hdc, int32(x0), int32(y0))
		win32.LineTo(hdc, int32(x1), int32(y1))
	}
	nx, ny := NeedleEnd(int(cx), int(cy), int(r-12), d.angle)
	win32.SelectObject(hdc, needle)
	win32.MoveTo(hdc, cx, cy)
	win32.LineTo(hdc, int32(nx), int32(ny))
	hub := win32.CreateBrush(win32.RGB(250, 204, 21))
	win32.SelectObject(hdc, hub)
	win32.Ellipse(hdc, cx-6, cy-6, cx+6, cy+6)
	win32.SelectObject(hdc, oldP)
	win32.SelectObject(hdc, oldB)
	win32.DeleteObject(bg)
	win32.DeleteObject(face)
	win32.DeleteObject(ring)
	win32.DeleteObject(tick)
	win32.DeleteObject(needle)
	win32.DeleteObject(hub)
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
	child(0, 0, "STATIC", "Dial connection", 24, 268, 400, 22, 0)
	d.combo = child(0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 32, 292, 400, 140, idCombo)
	child(0, 0, "STATIC", "On-screen overlay", 24, 336, 400, 22, 0)
	d.view = child(0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 32, 360, 400, 80, idView)
	child(0, 0, "STATIC", "Dial display rotation (drag the line)", 24, 404, 420, 22, 0)
	pad, err := win32.CreateWindow(0, win32.WS_CHILD|win32.WS_VISIBLE, "CLIDialAnglePad", "", 32, 430, 200, 200, d.hwnd, 0)
	if err == nil {
		d.rotPad = pad
	}
	d.rotLabel = child(0, 0, "STATIC", "0 degrees", 250, 510, 180, 24, 0)
	child(0, win32.BS_DEFPUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Save", 250, 660, 90, 28, idSave)
	child(0, win32.BS_PUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Cancel", 352, 660, 90, 28, idCancel)
}

func (d *Dialog) Show(cfg config.Config, ports []serial.PortInfo) {
	d.cfg = cfg
	d.ports = ports
	d.angle = NormalizeDeg(cfg.DisplayRotation)
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
	win32.ComboReset(d.view)
	win32.ComboAdd(d.view, "Classic list")
	win32.ComboAdd(d.view, "Graphical dial")
	vsel := 0
	if cfg.OverlayView == "graphical" {
		vsel = 1
	}
	win32.ComboSet(d.view, vsel)
	d.updateRotLabel()
	if d.rotPad != 0 {
		win32.Invalidate(d.rotPad)
	}
	win32.ShowWindow(d.hwnd, win32.SW_SHOW)
	win32.SetForegroundWindow(d.hwnd)
}

func (d *Dialog) hide() {
	d.drag = false
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
	if win32.ComboGet(d.view) == 1 {
		cfg.OverlayView = "graphical"
	} else {
		cfg.OverlayView = "classic"
	}
	cfg.DisplayRotation = NormalizeDeg(d.angle)
	if d.OnSave != nil {
		d.OnSave(cfg)
	}
	d.hide()
}

func (d *Dialog) Hwnd() windows.Handle { return d.hwnd }
