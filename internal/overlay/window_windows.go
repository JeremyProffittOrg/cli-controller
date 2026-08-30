package overlay

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

type Overlay struct {
	hwnd     windows.Handle
	items    []Item
	sel      int
	view     string
	theme    string
	shown    bool
	bg       windows.Handle
	selBr    windows.Handle
	knobBr   windows.Handle
	ringBr   windows.Handle
	font     windows.Handle
	fontB    windows.Handle
	fontS    windows.Handle
	fontItem windows.Handle
	fontSel  windows.Handle
	fontTech windows.Handle
}

func New() (*Overlay, error) {
	o := &Overlay{
		view:     ViewClassic,
		theme:    NormalizeTheme(""),
		bg:       win32.CreateBrush(win32.RGB(15, 23, 42)),
		selBr:    win32.CreateBrush(win32.RGB(30, 58, 95)),
		knobBr:   win32.CreateBrush(win32.RGB(30, 41, 59)),
		ringBr:   win32.CreateBrush(win32.RGB(8, 47, 73)),
		font:     win32.CreateFont(-16, win32.FW_NORMAL, "Segoe UI"),
		fontB:    win32.CreateFont(-16, win32.FW_BOLD, "Segoe UI"),
		fontS:    win32.CreateFont(-13, win32.FW_NORMAL, "Segoe UI"),
		fontItem: win32.CreateFont(-16, win32.FW_NORMAL, "Bahnschrift"),
		fontSel:  win32.CreateFont(-18, win32.FW_BOLD, "Bahnschrift"),
		fontTech: win32.CreateFont(-13, win32.FW_BOLD, "Bahnschrift"),
	}
	if err := win32.RegisterClass("CLIDialOverlay", overlayCB, o.bg); err != nil {
		return nil, err
	}
	h, err := win32.CreateWindow(
		win32.WS_EX_LAYERED|win32.WS_EX_TOPMOST|win32.WS_EX_TOOLWINDOW|win32.WS_EX_NOACTIVATE,
		win32.WS_POPUP,
		"CLIDialOverlay", "",
		0, 0, Width, 200, 0, 0,
	)
	if err != nil {
		return nil, err
	}
	o.hwnd = h
	win32.SetLayered(h, 248)
	overlayInst = o
	return o, nil
}

var overlayInst *Overlay
var overlayCB = windows.NewCallback(overlayProc)

func overlayProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	h := windows.Handle(hwnd)
	if overlayInst == nil {
		return win32.DefWindowProc(h, msg, wParam, lParam)
	}
	switch msg {
	case win32.WM_PAINT:
		overlayInst.paint(h)
		return 0
	case win32.WM_DESTROY:
		return 0
	}
	return win32.DefWindowProc(h, msg, wParam, lParam)
}

func (o *Overlay) SetView(view string) {
	if view != ViewGraphical {
		view = ViewClassic
	}
	o.view = view
}

func (o *Overlay) SetTheme(id string) {
	o.theme = NormalizeTheme(id)
}

func (o *Overlay) paint(h windows.Handle) {
	var ps win32.PAINTSTRUCT
	hdc := win32.BeginPaint(h, &ps)
	defer win32.EndPaint(h, &ps)
	rc := win32.GetClientRect(h)
	win32.FillRect(hdc, &rc, o.bg)
	if o.view == ViewGraphical {
		o.paintGraphical(hdc, rc)
		return
	}
	o.paintClassic(hdc, rc)
}

