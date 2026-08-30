package overlay

import (
	"image"
	"math"

	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

type Overlay struct {
	hwnd   windows.Handle
	items  []Item
	sel    int
	view   string
	shown  bool
	bg     windows.Handle
	selBr  windows.Handle
	knobBr windows.Handle
	ringBr windows.Handle
	font   windows.Handle
	fontB  windows.Handle
	fontS  windows.Handle
}

func New() (*Overlay, error) {
	o := &Overlay{
		view:   ViewClassic,
		bg:     win32.CreateBrush(win32.RGB(15, 23, 42)),
		selBr:  win32.CreateBrush(win32.RGB(30, 58, 95)),
		knobBr: win32.CreateBrush(win32.RGB(30, 41, 59)),
		ringBr: win32.CreateBrush(win32.RGB(8, 47, 73)),
		font:   win32.CreateFont(-16, win32.FW_NORMAL, "Segoe UI"),
		fontB:  win32.CreateFont(-16, win32.FW_BOLD, "Segoe UI"),
		fontS:  win32.CreateFont(-13, win32.FW_NORMAL, "Segoe UI"),
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
	win32.SetLayered(h, 235)
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
	cx := float64(rc.Left+rc.Right) / 2
	cy := float64(rc.Top+rc.Bottom) / 2
	side := float64(rc.Right - rc.Left)
	if h := float64(rc.Bottom - rc.Top); h < side {
		side = h
	}
	outer := side/2 - 8
	labelR := outer - 70
	knobR := outer * 0.38
	n := len(o.items)

	ringPen := win32.CreatePen(3, win32.RGB(56, 189, 248))
	tickPen := win32.CreatePen(2, win32.RGB(148, 163, 184))
	needlePen := win32.CreatePen(5, win32.RGB(250, 204, 21))
	oldPen := win32.SelectObject(hdc, ringPen)
	oldBr := win32.SelectObject(hdc, o.ringBr)
	win32.Ellipse(hdc, int32(cx-outer), int32(cy-outer), int32(cx+outer), int32(cy+outer))
	win32.SelectObject(hdc, o.knobBr)
	win32.SelectObject(hdc, tickPen)
	win32.Ellipse(hdc, int32(cx-knobR), int32(cy-knobR), int32(cx+knobR), int32(cy+knobR))

	rot := KnobRotation(o.sel, n)
	if n == 0 {
		rot = 0
	}
	win32.SelectObject(hdc, tickPen)
	for i := 0; i < 12; i++ {
		ang := -math.Pi/2 + rot + float64(i)*math.Pi/6
		x0, y0 := Polar(cx, cy, knobR-6, ang)
		x1, y1 := Polar(cx, cy, knobR-18, ang)
		win32.MoveTo(hdc, int32(x0), int32(y0))
		win32.LineTo(hdc, int32(x1), int32(y1))
	}
	nx, ny := Polar(cx, cy, knobR-10, -math.Pi/2+rot)
	hx, hy := Polar(cx, cy, 10, -math.Pi/2+rot+math.Pi)
	win32.SelectObject(hdc, needlePen)
	win32.MoveTo(hdc, int32(hx), int32(hy))
	win32.LineTo(hdc, int32(nx), int32(ny))
	hub := win32.CreateBrush(win32.RGB(250, 204, 21))
	win32.SelectObject(hdc, hub)
	win32.SelectObject(hdc, needlePen)
	win32.Ellipse(hdc, int32(cx-8), int32(cy-8), int32(cx+8), int32(cy+8))

	if n == 0 {
		body := win32.RECT{Left: int32(cx - 140), Top: int32(cy + knobR + 8), Right: int32(cx + 140), Bottom: int32(cy + knobR + 40)}
		win32.SelectObject(hdc, o.font)
		win32.SetTextColor(hdc, win32.RGB(148, 163, 184))
		win32.DrawText(hdc, "No matching CLI windows", &body, win32.DT_LEFT|win32.DT_NOPREFIX)
	} else {
		for i, it := range o.items {
			ang := RingAngle(i, o.sel, n)
			lx, ly := Polar(cx, cy, labelR, ang)
			label := Truncate(it.Title, 22)
			w := int32(168)
			h := int32(34)
			box := win32.RECT{Left: int32(lx) - w/2, Top: int32(ly) - h/2, Right: int32(lx) + w/2, Bottom: int32(ly) + h/2}
			if i == o.sel {
				win32.FillRect(hdc, &box, o.selBr)
				win32.SelectObject(hdc, o.fontB)
				win32.SetTextColor(hdc, win32.RGB(250, 204, 21))
			} else {
				win32.SelectObject(hdc, o.fontS)
				win32.SetTextColor(hdc, win32.RGB(226, 232, 240))
			}
			win32.DrawText(hdc, label, &box, win32.DT_LEFT|win32.DT_VCENTER|win32.DT_SINGLELINE|win32.DT_NOPREFIX|win32.DT_END_ELLIPSIS|win32.DT_CENTER)
		}
	}

	win32.SelectObject(hdc, oldPen)
	win32.SelectObject(hdc, oldBr)
	win32.DeleteObject(ringPen)
	win32.DeleteObject(tickPen)
	win32.DeleteObject(needlePen)
	win32.DeleteObject(hub)
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
