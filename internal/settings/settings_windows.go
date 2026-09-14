package settings

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"github.com/JeremyProffittOrg/cli-controller/internal/overlay"
	"github.com/JeremyProffittOrg/cli-controller/internal/protocol"
	"github.com/JeremyProffittOrg/cli-controller/internal/serial"
	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

const (
	idSave     = 1
	idCancel   = 2
	idTab      = 100
	idPort     = 101
	idDwell    = 102
	idView     = 103
	idTheme    = 104
	idScan     = 160
	idFound    = 161
	idControls = 162
	idAdd      = 163
	idRemove   = 164
)

type Dialog struct {
	hwnd, tab                                                 windows.Handle
	pages                                                     [5][]windows.Handle
	checks                                                    []windows.Handle
	port, dwell, view, theme, rotation                        windows.Handle
	kneeMode, leftRaises, rightDirection                      windows.Handle
	kneeRole, kneeThreshold, kneeStatus                       [4]windows.Handle
	deskEnabled, deskStatus, deskOrientation, deskSensitivity windows.Handle
	deskAction                                                [4]windows.Handle
	foundList, controlList, foundStatus                       windows.Handle
	ctrlRole, ctrlThreshold, ctrlEnabled                      windows.Handle
	ctrlOrientation, ctrlSensitivity                          windows.Handle
	ctrlAction                                                [4]windows.Handle
	ctrlKindLabel, ctrlLive, deskLive                         windows.Handle
	ctrlRoleLabel, ctrlThreshLabel, ctrlOrientLabel           windows.Handle
	ctrlSensLabel                                             windows.Handle
	ctrlDirLabel                                              [4]windows.Handle
	ports                                                     []serial.PortInfo
	cfg                                                       config.Config
	sensorOK                                                  [5]bool
	inventory                                                 []protocol.SensorStatus
	controls                                                  []config.SensorControl
	selectedControl                                           int
	font, fontB, fontTech, bg, panel                          windows.Handle
	OnSave                                                    func(config.Config)
	OnClose                                                   func()
	OnScan                                                    func()
	i2cPort                                                   string
	i2cSda, i2cScl                                            int
}

var inst *Dialog
var cb = windows.NewCallback(proc)

