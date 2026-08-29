package overlay

import (
	"image"

	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

type Overlay struct {
	hwnd   windows.Handle
	items  []Item
	sel    int
	shown  bool
	bg     windows.Handle
	selBr  windows.Handle
	font   windows.Handle
	fontB  windows.Handle
}

func New() (*Overlay, error) {
	o := &Overlay{
		bg:    win32.CreateBrush(win32.RGB(15, 23, 42)),
		selBr: win32.CreateBrush(win32.RGB(30, 58, 95)),
		font:  win32.CreateFont(-16, win32.FW_NORMAL, "Segoe UI"),
		fontB: win32.CreateFont(-16, win32.FW_BOLD, "Segoe UI"),
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

func (o *Overlay) paint(h windows.Handle) {
	var ps win32.PAINTSTRUCT
	hdc := win32.BeginPaint(h, &ps)
	defer win32.EndPaint(h, &ps)
	rc := win32.GetClientRect(h)
	win32.FillRect(hdc, &rc, o.bg)
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

func (o *Overlay) Show(work image.Rectangle, items []Item, sel int) {
	o.items = items
	o.sel = sel
	if o.sel < 0 {
		o.sel = 0
	}
	if len(o.items) > 0 && o.sel >= len(o.items) {
		o.sel = len(o.items) - 1
	}
	b := Bounds(work)
	rgn := win32.CreateRoundRectRgn(0, 0, int32(b.Dx()), int32(b.Dy()), 24, 24)
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

func (o *Overlay) Sel() int { return o.sel }

func (o *Overlay) Items() []Item { return o.items }
