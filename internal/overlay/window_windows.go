package overlay

import (
	"image"
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
	win32.FillRect(hdc, &rc, o.bg)
	dial := h
	if dial > w {
		dial = w
	}
	themeBox := win32.RECT{Left: 0, Top: 0, Right: dial, Bottom: dial}
	if !DrawTheme(hdc, o.theme, themeBox) {
		win32.FillRect(hdc, &themeBox, o.bg)
	}
	n := len(o.items)
	const slots = 5
	rightX := dial * 50 / 100
	slotW := dial * 39 / 100 * 2
	rightEdge := rightX + slotW
	if rightEdge > w-16 {
		rightEdge = w - 16
	}
	top := int32(float64(h) * 0.20)
	bottom := int32(float64(h) * 0.80)
	gap := int32(6)
	slotH := (bottom - top - gap*(slots-1)) / slots
	start := VisibleStart(n, o.sel, slots)
	for slot := 0; slot < slots; slot++ {
		y := top + int32(slot)*(slotH+gap)
		box := win32.RECT{Left: rightX, Top: y, Right: rightEdge, Bottom: y + slotH}
		idx := start + slot
		selected := idx < n && idx == o.sel
		DrawEmbedded(hdc, SlotFile(selected), box, true)
		iconSize := slotH * 3 / 4
		if iconSize < 28 {
			iconSize = 28
		}
		iconY := box.Top + (slotH-iconSize)/2
		icon := win32.RECT{Left: box.Left + 12, Top: iconY, Right: box.Left + 12 + iconSize, Bottom: iconY + iconSize}
		iconPath := "icons/empty.jpg"
		itemText := "EMPTY SLOT"
		if idx < n {
			it := o.items[idx]
			iconPath = IconFile(it.Brand, it.Title)
			itemText = strings.TrimSpace(FormatRow(false, it.Brand, it.Title))
		}
		DrawEmbedded(hdc, iconPath, icon, true)
		textBox := win32.RECT{Left: icon.Right + 8, Top: box.Top, Right: box.Right - 16, Bottom: box.Bottom}
		if selected {
			win32.SelectObject(hdc, o.fontSel)
			win32.SetTextColor(hdc, win32.RGB(255, 255, 255))
		} else {
			win32.SelectObject(hdc, o.fontItem)
			win32.SetTextColor(hdc, win32.RGB(210, 226, 236))
		}
		win32.DrawText(hdc, itemText, &textBox, win32.DT_LEFT|win32.DT_VCENTER|win32.DT_SINGLELINE|win32.DT_NOPREFIX|win32.DT_END_ELLIPSIS)
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
		hh := int32(b.Dy())
		rgn = win32.CreateRoundRectRgn(0, 0, int32(b.Dx()), hh, hh, hh)
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
