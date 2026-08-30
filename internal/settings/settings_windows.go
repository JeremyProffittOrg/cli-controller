package settings

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"github.com/JeremyProffittOrg/cli-controller/internal/overlay"
	"github.com/JeremyProffittOrg/cli-controller/internal/serial"
	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

const (
	idSave   = 1
	idCancel = 2
	idTab    = 100
	idPort   = 101
	idDwell  = 102
	idView   = 103
	idTheme  = 104
)

type Dialog struct {
	hwnd, tab                                                 windows.Handle
	pages                                                     [4][]windows.Handle
	checks                                                    []windows.Handle
	port, dwell, view, theme, rotation                        windows.Handle
	kneeMode, leftRaises, rightDirection                      windows.Handle
	kneeRole, kneeThreshold, kneeStatus                       [4]windows.Handle
	deskEnabled, deskStatus, deskOrientation, deskSensitivity windows.Handle
	deskAction                                                [4]windows.Handle
	ports                                                     []serial.PortInfo
	cfg                                                       config.Config
	sensorOK                                                  [5]bool
	font, fontB, fontTech, bg, panel                          windows.Handle
	OnSave                                                    func(config.Config)
	OnClose                                                   func()
}

var inst *Dialog
var cb = windows.NewCallback(proc)

func New(parent windows.Handle) (*Dialog, error) {
	win32.InitTabs()
	d := &Dialog{checks: make([]windows.Handle, len(config.BrandNames())), font: win32.CreateFont(-15, win32.FW_NORMAL, "Segoe UI"), fontB: win32.CreateFont(-18, win32.FW_BOLD, "Bahnschrift"), fontTech: win32.CreateFont(-13, win32.FW_BOLD, "Bahnschrift"), bg: win32.CreateBrush(win32.RGB(7, 12, 22)), panel: win32.CreateBrush(win32.RGB(12, 20, 34))}
	if err := win32.RegisterClass("CLIDialSettings", cb, d.bg); err != nil {
		return nil, err
	}
	h, err := win32.CreateWindow(win32.WS_EX_APPWINDOW, win32.WS_CAPTION|win32.WS_SYSMENU, "CLIDialSettings", "CLI Dial // Config", 180, 70, 680, 710, parent, 0)
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
		rc := win32.GetClientRect(h)
		win32.FillRect(windows.Handle(wParam), &rc, inst.bg)
		return 1
	case win32.WM_PAINT:
		inst.paintChrome(h)
		return 0
	case win32.WM_CTLCOLORSTATIC, win32.WM_CTLCOLORBTN:
		return inst.colorCtl(wParam, false)
	case win32.WM_CTLCOLOREDIT, win32.WM_CTLCOLORLISTBOX:
		return inst.colorCtl(wParam, true)
	case win32.WM_NOTIFY:
		if lParam != 0 {
			n := (*win32.NMHDR)(unsafe.Pointer(lParam))
			if n.IDFrom == idTab && n.Code == win32.TCN_SELCHANGE {
				inst.showPage(win32.TabGet(inst.tab))
				return 0
			}
		}
	case win32.WM_COMMAND:
		switch int(win32.LOWORD(wParam)) {
		case idSave:
			inst.save()
			return 0
		case idCancel:
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
	pen := win32.CreatePen(2, win32.RGB(34, 211, 238))
	old := win32.SelectObject(hdc, pen)
	win32.MoveTo(hdc, 0, 56)
	win32.LineTo(hdc, rc.Right, 56)
	win32.SelectObject(hdc, old)
	win32.DeleteObject(pen)
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

func (d *Dialog) build() {
	child := func(page int, ex, style uint32, class, title string, x, y, w, h int32, id uintptr) windows.Handle {
		hwnd, err := win32.CreateWindow(ex, style|win32.WS_CHILD|win32.WS_VISIBLE, class, title, x, y, w, h, d.hwnd, windows.Handle(id))
		if err != nil {
			return 0
		}
		win32.Send(hwnd, win32.WM_SETFONT, uintptr(d.font), 1)
		if page >= 0 {
			d.pages[page] = append(d.pages[page], hwnd)
		}
		return hwnd
	}
	label := func(page int, text string, x, y, w int32) windows.Handle {
		return child(page, 0, 0, "STATIC", text, x, y, w, 20, 0)
	}
	section := func(page int, text string, y int32) {
		h := label(page, text, 36, y, 580)
		win32.Send(h, win32.WM_SETFONT, uintptr(d.fontTech), 1)
	}
	d.tab = child(-1, 0, win32.WS_TABSTOP, "SysTabControl32", "", 24, 66, 632, 34, idTab)
	for i, name := range []string{"Controller", "Display", "Knees", "Desk"} {
		win32.TabInsert(d.tab, i, name)
	}
	section(0, "01  WHICH CLIS THIS CONTROLS", 112)
	labels := []string{"Cmd", "PowerShell", "Claude", "Grok", "Antigravity", "OpenCode", "Codex", "Unknown"}
	for i, text := range labels {
		x := int32(48 + (i%2)*280)
		y := int32(140 + (i/2)*30)
		d.checks[i] = child(0, 0, win32.BS_AUTOCHECKBOX|win32.WS_TABSTOP, "BUTTON", text, x, y, 240, 24, uintptr(200+i))
	}
	section(0, "02  DIAL CONNECTION", 274)
	d.port = child(0, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 48, 300, 560, 180, idPort)
	section(0, "03  ACTIVATION DELAY", 354)
	d.dwell = child(0, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 48, 380, 240, 180, idDwell)
	label(0, "Used by the physical Dial and knee gestures.", 310, 384, 300)
	section(1, "01  OVERLAY TYPE", 112)
	d.view = child(1, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 48, 140, 560, 120, idView)
	section(1, "02  GRAPHICAL THEME", 210)
	d.theme = child(1, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 48, 238, 560, 300, idTheme)
	section(1, "03  DISPLAY ROTATION", 308)
	d.rotation = child(1, win32.WS_EX_CLIENTEDGE, win32.WS_TABSTOP, "EDIT", "0", 48, 336, 160, 26, 105)
	label(1, "Degrees. Values wrap into the 0-359 range.", 230, 340, 370)
	section(2, "01  KNEE GESTURE", 112)
	label(2, "Mode", 48, 140, 120)
	d.kneeMode = child(2, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 170, 136, 220, 100, 110)
	label(2, "Left raises", 48, 178, 120)
	d.leftRaises = child(2, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 170, 174, 100, 100, 111)
	label(2, "Right movement", 300, 178, 130)
	d.rightDirection = child(2, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 430, 174, 178, 100, 112)
	section(2, "02  PCA9548 CHANNELS 0-3", 226)
	label(2, "Channel", 48, 254, 65)
	label(2, "Role", 125, 254, 150)
	label(2, "Threshold (mm)", 292, 254, 130)
	label(2, "Hardware", 455, 254, 130)
	for i := 0; i < 4; i++ {
		y := int32(282 + i*56)
		label(2, fmt.Sprintf("CH %d", i), 48, y+4, 65)
		d.kneeRole[i] = child(2, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 125, y, 145, 100, uintptr(120+i))
		d.kneeThreshold[i] = child(2, win32.WS_EX_CLIENTEDGE, win32.WS_TABSTOP, "EDIT", "75", 292, y, 125, 26, uintptr(130+i))
		d.kneeStatus[i] = label(2, "Not detected", 455, y+4, 150)
	}
	section(3, "01  DESK MOTION SENSOR", 112)
	d.deskEnabled = child(3, 0, win32.BS_AUTOCHECKBOX|win32.WS_TABSTOP, "BUTTON", "Enable ADXL345 desk motion", 48, 140, 280, 24, 140)
	d.deskStatus = label(3, "CH 4 // Not detected", 390, 142, 220)
	label(3, "Board orientation", 48, 188, 150)
	d.deskOrientation = child(3, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 210, 184, 160, 120, 141)
	label(3, "Sensitivity (milli-g)", 48, 230, 150)
	d.deskSensitivity = child(3, win32.WS_EX_CLIENTEDGE, win32.WS_TABSTOP, "EDIT", "350", 210, 226, 160, 26, 142)
	section(3, "02  DIRECTION ACTIONS", 286)
	dirs := []string{"Left", "Right", "Forward", "Back"}
	for i, name := range dirs {
		y := int32(320 + i*52)
		label(3, name, 48, y+4, 120)
		d.deskAction[i] = child(3, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 180, y, 240, 100, uintptr(150+i))
	}
	save := child(-1, 0, win32.BS_DEFPUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Commit", 430, 625, 100, 32, idSave)
	cancel := child(-1, 0, win32.BS_PUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Abort", 546, 625, 100, 32, idCancel)
	win32.Send(save, win32.WM_SETFONT, uintptr(d.fontB), 1)
	win32.Send(cancel, win32.WM_SETFONT, uintptr(d.fontB), 1)
	d.showPage(0)
}

func (d *Dialog) showPage(selected int) {
	if selected < 0 || selected > 3 {
		selected = 0
	}
	for p, handles := range d.pages {
		mode := win32.SW_HIDE
		if p == selected {
			mode = win32.SW_SHOW
		}
		for _, h := range handles {
			win32.ShowWindow(h, mode)
		}
	}
}
func fillCombo(h windows.Handle, values []string, selected int) {
	win32.ComboReset(h)
	for _, v := range values {
		win32.ComboAdd(h, v)
	}
	win32.ComboSet(h, selected)
}
func indexOf(values []string, want string) int {
	for i, v := range values {
		if v == want {
			return i
		}
	}
	return 0
}

func (d *Dialog) Show(cfg config.Config, ports []serial.PortInfo) {
	cfg.Normalize()
	d.cfg = cfg
	d.ports = ports
	for i, n := range config.BrandNames() {
		win32.SetCheck(d.checks[i], cfg.Enabled(n))
	}
	win32.ComboReset(d.port)
	win32.ComboAdd(d.port, "Automatically find the dial")
	portSel := 0
	for i, p := range ports {
		win32.ComboAdd(d.port, serial.PortLabel(p))
		if cfg.PortMode == "manual" && cfg.Port == p.Name {
			portSel = i + 1
		}
	}
	win32.ComboSet(d.port, portSel)
	delays := []int{250, 500, 750, 1000, 1500, 2000}
	delayLabels := make([]string, len(delays))
	delaySel := 0
	for i, v := range delays {
		delayLabels[i] = fmt.Sprintf("%d ms", v)
		if v == cfg.DwellMs {
			delaySel = i
		}
	}
	fillCombo(d.dwell, delayLabels, delaySel)
	viewSel := 0
	if cfg.OverlayView == "graphical" {
		viewSel = 1
	}
	fillCombo(d.view, []string{"Classic list", "Graphical dial"}, viewSel)
	themes := overlay.Catalog()
	themeLabels := make([]string, len(themes))
	for i, t := range themes {
		themeLabels[i] = t.Name
	}
	fillCombo(d.theme, themeLabels, overlay.ThemeIndex(cfg.OverlayTheme))
	win32.SetWindowText(d.rotation, strconv.Itoa(cfg.DisplayRotation))
	modeSel := 0
	if cfg.KneeMode == "confirm" {
		modeSel = 1
	}
	fillCombo(d.kneeMode, []string{"Arm then select", "Right then confirm"}, modeSel)
	fillCombo(d.leftRaises, []string{"1", "2", "3"}, cfg.KneeLeftRaises-1)
	dirSel := 0
	if cfg.KneeRightDirection == -1 {
		dirSel = 1
	}
	fillCombo(d.rightDirection, []string{"Down / advance", "Up / back"}, dirSel)
	roles := []string{"Off", "Left", "Right"}
	roleIDs := []string{"off", "left", "right"}
	for i, ch := range cfg.KneeChannels {
		fillCombo(d.kneeRole[i], roles, indexOf(roleIDs, ch.Role))
		win32.SetWindowText(d.kneeThreshold[i], strconv.Itoa(ch.ThresholdMM))
	}
	win32.SetCheck(d.deskEnabled, cfg.DeskEnabled)
	fillCombo(d.deskOrientation, []string{"0 degrees", "90 degrees", "180 degrees", "270 degrees"}, cfg.DeskOrientation/90)
	win32.SetWindowText(d.deskSensitivity, strconv.Itoa(cfg.DeskSensitivityMg))
	actions := []string{"None", "Tile", "Stack"}
	actionIDs := []string{"none", "tile", "stack"}
	for i, v := range []string{cfg.DeskLeft, cfg.DeskRight, cfg.DeskForward, cfg.DeskBack} {
		fillCombo(d.deskAction[i], actions, indexOf(actionIDs, v))
	}
	d.updateStatus()
	d.showPage(win32.TabGet(d.tab))
	win32.ShowWindow(d.hwnd, win32.SW_SHOW)
	win32.SetForegroundWindow(d.hwnd)
}

func (d *Dialog) SetSensorStatus(status [5]bool) { d.sensorOK = status; d.updateStatus() }
func (d *Dialog) updateStatus() {
	for i, h := range d.kneeStatus {
		if h == 0 {
			continue
		}
		s := "Not detected"
		if d.sensorOK[i] {
			s = "Detected"
		}
		win32.SetWindowText(h, s)
	}
	if d.deskStatus != 0 {
		s := "CH 4 // Not detected"
		if d.sensorOK[4] {
			s = "CH 4 // Detected"
		}
		win32.SetWindowText(d.deskStatus, s)
	}
}
func (d *Dialog) hide() {
	win32.ShowWindow(d.hwnd, win32.SW_HIDE)
	if d.OnClose != nil {
		d.OnClose()
	}
}
func parseInt(h windows.Handle, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(win32.GetWindowText(h)))
	if err != nil {
		return fallback
	}
	return v
}
func comboID(h windows.Handle, ids []string) string {
	i := win32.ComboGet(h)
	if i < 0 || i >= len(ids) {
		return ids[0]
	}
	return ids[i]
}

func (d *Dialog) save() {
	cfg := d.cfg
	cfg.Brands = config.DefaultBrands()
	for i, n := range config.BrandNames() {
		cfg.Brands[n] = win32.GetCheck(d.checks[i])
	}
	idx := win32.ComboGet(d.port)
	if idx <= 0 {
		cfg.PortMode = "auto"
		cfg.Port = ""
	} else if idx-1 < len(d.ports) {
		cfg.PortMode = "manual"
		cfg.Port = d.ports[idx-1].Name
	}
	delays := []int{250, 500, 750, 1000, 1500, 2000}
	di := win32.ComboGet(d.dwell)
	if di >= 0 && di < len(delays) {
		cfg.DwellMs = delays[di]
	}
	if win32.ComboGet(d.view) == 1 {
		cfg.OverlayView = "graphical"
	} else {
		cfg.OverlayView = "classic"
	}
	themes := overlay.Catalog()
	ti := win32.ComboGet(d.theme)
	if ti >= 0 && ti < len(themes) {
		cfg.OverlayTheme = themes[ti].ID
	}
	cfg.DisplayRotation = NormalizeDeg(parseInt(d.rotation, cfg.DisplayRotation))
	if win32.ComboGet(d.kneeMode) == 1 {
		cfg.KneeMode = "confirm"
	} else {
		cfg.KneeMode = "arm"
	}
	cfg.KneeLeftRaises = win32.ComboGet(d.leftRaises) + 1
	if win32.ComboGet(d.rightDirection) == 1 {
		cfg.KneeRightDirection = -1
	} else {
		cfg.KneeRightDirection = 1
	}
	roles := []string{"off", "left", "right"}
	cfg.KneeChannels = make([]config.KneeChannel, 4)
	for i := 0; i < 4; i++ {
		cfg.KneeChannels[i] = config.KneeChannel{Role: comboID(d.kneeRole[i], roles), ThresholdMM: parseInt(d.kneeThreshold[i], 75)}
	}
	cfg.DeskEnabled = win32.GetCheck(d.deskEnabled)
	cfg.DeskOrientation = win32.ComboGet(d.deskOrientation) * 90
	cfg.DeskSensitivityMg = parseInt(d.deskSensitivity, 350)
	actions := []string{"none", "tile", "stack"}
	cfg.DeskLeft = comboID(d.deskAction[0], actions)
	cfg.DeskRight = comboID(d.deskAction[1], actions)
	cfg.DeskForward = comboID(d.deskAction[2], actions)
	cfg.DeskBack = comboID(d.deskAction[3], actions)
	cfg.Normalize()
	if d.OnSave != nil {
		d.OnSave(cfg)
	}
	d.hide()
}
func (d *Dialog) Hwnd() windows.Handle { return d.hwnd }