func (o *Overlay) paintClassic(hdc windows.Handle, rc win32.RECT) {
	win32.SetBkMode(hdc, win32.TRANSPARENT)
	title := win32.RECT{Left: 16, Top: 10, Right: rc.Right - 16, Bottom: 34}
	win32.SelectObject(hdc, o.fontB)
	win32.SetTextColor(hdc, win32.RGB(226, 232, 240))
	win32.DrawText(hdc, "CLI Dial", &title, win32.DT_LEFT|win32.DT_VCENTER|win32.DT_SINGLELINE|win32.DT_NOPREFIX)
	if len(o.items) == 0 {
		body := win32.RECT{Left: 16, Top: 48, Right: rc.Right - 16, Bottom: rc.Bottom - 12}
		win32.SelectObject(hdc, o.font)
		win32.SetTextColor(hdc, win32.RGB(148, 163, 184))
		win32.DrawText(hdc, "No matching CLI windows", &body, win32.DT_LEFT|win32.DT_NOPREFIX)
		return
	}
	vis := VisibleCount(image.Rect(0, 0, int(rc.Right), int(rc.Bottom)))
	start := 0
	if o.sel >= vis {
		start = o.sel - vis + 1
	}
	y := int32(40)
	for i := start; i < len(o.items) && i < start+vis; i++ {
		row := win32.RECT{Left: 8, Top: y, Right: rc.Right - 8, Bottom: y + RowH}
		if i == o.sel {
			win32.FillRect(hdc, &row, o.selBr)
			win32.SelectObject(hdc, o.fontB)
		} else {
			win32.SelectObject(hdc, o.font)
		}
		win32.SetTextColor(hdc, win32.RGB(226, 232, 240))
		text := FormatRow(i == o.sel, o.items[i].Brand, o.items[i].Title)
		pad := row
		pad.Left += 8
		win32.DrawText(hdc, text, &pad, win32.DT_LEFT|win32.DT_VCENTER|win32.DT_SINGLELINE|win32.DT_NOPREFIX|win32.DT_END_ELLIPSIS)
		y += RowH
	}
}

func (o *Overlay) paintGraphical(hdc windows.Handle, rc win32.RECT) {
	win32.SetBkMode(hdc, win32.TRANSPARENT)
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top
	if !DrawTheme(hdc, o.theme, rc) {
		win32.FillRect(hdc, &rc, o.bg)
	}
	cx := float64(w) * 0.32
	cy := float64(h) * 0.52
	outer := float64(h) * 0.22
	n := len(o.items)

	cyan := win32.CreatePen(3, win32.RGB(34, 211, 238))
	redGlow := win32.CreatePen(11, win32.RGB(80, 7, 22))
	red := win32.CreatePen(4, win32.RGB(255, 25, 62))
	redFine := win32.CreatePen(2, win32.RGB(255, 78, 104))
	panel := win32.CreateBrush(win32.RGB(6, 10, 18))
	panelEdge := win32.CreateBrush(win32.RGB(34, 211, 238))
	selectedPanel := win32.CreateBrush(win32.RGB(45, 8, 20))
	selectedEdge := win32.CreateBrush(win32.RGB(255, 25, 62))
	hub := win32.CreateBrush(win32.RGB(255, 25, 62))
	oldPen := win32.SelectObject(hdc, cyan)
	oldBr := win32.SelectObject(hdc, o.bg)

	rot := KnobRotation(o.sel, n)
	if n == 0 {
		rot = 0
	}
	needleAngle := -math.Pi/2 + rot
	nx, ny := Polar(cx, cy, outer, needleAngle)
	hx, hy := Polar(cx, cy, 18, needleAngle+math.Pi)
	win32.SelectObject(hdc, redGlow)
	win32.MoveTo(hdc, int32(hx), int32(hy))
	win32.LineTo(hdc, int32(nx), int32(ny))
	win32.SelectObject(hdc, red)
	win32.MoveTo(hdc, int32(hx), int32(hy))
	win32.LineTo(hdc, int32(nx), int32(ny))
	win32.SelectObject(hdc, hub)
	win32.Ellipse(hdc, int32(cx-10), int32(cy-10), int32(cx+10), int32(cy+10))

	const slots = 5
	rightX := w * 52 / 100
	rightEdge := w - int32(float64(w)*0.12)
	top := int32(float64(h) * 0.20)
	bottom := int32(float64(h) * 0.80)
	gap := int32(8)
	slotH := (bottom - top - gap*(slots-1)) / slots
	start := VisibleStart(n, o.sel, slots)
	for slot := 0; slot < slots; slot++ {
		y := top + int32(slot)*(slotH+gap)
		box := win32.RECT{Left: rightX, Top: y, Right: rightEdge, Bottom: y + slotH}
		idx := start + slot
		selected := idx < n && idx == o.sel
		if selected {
			win32.FillRect(hdc, &box, selectedEdge)
			inner := box
			inner.Left += 3
			inner.Top += 3
			inner.Right -= 3
			inner.Bottom -= 3
			win32.FillRect(hdc, &inner, selectedPanel)
		} else {
			win32.FillRect(hdc, &box, panelEdge)
			inner := box
			inner.Left++
			inner.Top++
			inner.Right--
			inner.Bottom--
			win32.FillRect(hdc, &inner, panel)
		}

		number := win32.RECT{Left: box.Left + 10, Top: box.Top, Right: box.Left + 46, Bottom: box.Bottom}
		win32.SelectObject(hdc, o.fontTech)
		win32.SetTextColor(hdc, win32.RGB(34, 211, 238))
		label := "--"
		if idx < n {
			label = fmt.Sprintf("%02d", idx+1)
		}
		win32.DrawText(hdc, label, &number, win32.DT_LEFT|win32.DT_VCENTER|win32.DT_SINGLELINE|win32.DT_NOPREFIX)

		textBox := win32.RECT{Left: box.Left + 50, Top: box.Top, Right: box.Right - 12, Bottom: box.Bottom}
		itemText := "EMPTY SLOT"
		if idx < n {
			it := o.items[idx]
			itemText = strings.TrimSpace(FormatRow(false, it.Brand, it.Title))
		}
		if selected {
			win32.SelectObject(hdc, o.fontSel)
			win32.SetTextColor(hdc, win32.RGB(255, 255, 255))
		} else {
			win32.SelectObject(hdc, o.fontItem)
			win32.SetTextColor(hdc, win32.RGB(186, 210, 224))
		}
		win32.DrawText(hdc, itemText, &textBox, win32.DT_LEFT|win32.DT_VCENTER|win32.DT_SINGLELINE|win32.DT_NOPREFIX|win32.DT_END_ELLIPSIS)

		if selected {
			arrowX := box.Left - 28
			arrowY := (box.Top + box.Bottom) / 2
			drawChevron(hdc, arrowX, arrowY, redGlow)
			drawChevron(hdc, arrowX, arrowY, red)
		}
	}

	readout := win32.RECT{Left: int32(cx - 70), Top: int32(cy + outer + 8), Right: int32(cx + 70), Bottom: int32(cy + outer + 32)}
	win32.SelectObject(hdc, o.fontTech)
	win32.SetTextColor(hdc, win32.RGB(226, 232, 240))
	win32.DrawText(hdc, fmtIndex(o.sel, n), &readout, win32.DT_CENTER|win32.DT_VCENTER|win32.DT_SINGLELINE|win32.DT_NOPREFIX)

	win32.SelectObject(hdc, oldPen)
	win32.SelectObject(hdc, oldBr)
	for _, obj := range []windows.Handle{cyan, redGlow, red, redFine, panel, panelEdge, selectedPanel, selectedEdge, hub} {
		win32.DeleteObject(obj)
	}
}

