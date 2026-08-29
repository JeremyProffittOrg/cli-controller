package tray

import (
	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

const (
	IDSettings = 1
	IDExit     = 2
)

type Tray struct {
	hwnd         windows.Handle
	on           []windows.Handle
	off          []windows.Handle
	cur          windows.Handle
	frame        int
	connected    bool
	port         string
}

func New(hwnd windows.Handle) *Tray {
	on, off := makeIcons()
	t := &Tray{hwnd: hwnd, on: on, off: off, cur: off[0]}
	t.add()
	return t
}

func (t *Tray) add() {
	var d win32.NOTIFYICONDATA
	d.Wnd = t.hwnd
	d.ID = 1
	d.Flags = win32.NIF_MESSAGE | win32.NIF_ICON | win32.NIF_TIP
	d.CallbackMessage = win32.WM_TRAY
	d.Icon = t.cur
	win32.UTF16Copy(d.Tip[:], t.tip())
	win32.ShellNotifyIcon(win32.NIM_ADD, &d)
}

func (t *Tray) tip() string {
	if t.connected && t.port != "" {
		return "CLI Dial: connected (" + t.port + ")"
	}
	return "CLI Dial: not connected"
}

func (t *Tray) SetConnected(on bool, port string) {
	t.connected = on
	t.port = port
	if !on {
		t.frame = 0
		if len(t.off) > 0 {
			t.cur = t.off[0]
		}
	}
	t.modify(false)
}

func (t *Tray) Tick() {
	if !t.connected || len(t.on) == 0 {
		return
	}
	t.frame = (t.frame + 1) % len(t.on)
	t.cur = t.on[t.frame]
	t.modify(false)
}

func (t *Tray) modify(balloon bool) {
	var d win32.NOTIFYICONDATA
	d.Wnd = t.hwnd
	d.ID = 1
	d.Flags = win32.NIF_MESSAGE | win32.NIF_ICON | win32.NIF_TIP
	d.CallbackMessage = win32.WM_TRAY
	d.Icon = t.cur
	win32.UTF16Copy(d.Tip[:], t.tip())
	if balloon {
		d.Flags |= win32.NIF_INFO
		d.InfoFlags = win32.NIIF_INFO | win32.NIIF_NOSOUND
		d.TimeoutOrVer = 2000
		win32.UTF16Copy(d.InfoTitle[:], "CLI Dial")
		win32.UTF16Copy(d.Info[:], t.tip())
	}
	win32.ShellNotifyIcon(win32.NIM_MODIFY, &d)
}

func (t *Tray) Balloon() { t.modify(true) }

func (t *Tray) Popup() uintptr {
	p := win32.GetCursorPos()
	m := win32.CreatePopupMenu()
	win32.AppendMenu(m, win32.MF_STRING, IDSettings, "Settings")
	win32.AppendMenu(m, win32.MF_STRING, IDExit, "Exit")
	win32.SetForegroundWindow(t.hwnd)
	cmd := win32.TrackPopupMenu(m, win32.TPM_RIGHTBUTTON|win32.TPM_BOTTOMALIGN|win32.TPM_RIGHTALIGN|win32.TPM_RETURNCMD, p.X, p.Y, t.hwnd)
	win32.DestroyMenu(m)
	win32.Post(t.hwnd, 0, 0, 0)
	return cmd
}

func (t *Tray) Close() {
	var d win32.NOTIFYICONDATA
	d.Wnd = t.hwnd
	d.ID = 1
	win32.ShellNotifyIcon(win32.NIM_DELETE, &d)
	for _, h := range t.on {
		if h != 0 {
			win32.DestroyIcon(h)
		}
	}
	for _, h := range t.off {
		if h != 0 {
			win32.DestroyIcon(h)
		}
	}
}
