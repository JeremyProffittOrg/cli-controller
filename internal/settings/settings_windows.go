package settings

import (
	"fmt"
	"strings"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"github.com/JeremyProffittOrg/cli-controller/internal/overlay"
	"github.com/JeremyProffittOrg/cli-controller/internal/serial"
	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

const (
	idCombo = 1010
	idView  = 1011
	idTheme = 1012
	idSave  = 1
	idCancel = 2
	idBase  = 1001
)

type Dialog struct {
	hwnd      windows.Handle
	checks    []windows.Handle
	combo     windows.Handle
	view      windows.Handle
	theme     windows.Handle
	themePad  windows.Handle
	themeLab  windows.Handle
	rotPad    windows.Handle
	rotLabel  windows.Handle
	ports     []serial.PortInfo
	cfg       config.Config
	angle     int
	drag      bool
	font      windows.Handle
	fontB     windows.Handle
	fontTech  windows.Handle
	bg        windows.Handle
	panel     windows.Handle
	OnSave    func(config.Config)
	OnClose   func()
}

var inst *Dialog
var cb = windows.NewCallback(proc)
var rotCB = windows.NewCallback(rotProc)
var themeCB = windows.NewCallback(themeProc)

func New(parent windows.Handle) (*Dialog, error) {
	d := &Dialog{
		checks:   make([]windows.Handle, len(config.BrandNames())),
		font:     win32.CreateFont(-15, win32.FW_NORMAL, "Segoe UI"),
		fontB:    win32.CreateFont(-18, win32.FW_BOLD, "Bahnschrift"),
		fontTech: win32.CreateFont(-13, win32.FW_BOLD, "Bahnschrift"),
		bg:       win32.CreateBrush(win32.RGB(7, 12, 22)),
		panel:    win32.CreateBrush(win32.RGB(12, 20, 34)),
	}
	if err := win32.RegisterClass("CLIDialSettings", cb, d.bg); err != nil {
		return nil, err
	}
	if err := win32.RegisterClass("CLIDialAnglePad", rotCB, win32.CreateBrush(win32.RGB(8, 14, 28))); err != nil {
		return nil, err
	}
	if err := win32.RegisterClass("CLIDialThemePad", themeCB, win32.CreateBrush(win32.RGB(8, 14, 28))); err != nil {
		return nil, err
	}
	h, err := win32.CreateWindow(
		win32.WS_EX_APPWINDOW,
		win32.WS_CAPTION|win32.WS_SYSMENU,
		"CLIDialSettings", "CLI Dial // Config",
		200, 40, 560, 860, parent, 0,
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
	case win32.WM_ERASEBKGND:
		hdc := windows.Handle(wParam)
		rc := win32.GetClientRect(h)
		win32.FillRect(hdc, &rc, inst.bg)
		return 1
	case win32.WM_PAINT:
		inst.paintChrome(h)
		return 0
	case win32.WM_CTLCOLORSTATIC, win32.WM_CTLCOLORBTN:
		return inst.colorCtl(wParam, false)
	case win32.WM_CTLCOLOREDIT, win32.WM_CTLCOLORLISTBOX:
		return inst.colorCtl(wParam, true)
	case win32.WM_COMMAND:
		id := int(win32.LOWORD(wParam))
		note := int(win32.HIWORD(wParam))
		if note == win32.CBN_SELCHANGE && (id == idTheme || id == idView) {
			inst.syncThemePreview()
			return 0
		}
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

func (d *Dialog) colorCtl(wParam uintptr, field bool) uintptr {
	hdc := windows.Handle(wParam)
	if field {
		win32.SetTextColor(hdc, win32.RGB(226, 232, 240))
		win32.SetBkColor(hdc, win32.RGB(12, 20, 34))
		return uintptr(d.panel)
	}
	win32.SetTextColor(hdc, win32.RGB(125, 211, 252))
	win32.SetBkColor(hdc, win32.RGB(7, 12, 22))
	win32.SetBkMode(hdc, win32.TRANSPARENT)
	return uintptr(d.bg)
}

func (d *Dialog) paintChrome(h windows.Handle) {
	var ps win32.PAINTSTRUCT
	hdc := win32.BeginPaint(h, &ps)
	defer win32.EndPaint(h, &ps)
	rc := win32.GetClientRect(h)
	head := win32.RECT{Left: 0, Top: 0, Right: rc.Right, Bottom: 56}
	win32.FillRect(hdc, &head, d.panel)
	cyan := win32.CreatePen(2, win32.RGB(34, 211, 238))
	old := win32.SelectObject(hdc, cyan)
	win32.MoveTo(hdc, 0, 56)
	win32.LineTo(hdc, rc.Right, 56)
	for x := int32(16); x < rc.Right; x += 28 {
		win32.MoveTo(hdc, x, 56)
		win32.LineTo(hdc, x, 48)
	}
	win32.SelectObject(hdc, old)
	win32.DeleteObject(cyan)
	win32.SetBkMode(hdc, win32.TRANSPARENT)
	win32.SelectObject(hdc, d.fontB)
	win32.SetTextColor(hdc, win32.RGB(34, 211, 238))
	title := win32.RECT{Left: 24, Top: 8, Right: rc.Right - 24, Bottom: 32}
	win32.DrawText(hdc, "CLI DIAL // CONFIG", &title, win32.DT_LEFT|win32.DT_VCENTER|win32.DT_SINGLELINE|win32.DT_NOPREFIX)
	win32.SelectObject(hdc, d.fontTech)
	win32.SetTextColor(hdc, win32.RGB(148, 163, 184))
	sub := win32.RECT{Left: 24, Top: 30, Right: rc.Right - 24, Bottom: 50}
	win32.DrawText(hdc, "HOST CONTROL SURFACE", &sub, win32.DT_LEFT|win32.DT_VCENTER|win32.DT_SINGLELINE|win32.DT_NOPREFIX)
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

func themeProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	h := windows.Handle(hwnd)
	if inst == nil {
		return win32.DefWindowProc(h, msg, wParam, lParam)
	}
	if msg == win32.WM_PAINT {
		inst.paintTheme(h)
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
		win32.SetWindowText(d.rotLabel, fmt.Sprintf("ROTATION // %d DEG", d.angle))
	}
}

func (d *Dialog) selectedThemeID() string {
	idx := win32.ComboGet(d.theme)
	cat := overlay.Catalog()
	if idx < 0 || idx >= len(cat) {
		return overlay.NormalizeTheme("")
	}
	return cat[idx].ID
}

func (d *Dialog) syncThemePreview() {
	th := overlay.ThemeByID(d.selectedThemeID())
	if d.themeLab != 0 {
		win32.SetWindowText(d.themeLab, "THEME // "+strings.ToUpper(th.Name))
	}
	if d.themePad != 0 {
		win32.Invalidate(d.themePad)
	}
}

func (d *Dialog) paintTheme(h windows.Handle) {
	var ps win32.PAINTSTRUCT
	hdc := win32.BeginPaint(h, &ps)
	defer win32.EndPaint(h, &ps)
	rc := win32.GetClientRect(h)
	win32.FillRect(hdc, &rc, d.bg)
	overlay.DrawTheme(hdc, d.selectedThemeID(), rc)
	ring := win32.CreatePen(3, win32.RGB(34, 211, 238))
	oldP := win32.SelectObject(hdc, ring)
	oldB := win32.SelectObject(hdc, win32.NullBrush())
	win32.Ellipse(hdc, 2, 2, rc.Right-2, rc.Bottom-2)
	win32.SelectObject(hdc, oldP)
	win32.SelectObject(hdc, oldB)
	win32.DeleteObject(ring)
}

func (d *Dialog) paintDial(h windows.Handle) {
	var ps win32.PAINTSTRUCT
	hdc := win32.BeginPaint(h, &ps)
	defer win32.EndPaint(h, &ps)
	rc := win32.GetClientRect(h)
	bg := win32.CreateBrush(win32.RGB(8, 14, 28))
	face := win32.CreateBrush(win32.RGB(15, 23, 42))
	win32.FillRect(hdc, &rc, bg)
	cx := rc.Right / 2
	cy := rc.Bottom / 2
	r := rc.Right/2 - 8
	if rc.Bottom/2-8 < r {
		r = rc.Bottom/2 - 8
	}
	ring := win32.CreatePen(3, win32.RGB(34, 211, 238))
	tick := win32.CreatePen(2, win32.RGB(56, 189, 248))
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
	section := func(title string, y int32) {
		hwnd := child(0, 0, "STATIC", title, 24, y, 500, 20, 0)
		if hwnd != 0 {
			win32.Send(hwnd, win32.WM_SETFONT, uintptr(d.fontTech), 1)
		}
	}
	section("01  WHICH CLIS THIS CONTROLS", 68)
	for i, lab := range labels {
		y := int32(92 + i*26)
		d.checks[i] = child(0, win32.BS_AUTOCHECKBOX|win32.WS_TABSTOP, "BUTTON", lab, 36, y, 480, 22, uintptr(idBase+i))
		_ = names
	}
	section("02  DIAL CONNECTION", 308)
	d.combo = child(0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 36, 332, 480, 160, idCombo)
	section("03  ON-SCREEN OVERLAY", 376)
	d.view = child(0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 36, 400, 480, 80, idView)
	section("04  GRAPHICAL THEME", 444)
	d.theme = child(0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 36, 468, 480, 280, idTheme)
	pad, err := win32.CreateWindow(0, win32.WS_CHILD|win32.WS_VISIBLE, "CLIDialThemePad", "", 36, 508, 196, 196, d.hwnd, 0)
	if err == nil {
		d.themePad = pad
		rgn := win32.CreateEllipticRgn(0, 0, 196, 196)
		win32.SetWindowRgn(pad, rgn, true)
	}
	rot, err := win32.CreateWindow(0, win32.WS_CHILD|win32.WS_VISIBLE, "CLIDialAnglePad", "", 300, 508, 196, 196, d.hwnd, 0)
	if err == nil {
		d.rotPad = rot
		rgn := win32.CreateEllipticRgn(0, 0, 196, 196)
		win32.SetWindowRgn(rot, rgn, true)
	}
	d.themeLab = child(0, 0, "STATIC", "THEME // NEON CORE", 36, 712, 220, 20, 0)
	d.rotLabel = child(0, 0, "STATIC", "ROTATION // 0 DEG", 300, 712, 220, 20, 0)
	if d.themeLab != 0 {
		win32.Send(d.themeLab, win32.WM_SETFONT, uintptr(d.fontTech), 1)
	}
	if d.rotLabel != 0 {
		win32.Send(d.rotLabel, win32.WM_SETFONT, uintptr(d.fontTech), 1)
	}
	save := child(0, win32.BS_DEFPUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Commit", 300, 748, 100, 32, idSave)
	cancel := child(0, win32.BS_PUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Abort", 416, 748, 100, 32, idCancel)
	if save != 0 {
		win32.Send(save, win32.WM_SETFONT, uintptr(d.fontB), 1)
	}
	if cancel != 0 {
		win32.Send(cancel, win32.WM_SETFONT, uintptr(d.fontB), 1)
	}
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
	win32.ComboReset(d.theme)
	want := overlay.ThemeIndex(cfg.OverlayTheme)
	for _, th := range overlay.Catalog() {
		win32.ComboAdd(d.theme, th.Name)
	}
	win32.ComboSet(d.theme, want)
	d.updateRotLabel()
	d.syncThemePreview()
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
	cfg.OverlayTheme = d.selectedThemeID()
	cfg.DisplayRotation = NormalizeDeg(d.angle)
	if d.OnSave != nil {
		d.OnSave(cfg)
	}
	d.hide()
}

func (d *Dialog) Hwnd() windows.Handle { return d.hwnd }