func fmtIndex(sel, n int) string {
	if n <= 0 || sel < 0 || sel >= n {
		return "--/--"
	}
	return fmt.Sprintf("%02d/%02d", sel+1, n)
}

func drawChevron(hdc windows.Handle, x, y int32, pen windows.Handle) {
	win32.SelectObject(hdc, pen)
	for offset := int32(0); offset < 18; offset += 9 {
		win32.MoveTo(hdc, x+offset, y-15)
		win32.LineTo(hdc, x+offset+14, y)
		win32.LineTo(hdc, x+offset, y+15)
	}
}

func (o *Overlay) Show(work image.Rectangle, items []Item, sel int) {
	o.items = items
	o.sel = sel
	if o.sel < 0 {
		o.sel = 0
	}
	if len(o.items) > 0 && o.sel >= len(o.items) {
		o.sel = len(o.items) - 1
	}
	b := BoundsFor(work, o.view)
	var rgn windows.Handle
	if o.view == ViewGraphical {
		rgn = win32.CreateEllipticRgn(0, 0, int32(b.Dx()), int32(b.Dy()))
	} else {
		rgn = win32.CreateRoundRectRgn(0, 0, int32(b.Dx()), int32(b.Dy()), 24, 24)
	}
	win32.SetWindowPos(o.hwnd, win32.HWND_TOPMOST, int32(b.Min.X), int32(b.Min.Y), int32(b.Dx()), int32(b.Dy()), win32.SWP_NOACTIVATE|win32.SWP_SHOWWINDOW)
	win32.SetWindowRgn(o.hwnd, rgn, true)
	win32.ShowWindow(o.hwnd, win32.SW_SHOWNOACTIVATE)
	o.shown = true
	win32.Invalidate(o.hwnd)
}

func (o *Overlay) Hide() {
	o.shown = false
	win32.ShowWindow(o.hwnd, win32.SW_HIDE)
}

func (o *Overlay) Visible() bool { return o.shown }

func (o *Overlay) Hwnd() windows.Handle { return o.hwnd }

func (o *Overlay) Sel() int { return o.sel }

func (o *Overlay) Items() []Item { return o.items }