func New(parent windows.Handle) (*Dialog, error) {
	win32.InitTabs()
	d := &Dialog{checks: make([]windows.Handle, len(config.BrandNames())), font: win32.CreateFont(-15, win32.FW_NORMAL, "Segoe UI"), fontB: win32.CreateFont(-18, win32.FW_BOLD, "Bahnschrift"), fontTech: win32.CreateFont(-13, win32.FW_BOLD, "Bahnschrift"), bg: win32.CreateBrush(win32.RGB(7, 12, 22)), panel: win32.CreateBrush(win32.RGB(12, 20, 34))}
	if err := win32.RegisterClass("CLIDialSettings", cb, d.bg); err != nil {
		return nil, err
	}
	h, err := win32.CreateWindow(win32.WS_EX_APPWINDOW, win32.WS_CAPTION|win32.WS_SYSMENU, "CLIDialSettings", "CLI Dial // Config", 180, 40, 700, 820, parent, 0)
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
		id := int(win32.LOWORD(wParam))
		note := win32.HIWORD(wParam)
		switch id {
		case idSave:
			inst.save()
			return 0
		case idCancel:
			inst.hide()
			return 0
		case idScan:
			if inst.OnScan != nil {
				inst.OnScan()
			}
			return 0
		case idAdd:
			inst.addControl()
			return 0
		case idRemove:
			inst.removeControl()
			return 0
		case idFound:
			if note == win32.LBN_SELCHANGE {
				inst.updateLiveReadouts()
				return 0
			}
		case idControls:
			if note == win32.LBN_SELCHANGE {
				inst.onControlSelect()
				return 0
			}
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
	d.tab = child(-1, 0, win32.WS_TABSTOP, "SysTabControl32", "", 24, 66, 652, 34, idTab)
	for i, name := range []string{"Controller", "Display", "Knees", "Desk", "Sensors"} {
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
	label(2, "Live reading", 455, 254, 160)
	for i := 0; i < 4; i++ {
		y := int32(282 + i*56)
		label(2, fmt.Sprintf("CH %d", i), 48, y+4, 65)
		d.kneeRole[i] = child(2, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 125, y, 145, 100, uintptr(120+i))
		d.kneeThreshold[i] = child(2, win32.WS_EX_CLIENTEDGE, win32.WS_TABSTOP, "EDIT", "75", 292, y, 125, 26, uintptr(130+i))
		d.kneeStatus[i] = label(2, "no signal", 455, y+4, 190)
	}
	section(3, "01  DESK MOTION SENSOR", 112)
	d.deskEnabled = child(3, 0, win32.BS_AUTOCHECKBOX|win32.WS_TABSTOP, "BUTTON", "Enable ADXL345 desk motion", 48, 140, 280, 24, 140)
	d.deskStatus = label(3, "no signal", 340, 142, 300)
	d.deskLive = label(3, "Live  waiting for sample", 48, 172, 600)
	label(3, "Board orientation", 48, 208, 150)
	d.deskOrientation = child(3, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 210, 204, 160, 120, 141)
	label(3, "Sensitivity (milli-g)", 48, 250, 150)
	d.deskSensitivity = child(3, win32.WS_EX_CLIENTEDGE, win32.WS_TABSTOP, "EDIT", "350", 210, 246, 160, 26, 142)
	section(3, "02  DIRECTION ACTIONS", 306)
	dirs := []string{"Left", "Right", "Forward", "Back"}
	for i, name := range dirs {
		y := int32(340 + i*52)
		label(3, name, 48, y+4, 120)
		d.deskAction[i] = child(3, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 180, y, 240, 100, uintptr(150+i))
	}
	section(4, "01  DISCOVERED HARDWARE", 112)
	d.foundList = child(4, win32.WS_EX_CLIENTEDGE, win32.LBS_NOTIFY|win32.LBS_HASSTRINGS|win32.LBS_NOINTEGRALHEIGHT|win32.WS_VSCROLL|win32.WS_TABSTOP, "LISTBOX", "", 48, 136, 600, 118, idFound)
	d.foundStatus = label(4, "Mux 0x70-0x77 and the root I2C bus. Live readings update while this window is open.", 48, 258, 490)
	child(4, 0, win32.BS_PUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Scan now", 548, 254, 100, 26, idScan)
	section(4, "02  CONTROLS", 290)
	d.controlList = child(4, win32.WS_EX_CLIENTEDGE, win32.LBS_NOTIFY|win32.LBS_HASSTRINGS|win32.LBS_NOINTEGRALHEIGHT|win32.WS_VSCROLL|win32.WS_TABSTOP, "LISTBOX", "", 48, 314, 600, 100, idControls)
	child(4, 0, win32.BS_PUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Add control", 48, 420, 120, 26, idAdd)
	child(4, 0, win32.BS_PUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Remove", 178, 420, 100, 26, idRemove)
	section(4, "03  SELECTED CONTROL", 456)
	d.ctrlKindLabel = label(4, "Select a discovered sensor, then Add control.", 48, 484, 600)
	d.ctrlLive = label(4, "Live  no signal", 48, 504, 600)
	d.ctrlRoleLabel = label(4, "Role", 48, 532, 80)
	d.ctrlRole = child(4, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 130, 528, 160, 100, 165)
	d.ctrlThreshLabel = label(4, "Threshold mm", 310, 532, 110)
	d.ctrlThreshold = child(4, win32.WS_EX_CLIENTEDGE, win32.WS_TABSTOP, "EDIT", "75", 430, 528, 90, 26, 166)
	d.ctrlEnabled = child(4, 0, win32.BS_AUTOCHECKBOX|win32.WS_TABSTOP, "BUTTON", "Enable desk motion", 48, 564, 220, 24, 167)
	d.ctrlOrientLabel = label(4, "Orientation", 280, 564, 90)
	d.ctrlOrientation = child(4, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", 370, 560, 150, 120, 168)
	d.ctrlSensLabel = label(4, "Sensitivity mg", 48, 600, 120)
	d.ctrlSensitivity = child(4, win32.WS_EX_CLIENTEDGE, win32.WS_TABSTOP, "EDIT", "350", 170, 596, 100, 26, 169)
	dirs2 := []string{"Left", "Right", "Fwd", "Back"}
	for i, name := range dirs2 {
		x := int32(48 + i*155)
		d.ctrlDirLabel[i] = label(4, name, x, 628, 70)
		d.ctrlAction[i] = child(4, 0, win32.CBS_DROPDOWNLIST|win32.CBS_HASSTRINGS|win32.WS_TABSTOP, "COMBOBOX", "", x, 648, 145, 100, uintptr(170+i))
	}
	save := child(-1, 0, win32.BS_DEFPUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Commit", 450, 740, 100, 32, idSave)
	cancel := child(-1, 0, win32.BS_PUSHBUTTON|win32.WS_TABSTOP, "BUTTON", "Abort", 566, 740, 100, 32, idCancel)
	win32.Send(save, win32.WM_SETFONT, uintptr(d.fontB), 1)
	win32.Send(cancel, win32.WM_SETFONT, uintptr(d.fontB), 1)
	d.showPage(0)
}

func (d *Dialog) showPage(selected int) {
	if selected < 0 || selected > 4 {
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
	if selected == 4 {
		d.loadControlFields()
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
	d.controls = append([]config.SensorControl(nil), cfg.Sensors...)
	d.selectedControl = 0
	if len(d.controls) == 0 {
		d.selectedControl = -1
	}
	fillCombo(d.ctrlRole, roles, 0)
	fillCombo(d.ctrlOrientation, []string{"0 degrees", "90 degrees", "180 degrees", "270 degrees"}, 0)
	for i := 0; i < 4; i++ {
		fillCombo(d.ctrlAction[i], actions, 0)
	}
	d.refreshSensorLists()
	d.loadControlFields()
	d.showPage(win32.TabGet(d.tab))
	win32.ShowWindow(d.hwnd, win32.SW_SHOW)
	win32.SetForegroundWindow(d.hwnd)
}

func (d *Dialog) SetSensorStatus(status [5]bool) { d.sensorOK = status; d.updateLiveReadouts() }

func (d *Dialog) SetInventory(list []protocol.SensorStatus) {
	d.inventory = mergeInventory(d.inventory, list)
	d.syncLegacyFromInventory()
	d.refreshSensorLists()
}

func (d *Dialog) SetI2CPort(port string, sda, scl int) {
	d.i2cPort = port
	d.i2cSda = sda
	d.i2cScl = scl
	d.updateLiveReadouts()
}

func (d *Dialog) SetLive(list []protocol.SensorStatus) {
	d.inventory = append([]protocol.SensorStatus(nil), list...)
	d.syncLegacyFromInventory()
	if !win32.IsWindowVisible(d.hwnd) {
		return
	}
	d.patchFoundList()
	d.updateLiveReadouts()
}

func mergeInventory(old, next []protocol.SensorStatus) []protocol.SensorStatus {
	prev := map[string]protocol.SensorStatus{}
	for _, s := range old {
		prev[s.ID] = s
	}
	out := make([]protocol.SensorStatus, len(next))
	copy(out, next)
	for i, s := range out {
		p, ok := prev[s.ID]
		if !ok || !p.Live {
			continue
		}
		out[i].Live = true
		out[i].MM = p.MM
		out[i].X = p.X
		out[i].Y = p.Y
		out[i].Z = p.Z
	}
	return out
}

func (d *Dialog) syncLegacyFromInventory() {
	var legacy [5]bool
	for _, s := range d.inventory {
		if !s.OK {
			continue
		}
		if s.Kind == "accel" && (s.ID == protocol.FormatSensorID(0x70, 4, "accel") || s.Ch == 4) {
			legacy[4] = true
		}
		if s.Kind == "tof" && s.Ch >= 0 && s.Ch < 4 && (s.Mux == 0x70 || strings.HasPrefix(s.ID, "mux:70:")) {
			legacy[s.Ch] = true
		}
	}
	d.sensorOK = legacy
}

func (d *Dialog) liveByID(id string) (protocol.SensorStatus, bool) {
	for _, s := range d.inventory {
		if s.ID == id {
			return s, true
		}
	}
	return protocol.SensorStatus{}, false
}

func (d *Dialog) liveTof(ch int) protocol.SensorStatus {
	id := protocol.FormatSensorID(0x70, ch, "tof")
	if s, ok := d.liveByID(id); ok {
		return s
	}
	for _, s := range d.inventory {
		if s.Kind == "tof" && s.Ch == ch && (s.Mux == 0x70 || s.Mux == 0) {
			return s
		}
	}
	return protocol.SensorStatus{Kind: "tof"}
}

func (d *Dialog) liveDesk() protocol.SensorStatus {
	if s, ok := d.liveByID(protocol.FormatSensorID(0x70, 4, "accel")); ok {
		return s
	}
	for _, s := range d.inventory {
		if s.Kind == "accel" {
			return s
		}
	}
	return protocol.SensorStatus{Kind: "accel"}
}

func (d *Dialog) patchFoundList() {
	if len(d.inventory) == 0 || win32.ListCount(d.foundList) != len(d.inventory) {
		return
	}
	for i, s := range d.inventory {
		win32.ListSetText(d.foundList, i, hardwareLabel(s))
	}
}

func (d *Dialog) updateLiveReadouts() {
	for i, h := range d.kneeStatus {
		if h == 0 {
			continue
		}
		win32.SetWindowText(h, d.liveTof(i).LiveText())
	}
	desk := d.liveDesk()
	if d.deskStatus != 0 {
		state := "no signal"
		if desk.OK && desk.Live {
			state = "live"
		} else if desk.OK {
			state = "waiting"
		}
		win32.SetWindowText(d.deskStatus, state)
	}
	if d.deskLive != 0 {
		win32.SetWindowText(d.deskLive, "Live  "+desk.LiveText())
	}
	if d.ctrlLive != 0 {
		text := "Live  no signal"
		if d.selectedControl >= 0 && d.selectedControl < len(d.controls) {
			if s, ok := d.liveByID(d.controls[d.selectedControl].ID); ok {
				text = "Live  " + s.LiveText()
			}
		}
		win32.SetWindowText(d.ctrlLive, text)
	}
	ok := 0
	live := 0
	for _, s := range d.inventory {
		if s.OK {
			ok++
		}
		if s.Live {
			live++
		}
	}
	if d.foundStatus != 0 {
		sel := ""
		idx := win32.ListGet(d.foundList)
		if idx >= 0 && idx < len(d.inventory) {
			sel = "  Selected " + d.inventory[idx].LiveText()
		}
		port := "Port A or B"
		if d.i2cPort == "a" {
			port = fmt.Sprintf("Port A (GPIO%d SDA, GPIO%d SCL)", d.i2cSda, d.i2cScl)
		} else if d.i2cPort == "b" {
			port = fmt.Sprintf("Port B (GPIO%d SDA, GPIO%d SCL)", d.i2cSda, d.i2cScl)
		}
		win32.SetWindowText(d.foundStatus, fmt.Sprintf("%s. %d device(s), %d connected, %d streaming.%s", port, len(d.inventory), ok, live, sel))
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
	d.storeControlFields()
	cfg.Sensors = append([]config.SensorControl(nil), d.controls...)
	for i := 0; i < 4; i++ {
		s := config.SensorControl{
			ID: protocol.FormatSensorID(0x70, i, "tof"), Kind: "tof",
			Role: cfg.KneeChannels[i].Role, ThresholdMM: cfg.KneeChannels[i].ThresholdMM,
		}
		cfg.UpsertSensor(s)
	}
	cfg.UpsertSensor(config.SensorControl{
		ID: protocol.FormatSensorID(0x70, 4, "accel"), Kind: "accel", Enabled: cfg.DeskEnabled,
		SensitivityMg: cfg.DeskSensitivityMg, Orientation: cfg.DeskOrientation,
		Left: cfg.DeskLeft, Right: cfg.DeskRight, Forward: cfg.DeskForward, Back: cfg.DeskBack,
	})
	cfg.Normalize()
	if d.OnSave != nil {
		d.OnSave(cfg)
	}
	d.hide()
}
func (d *Dialog) Hwnd() windows.Handle { return d.hwnd }

func hardwareLabel(s protocol.SensorStatus) string {
	kind := "VL53L4CD"
	if s.Kind == "accel" {
		kind = "ADXL345"
	}
	state := "Missing"
	if s.OK {
		state = "Detected"
	}
	live := s.LiveText()
	if strings.HasPrefix(s.ID, "root:") || s.Mux == 0 && strings.HasPrefix(s.ID, "root") {
		return fmt.Sprintf("root bus  %s  %s  %s  %s", kind, s.ID, state, live)
	}
	mux := s.Mux
	if mux == 0 {
		mux = 0x70
	}
	return fmt.Sprintf("mux 0x%02X ch %d  %s  %s  %s  %s", mux, s.Ch, kind, s.ID, state, live)
}

func controlLabel(s config.SensorControl) string {
	if s.Kind == "accel" {
		on := "off"
		if s.Enabled {
			on = "on"
		}
		return fmt.Sprintf("%s  ADXL345  desk %s", s.ID, on)
	}
	return fmt.Sprintf("%s  VL53L4CD  %s  %d mm", s.ID, s.Role, s.ThresholdMM)
}

func (d *Dialog) refreshSensorLists() {
	selFound := win32.ListGet(d.foundList)
	selCtrl := win32.ListGet(d.controlList)
	win32.ListReset(d.foundList)
	if len(d.inventory) == 0 {
		win32.ListAdd(d.foundList, "(none yet — press Scan now, or wait for the Dial)")
	} else {
		for _, s := range d.inventory {
			win32.ListAdd(d.foundList, hardwareLabel(s))
		}
	}
	if selFound >= 0 && selFound < win32.ListCount(d.foundList) {
		win32.ListSet(d.foundList, selFound)
	} else if win32.ListCount(d.foundList) > 0 {
		win32.ListSet(d.foundList, 0)
	}
	win32.ListReset(d.controlList)
	for _, s := range d.controls {
		win32.ListAdd(d.controlList, controlLabel(s))
	}
	if len(d.controls) == 0 {
		d.selectedControl = -1
		return
	}
	if selCtrl < 0 || selCtrl >= len(d.controls) {
		selCtrl = d.selectedControl
	}
	if selCtrl < 0 || selCtrl >= len(d.controls) {
		selCtrl = 0
	}
	d.selectedControl = selCtrl
	win32.ListSet(d.controlList, selCtrl)
	d.updateLiveReadouts()
}

func (d *Dialog) addControl() {
	idx := win32.ListGet(d.foundList)
	if idx < 0 || idx >= len(d.inventory) {
		return
	}
	st := d.inventory[idx]
	id := st.ID
	if id == "" {
		id = protocol.FormatSensorID(st.Mux, st.Ch, st.Kind)
	}
	for i, s := range d.controls {
		if s.ID == id {
			d.storeControlFields()
			d.selectedControl = i
			win32.ListSet(d.controlList, i)
			d.loadControlFields()
			return
		}
	}
	d.storeControlFields()
	d.controls = append(d.controls, config.NewControlFromStatus(st, d.cfg))
	d.selectedControl = len(d.controls) - 1
	d.refreshSensorLists()
	d.loadControlFields()
}

func (d *Dialog) removeControl() {
	idx := win32.ListGet(d.controlList)
	if idx < 0 || idx >= len(d.controls) {
		return
	}
	d.controls = append(d.controls[:idx], d.controls[idx+1:]...)
	if idx >= len(d.controls) {
		idx = len(d.controls) - 1
	}
	d.selectedControl = idx
	d.refreshSensorLists()
	d.loadControlFields()
}

func (d *Dialog) onControlSelect() {
	d.storeControlFields()
	d.selectedControl = win32.ListGet(d.controlList)
	d.loadControlFields()
}

func (d *Dialog) storeControlFields() {
	i := d.selectedControl
	if i < 0 || i >= len(d.controls) {
		return
	}
	s := d.controls[i]
	if s.Kind == "accel" {
		s.Enabled = win32.GetCheck(d.ctrlEnabled)
		s.Orientation = win32.ComboGet(d.ctrlOrientation) * 90
		s.SensitivityMg = parseInt(d.ctrlSensitivity, 350)
		actions := []string{"none", "tile", "stack"}
		s.Left = comboID(d.ctrlAction[0], actions)
		s.Right = comboID(d.ctrlAction[1], actions)
		s.Forward = comboID(d.ctrlAction[2], actions)
		s.Back = comboID(d.ctrlAction[3], actions)
	} else {
		s.Role = comboID(d.ctrlRole, []string{"off", "left", "right"})
		s.ThresholdMM = parseInt(d.ctrlThreshold, 75)
	}
	d.controls[i] = s
	for ch := 0; ch < 4; ch++ {
		if s.ID == protocol.FormatSensorID(0x70, ch, "tof") && s.Kind == "tof" {
			win32.ComboSet(d.kneeRole[ch], indexOf([]string{"off", "left", "right"}, s.Role))
			win32.SetWindowText(d.kneeThreshold[ch], strconv.Itoa(s.ThresholdMM))
		}
	}
	if s.ID == protocol.FormatSensorID(0x70, 4, "accel") && s.Kind == "accel" {
		win32.SetCheck(d.deskEnabled, s.Enabled)
		win32.ComboSet(d.deskOrientation, s.Orientation/90)
		win32.SetWindowText(d.deskSensitivity, strconv.Itoa(s.SensitivityMg))
		actions := []string{"none", "tile", "stack"}
		win32.ComboSet(d.deskAction[0], indexOf(actions, s.Left))
		win32.ComboSet(d.deskAction[1], indexOf(actions, s.Right))
		win32.ComboSet(d.deskAction[2], indexOf(actions, s.Forward))
		win32.ComboSet(d.deskAction[3], indexOf(actions, s.Back))
	}
}

func (d *Dialog) loadControlFields() {
	i := d.selectedControl
	show := func(h windows.Handle, on bool) {
		mode := win32.SW_HIDE
		if on {
			mode = win32.SW_SHOW
		}
		if h != 0 {
			win32.ShowWindow(h, mode)
		}
	}
	tofOn := false
	accelOn := false
	if i < 0 || i >= len(d.controls) {
		win32.SetWindowText(d.ctrlKindLabel, "Select a discovered sensor, then Add control.")
		if d.ctrlLive != 0 {
			win32.SetWindowText(d.ctrlLive, "Live  no signal")
		}
	} else {
		s := d.controls[i]
		if s.Kind == "accel" {
			accelOn = true
			win32.SetWindowText(d.ctrlKindLabel, fmt.Sprintf("%s  ADXL345 desk-motion control", s.ID))
			win32.SetCheck(d.ctrlEnabled, s.Enabled)
			win32.ComboSet(d.ctrlOrientation, s.Orientation/90)
			win32.SetWindowText(d.ctrlSensitivity, strconv.Itoa(s.SensitivityMg))
			actions := []string{"none", "tile", "stack"}
			win32.ComboSet(d.ctrlAction[0], indexOf(actions, s.Left))
			win32.ComboSet(d.ctrlAction[1], indexOf(actions, s.Right))
			win32.ComboSet(d.ctrlAction[2], indexOf(actions, s.Forward))
			win32.ComboSet(d.ctrlAction[3], indexOf(actions, s.Back))
		} else {
			tofOn = true
			win32.SetWindowText(d.ctrlKindLabel, fmt.Sprintf("%s  VL53L4CD distance control", s.ID))
			win32.ComboSet(d.ctrlRole, indexOf([]string{"off", "left", "right"}, s.Role))
			win32.SetWindowText(d.ctrlThreshold, strconv.Itoa(s.ThresholdMM))
		}
	}
	show(d.ctrlRoleLabel, tofOn)
	show(d.ctrlRole, tofOn)
	show(d.ctrlThreshLabel, tofOn)
	show(d.ctrlThreshold, tofOn)
	show(d.ctrlEnabled, accelOn)
	show(d.ctrlOrientLabel, accelOn)
	show(d.ctrlOrientation, accelOn)
	show(d.ctrlSensLabel, accelOn)
	show(d.ctrlSensitivity, accelOn)
	for i := range d.ctrlAction {
		show(d.ctrlDirLabel[i], accelOn)
		show(d.ctrlAction[i], accelOn)
	}
	d.updateLiveReadouts()
}
