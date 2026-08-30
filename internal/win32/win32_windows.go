package win32

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	WS_POPUP          = 0x80000000
	WS_VISIBLE        = 0x10000000
	WS_CAPTION        = 0x00C00000
	WS_SYSMENU        = 0x00080000
	WS_CHILD          = 0x40000000
	WS_TABSTOP        = 0x00010000
	WS_OVERLAPPEDWINDOW = 0x00CF0000

	WS_EX_LAYERED     = 0x00080000
	WS_EX_TOPMOST     = 0x00000008
	WS_EX_TOOLWINDOW  = 0x00000080
	WS_EX_NOACTIVATE  = 0x08000000
	WS_EX_APPWINDOW   = 0x00040000
	WS_EX_CLIENTEDGE  = 0x00000200

	WM_DESTROY       = 0x0002
	WM_CLOSE         = 0x0010
	WM_PAINT         = 0x000F
	WM_COMMAND       = 0x0111
	WM_TIMER         = 0x0113
	WM_RBUTTONUP     = 0x0205
	WM_LBUTTONDOWN   = 0x0201
	WM_LBUTTONUP     = 0x0202
	WM_MOUSEMOVE     = 0x0200
	MK_LBUTTON       = 0x0001
	WM_CONTEXTMENU   = 0x007B
	WM_APP           = 0x8000
	WM_USER          = 0x0400
	WM_QUIT          = 0x0012
	WM_SETFONT       = 0x0030
	WM_ERASEBKGND       = 0x0014
	WM_CTLCOLOREDIT     = 0x0133
	WM_CTLCOLORLISTBOX  = 0x0134
	WM_CTLCOLORBTN      = 0x0135
	WM_CTLCOLORSTATIC = 0x0138
	CBN_SELCHANGE       = 1
	HALFTONE            = 4

	SW_HIDE            = 0
	SW_SHOW            = 5
	SW_SHOWNOACTIVATE  = 4
	SW_RESTORE         = 9
	SW_SHOWMINNOACTIVE = 7

	SWP_NOSIZE         = 0x0001
	SWP_NOMOVE         = 0x0002
	SWP_NOZORDER       = 0x0004
	SWP_NOACTIVATE     = 0x0010
	SWP_SHOWWINDOW     = 0x0040
	SWP_HIDEWINDOW     = 0x0080

	HWND_TOPMOST = ^uintptr(0) // -1
	HWND_TOP     = 0
	HWND_MESSAGE = ^uintptr(2) // -3

	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001
	CS_DBLCLKS = 0x0008

	COLOR_WINDOW = 5
	WHITE_BRUSH  = 0
	NULL_BRUSH   = 5

	DT_LEFT     = 0x0000
	DT_CENTER   = 0x0001
	DT_RIGHT    = 0x0002
	DT_VCENTER  = 0x0004
	DT_SINGLELINE = 0x0020
	DT_NOPREFIX = 0x0800
	DT_END_ELLIPSIS = 0x00008000

	TRANSPARENT = 1

	LWA_ALPHA = 0x00000002

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
	NIF_INFO    = 0x00000010
	NIF_STATE   = 0x00000008

	NIM_ADD    = 0
	NIM_MODIFY = 1
	NIM_DELETE = 2

	NIIF_INFO  = 0x00000001
	NIIF_NOSOUND = 0x00000010

	TPM_RIGHTBUTTON = 0x0002
	TPM_BOTTOMALIGN = 0x0020
	TPM_RIGHTALIGN  = 0x0008
	TPM_RETURNCMD   = 0x0100
	TPM_NONOTIFY    = 0x0080

	MF_STRING = 0x0000
	MF_SEPARATOR = 0x0800

	BS_AUTOCHECKBOX = 0x00000003
	BS_DEFPUSHBUTTON = 0x00000001
	BS_PUSHBUTTON   = 0x00000000
	CBS_DROPDOWNLIST = 0x0003
	CBS_HASSTRINGS  = 0x0200

	CW_USEDEFAULT = ^int32(0x7fffffff) + 1 // -2147483648 as int32 via bit

	MONITOR_DEFAULTTOPRIMARY  = 1
	MONITOR_DEFAULTTONEAREST = 2
	MONITORINFOF_PRIMARY     = 1

	SM_CXSCREEN  = 0
	SM_CYSCREEN  = 1
	SM_CYCAPTION = 4
	SM_CYFRAME   = 33

	GWL_EXSTYLE = -20

	IDOK     = 1
	IDCANCEL = 2

	CB_ADDSTRING = 0x0143
	CB_SETCURSEL = 0x014E
	CB_GETCURSEL = 0x0147
	CB_GETLBTEXT = 0x0148
	CB_RESETCONTENT = 0x014B

	BM_GETCHECK = 0x00F0
	BM_SETCHECK = 0x00F1
	BST_CHECKED = 1
	BST_UNCHECKED = 0

	DEFAULT_GUI_FONT = 17
	FW_NORMAL = 400
	FW_BOLD   = 700
	ANSI_CHARSET = 0
	OUT_TT_PRECIS = 4
	CLIP_DEFAULT_PRECIS = 0
	DEFAULT_QUALITY = 0
	FF_DONTCARE = 0

	SRCCOPY = 0x00CC0020
	BI_RGB  = 0
	DIB_RGB_COLORS = 0

	WM_TRAY        = WM_APP + 1
	WM_SET_OVERLAY = WM_APP + 2
	WM_HIDE_OVERLAY = WM_APP + 3
	WM_CONN        = WM_APP + 4
	WM_OPEN_SETTINGS = WM_APP + 5
	WM_APPLY_FOCUS = WM_APP + 6
	WM_APPLY_TILE  = WM_APP + 7
	WM_APPLY_STACK = WM_APP + 8
	WM_REFRESH     = WM_APP + 9
	WM_ENC         = WM_APP + 10
	WM_TAP         = WM_APP + 11
	WM_BTN         = WM_APP + 12
	WM_DWELL       = WM_APP + 13
	WM_EXIT        = WM_APP + 14
)

type POINT struct {
	X, Y int32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type MSG struct {
	Hwnd    windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type WNDCLASSEX struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

type PAINTSTRUCT struct {
	Hdc         windows.Handle
	Erase       int32
	RcPaint     RECT
	Restore     int32
	IncUpdate   int32
	RgbReserved [32]byte
}

type MONITORINFO struct {
	Size    uint32
	Monitor RECT
	Work    RECT
	Flags   uint32
}

type NOTIFYICONDATA struct {
	Size            uint32
	Wnd             windows.Handle
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            windows.Handle
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	TimeoutOrVer    uint32
	InfoTitle       [64]uint16
	InfoFlags       uint32
	GuidItem        windows.GUID
	BalloonIcon     windows.Handle
}

type ICONINFO struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  windows.Handle
	HbmColor windows.Handle
}

type BITMAPINFOHEADER struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type BITMAPINFO struct {
	Header BITMAPINFOHEADER
	Colors [1]uint32
}

var (
	modUser32   = windows.NewLazySystemDLL("user32.dll")
	modGdi32    = windows.NewLazySystemDLL("gdi32.dll")
	modShell32  = windows.NewLazySystemDLL("shell32.dll")
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procRegisterClassExW     = modUser32.NewProc("RegisterClassExW")
	procCreateWindowExW      = modUser32.NewProc("CreateWindowExW")
	procDestroyWindow        = modUser32.NewProc("DestroyWindow")
	procDefWindowProcW       = modUser32.NewProc("DefWindowProcW")
	procShowWindow           = modUser32.NewProc("ShowWindow")
	procUpdateWindow         = modUser32.NewProc("UpdateWindow")
	procInvalidateRect       = modUser32.NewProc("InvalidateRect")
	procGetMessageW          = modUser32.NewProc("GetMessageW")
	procPeekMessageW         = modUser32.NewProc("PeekMessageW")
	procTranslateMessage     = modUser32.NewProc("TranslateMessage")
	procDispatchMessageW     = modUser32.NewProc("DispatchMessageW")
	procPostMessageW         = modUser32.NewProc("PostMessageW")
	procPostQuitMessage      = modUser32.NewProc("PostQuitMessage")
	procSetTimer             = modUser32.NewProc("SetTimer")
	procKillTimer            = modUser32.NewProc("KillTimer")
	procBeginPaint           = modUser32.NewProc("BeginPaint")
	procEndPaint             = modUser32.NewProc("EndPaint")
	procFillRect             = modUser32.NewProc("FillRect")
	procDrawTextW            = modUser32.NewProc("DrawTextW")
	procSetWindowPos         = modUser32.NewProc("SetWindowPos")
	procSetLayeredWindowAttributes = modUser32.NewProc("SetLayeredWindowAttributes")
	procGetWindowRect        = modUser32.NewProc("GetWindowRect")
	procGetClientRect        = modUser32.NewProc("GetClientRect")
	procEnumWindows          = modUser32.NewProc("EnumWindows")
	procIsWindowVisible      = modUser32.NewProc("IsWindowVisible")
	procIsIconic             = modUser32.NewProc("IsIconic")
	procGetWindowTextW       = modUser32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = modUser32.NewProc("GetWindowTextLengthW")
	procGetWindowThreadProcessId = modUser32.NewProc("GetWindowThreadProcessId")
	procGetClassNameW        = modUser32.NewProc("GetClassNameW")
	procSetForegroundWindow  = modUser32.NewProc("SetForegroundWindow")
	procBringWindowToTop     = modUser32.NewProc("BringWindowToTop")
	procGetForegroundWindow  = modUser32.NewProc("GetForegroundWindow")
	procAttachThreadInput    = modUser32.NewProc("AttachThreadInput")
	procAllowSetForegroundWindow = modUser32.NewProc("AllowSetForegroundWindow")
	procEnumDisplayMonitors  = modUser32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW      = modUser32.NewProc("GetMonitorInfoW")
	procMonitorFromPoint     = modUser32.NewProc("MonitorFromPoint")
	procMonitorFromWindow    = modUser32.NewProc("MonitorFromWindow")
	procGetSystemMetrics     = modUser32.NewProc("GetSystemMetrics")
	procGetCursorPos         = modUser32.NewProc("GetCursorPos")
	procSetForeground        = modUser32.NewProc("SetForegroundWindow")
	procCreatePopupMenu      = modUser32.NewProc("CreatePopupMenu")
	procAppendMenuW          = modUser32.NewProc("AppendMenuW")
	procTrackPopupMenu       = modUser32.NewProc("TrackPopupMenu")
	procDestroyMenu          = modUser32.NewProc("DestroyMenu")
	procSetMenuDefaultItem   = modUser32.NewProc("SetMenuDefaultItem")
	procSendMessageW         = modUser32.NewProc("SendMessageW")
	procGetDlgItem           = modUser32.NewProc("GetDlgItem")
	procSetWindowTextW       = modUser32.NewProc("SetWindowTextW")
	procGetWindowTextW2      = modUser32.NewProc("GetWindowTextW")
	procIsWindow             = modUser32.NewProc("IsWindow")
	procSetWindowRgn         = modUser32.NewProc("SetWindowRgn")
	procLoadCursorW          = modUser32.NewProc("LoadCursorW")
	procGetDC                = modUser32.NewProc("GetDC")
	procReleaseDC            = modUser32.NewProc("ReleaseDC")
	procSetFocus             = modUser32.NewProc("SetFocus")
	procEnableWindow         = modUser32.NewProc("EnableWindow")

	procCreateSolidBrush     = modGdi32.NewProc("CreateSolidBrush")
	procDeleteObject         = modGdi32.NewProc("DeleteObject")
	procSetBkMode            = modGdi32.NewProc("SetBkMode")
	procSetTextColor         = modGdi32.NewProc("SetTextColor")
	procSelectObject         = modGdi32.NewProc("SelectObject")
	procCreateFontW          = modGdi32.NewProc("CreateFontW")
	procGetStockObject       = modGdi32.NewProc("GetStockObject")
	procCreateCompatibleDC   = modGdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = modGdi32.NewProc("CreateCompatibleBitmap")
	procCreateDIBSection     = modGdi32.NewProc("CreateDIBSection")
	procCreateBitmap         = modGdi32.NewProc("CreateBitmap")
	procBitBlt               = modGdi32.NewProc("BitBlt")
	procStretchBlt           = modGdi32.NewProc("StretchBlt")
	procSetStretchBltMode    = modGdi32.NewProc("SetStretchBltMode")
	procDeleteDC             = modGdi32.NewProc("DeleteDC")
	procCreateRoundRectRgn   = modGdi32.NewProc("CreateRoundRectRgn")
	procCreateEllipticRgn    = modGdi32.NewProc("CreateEllipticRgn")
	procEllipse              = modGdi32.NewProc("Ellipse")
	procMoveToEx             = modGdi32.NewProc("MoveToEx")
	procLineTo               = modGdi32.NewProc("LineTo")
	procCreatePen            = modGdi32.NewProc("CreatePen")
	procSetBkColor           = modGdi32.NewProc("SetBkColor")

	procCreateIconIndirect   = modUser32.NewProc("CreateIconIndirect")
	procDestroyIcon          = modUser32.NewProc("DestroyIcon")
	procShellNotifyIconW     = modShell32.NewProc("Shell_NotifyIconW")

	procGetModuleHandleW     = modKernel32.NewProc("GetModuleHandleW")
	procGetCurrentThreadId   = modKernel32.NewProc("GetCurrentThreadId")
	procFreeConsole          = modKernel32.NewProc("FreeConsole")
	procSetCapture           = modUser32.NewProc("SetCapture")
	procReleaseCapture       = modUser32.NewProc("ReleaseCapture")
)

func LOWORD(v uintptr) uint16 { return uint16(v & 0xFFFF) }
func HIWORD(v uintptr) uint16 { return uint16((v >> 16) & 0xFFFF) }

func GetModuleHandle() windows.Handle {
	r, _, _ := procGetModuleHandleW.Call(0)
	return windows.Handle(r)
}

func RegisterClass(name string, proc uintptr, bg windows.Handle) error {
	cn, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	var wc WNDCLASSEX
	wc.Size = uint32(unsafe.Sizeof(wc))
	wc.Style = CS_HREDRAW | CS_VREDRAW
	wc.WndProc = proc
	wc.Instance = GetModuleHandle()
	cur, _, _ := procLoadCursorW.Call(0, 32512) // IDC_ARROW
	wc.Cursor = windows.Handle(cur)
	wc.Background = bg
	wc.ClassName = cn
	r, _, e := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		if errno, ok := e.(syscall.Errno); ok && errno == 1410 { // already registered
			return nil
		}
		return e
	}
	return nil
}

func CreateWindow(ex, style uint32, class, title string, x, y, w, h int32, parent windows.Handle, menu windows.Handle) (windows.Handle, error) {
	cn, err := syscall.UTF16PtrFromString(class)
	if err != nil {
		return 0, err
	}
	var tn *uint16
	if title != "" {
		tn, err = syscall.UTF16PtrFromString(title)
		if err != nil {
			return 0, err
		}
	}
	r, _, e := procCreateWindowExW.Call(
		uintptr(ex),
		uintptr(unsafe.Pointer(cn)),
		uintptr(unsafe.Pointer(tn)),
		uintptr(style),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		uintptr(parent),
		uintptr(menu),
		uintptr(GetModuleHandle()),
		0,
	)
	if r == 0 {
		return 0, e
	}
	return windows.Handle(r), nil
}

func DestroyWindow(h windows.Handle) {
	procDestroyWindow.Call(uintptr(h))
}

func DefWindowProc(hwnd windows.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func ShowWindow(h windows.Handle, cmd int) {
	procShowWindow.Call(uintptr(h), uintptr(cmd))
}

func Invalidate(h windows.Handle) {
	procInvalidateRect.Call(uintptr(h), 0, 1)
}

func Post(h windows.Handle, msg uint32, wParam, lParam uintptr) {
	procPostMessageW.Call(uintptr(h), uintptr(msg), wParam, lParam)
}

func PostQuit(code int) {
	procPostQuitMessage.Call(uintptr(code))
}

func Send(h windows.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procSendMessageW.Call(uintptr(h), uintptr(msg), wParam, lParam)
	return r
}

func GetMessage(m *MSG) int {
	r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(m)), 0, 0, 0)
	return int(int32(r))
}

func PeekMessage(m *MSG) bool {
	r, _, _ := procPeekMessageW.Call(uintptr(unsafe.Pointer(m)), 0, 0, 0, 1) // PM_REMOVE
	return r != 0
}

func Translate(m *MSG) { procTranslateMessage.Call(uintptr(unsafe.Pointer(m))) }
func Dispatch(m *MSG)  { procDispatchMessageW.Call(uintptr(unsafe.Pointer(m))) }

func SetTimer(h windows.Handle, id, ms uintptr) {
	procSetTimer.Call(uintptr(h), id, ms, 0)
}

func KillTimer(h windows.Handle, id uintptr) {
	procKillTimer.Call(uintptr(h), id)
}

func BeginPaint(h windows.Handle, ps *PAINTSTRUCT) windows.Handle {
	r, _, _ := procBeginPaint.Call(uintptr(h), uintptr(unsafe.Pointer(ps)))
	return windows.Handle(r)
}

func EndPaint(h windows.Handle, ps *PAINTSTRUCT) {
	procEndPaint.Call(uintptr(h), uintptr(unsafe.Pointer(ps)))
}

func CreateBrush(rgb uint32) windows.Handle {
	r, _, _ := procCreateSolidBrush.Call(uintptr(rgb))
	return windows.Handle(r)
}

func DeleteObject(h windows.Handle) {
	procDeleteObject.Call(uintptr(h))
}

func FillRect(hdc windows.Handle, rc *RECT, brush windows.Handle) {
	procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(rc)), uintptr(brush))
}

func RGB(r, g, b uint8) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func SetBkMode(hdc windows.Handle, mode int) {
	procSetBkMode.Call(uintptr(hdc), uintptr(mode))
}

func SetTextColor(hdc windows.Handle, rgb uint32) {
	procSetTextColor.Call(uintptr(hdc), uintptr(rgb))
}

func SetBkColor(hdc windows.Handle, rgb uint32) {
	procSetBkColor.Call(uintptr(hdc), uintptr(rgb))
}

func DrawText(hdc windows.Handle, s string, rc *RECT, fmt uint32) {
	p, _ := syscall.UTF16PtrFromString(s)
	n := int32(len([]rune(s)))
	procDrawTextW.Call(uintptr(hdc), uintptr(unsafe.Pointer(p)), uintptr(n), uintptr(unsafe.Pointer(rc)), uintptr(fmt))
}

func SelectObject(hdc, obj windows.Handle) windows.Handle {
	r, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(obj))
	return windows.Handle(r)
}

func CreateFont(h int32, weight int32, name string) windows.Handle {
	pn, _ := syscall.UTF16PtrFromString(name)
	r, _, _ := procCreateFontW.Call(
		uintptr(h), 0, 0, 0,
		uintptr(weight), 0, 0, 0,
		ANSI_CHARSET, OUT_TT_PRECIS, CLIP_DEFAULT_PRECIS, DEFAULT_QUALITY, FF_DONTCARE,
		uintptr(unsafe.Pointer(pn)),
	)
	return windows.Handle(r)
}

func GetStockFont() windows.Handle {
	r, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)
	return windows.Handle(r)
}

func NullBrush() windows.Handle {
	r, _, _ := procGetStockObject.Call(NULL_BRUSH)
	return windows.Handle(r)
}

func SetWindowPos(h windows.Handle, insertAfter uintptr, x, y, w, ht int32, flags uint32) {
	procSetWindowPos.Call(uintptr(h), insertAfter, uintptr(x), uintptr(y), uintptr(w), uintptr(ht), uintptr(flags))
}

func SetLayered(h windows.Handle, alpha byte) {
	procSetLayeredWindowAttributes.Call(uintptr(h), 0, uintptr(alpha), LWA_ALPHA)
}

func GetClientRect(h windows.Handle) RECT {
	var rc RECT
	procGetClientRect.Call(uintptr(h), uintptr(unsafe.Pointer(&rc)))
	return rc
}

func GetWindowRect(h windows.Handle) RECT {
	var rc RECT
	procGetWindowRect.Call(uintptr(h), uintptr(unsafe.Pointer(&rc)))
	return rc
}

func EnumWindows(cb uintptr) {
	procEnumWindows.Call(cb, 0)
}

func IsWindowVisible(h windows.Handle) bool {
	r, _, _ := procIsWindowVisible.Call(uintptr(h))
	return r != 0
}

func IsIconic(h windows.Handle) bool {
	r, _, _ := procIsIconic.Call(uintptr(h))
	return r != 0
}

func IsWindow(h windows.Handle) bool {
	r, _, _ := procIsWindow.Call(uintptr(h))
	return r != 0
}

func GetWindowText(h windows.Handle) string {
	n, _, _ := procGetWindowTextLengthW.Call(uintptr(h))
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+2)
	procGetWindowTextW.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func GetClassName(h windows.Handle) string {
	buf := make([]uint16, 256)
	procGetClassNameW.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func GetWindowThreadProcessId(h windows.Handle) (tid uint32, pid uint32) {
	var p uint32
	r, _, _ := procGetWindowThreadProcessId.Call(uintptr(h), uintptr(unsafe.Pointer(&p)))
	return uint32(r), p
}

func GetCurrentThreadId() uint32 {
	r, _, _ := procGetCurrentThreadId.Call()
	return uint32(r)
}

func FreeConsole() {
	procFreeConsole.Call()
}

func SetCapture(h windows.Handle) {
	procSetCapture.Call(uintptr(h))
}

func ReleaseCapture() {
	procReleaseCapture.Call()
}

func MouseXY(lParam uintptr) (int32, int32) {
	return int32(int16(lParam & 0xFFFF)), int32(int16((lParam >> 16) & 0xFFFF))
}

func GetSystemMetrics(idx int) int {
	r, _, _ := procGetSystemMetrics.Call(uintptr(idx))
	return int(int32(r))
}

func CascadeStep() int {
	step := GetSystemMetrics(SM_CYCAPTION) + GetSystemMetrics(SM_CYFRAME)
	if step < 32 {
		step = 32
	}
	return step
}

func WorkRECTFromWindow(h windows.Handle) RECT {
	mon, _, _ := procMonitorFromWindow.Call(uintptr(h), MONITOR_DEFAULTTONEAREST)
	if mon == 0 {
		return PrimaryWorkRECT()
	}
	var mi MONITORINFO
	mi.Size = uint32(unsafe.Sizeof(mi))
	procGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
	if mi.Work.Right <= mi.Work.Left || mi.Work.Bottom <= mi.Work.Top {
		return PrimaryWorkRECT()
	}
	return mi.Work
}

func SetForegroundWindow(h windows.Handle) bool {
	r, _, _ := procSetForegroundWindow.Call(uintptr(h))
	return r != 0
}

func BringWindowToTop(h windows.Handle) {
	procBringWindowToTop.Call(uintptr(h))
}

func GetForegroundWindow() windows.Handle {
	r, _, _ := procGetForegroundWindow.Call()
	return windows.Handle(r)
}

func AttachThreadInput(a, b uint32, attach bool) {
	v := uintptr(0)
	if attach {
		v = 1
	}
	procAttachThreadInput.Call(uintptr(a), uintptr(b), v)
}

func AllowSetForegroundWindow(pid uint32) {
	procAllowSetForegroundWindow.Call(uintptr(pid))
}

func GetCursorPos() POINT {
	var p POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	return p
}

func ShellNotifyIcon(cmd uint32, data *NOTIFYICONDATA) bool {
	data.Size = uint32(unsafe.Sizeof(*data))
	r, _, _ := procShellNotifyIconW.Call(uintptr(cmd), uintptr(unsafe.Pointer(data)))
	return r != 0
}

func UTF16Copy(dst []uint16, s string) {
	u, _ := syscall.UTF16FromString(s)
	n := len(u)
	if n > len(dst) {
		n = len(dst)
		u = u[:n]
		u[n-1] = 0
	}
	copy(dst, u)
}

func CreatePopupMenu() windows.Handle {
	r, _, _ := procCreatePopupMenu.Call()
	return windows.Handle(r)
}

func AppendMenu(menu windows.Handle, flags uint32, id uintptr, s string) {
	p, _ := syscall.UTF16PtrFromString(s)
	procAppendMenuW.Call(uintptr(menu), uintptr(flags), id, uintptr(unsafe.Pointer(p)))
}

func TrackPopupMenu(menu windows.Handle, flags uint32, x, y int32, owner windows.Handle) uintptr {
	r, _, _ := procTrackPopupMenu.Call(uintptr(menu), uintptr(flags), uintptr(x), uintptr(y), 0, uintptr(owner), 0)
	return r
}

func DestroyMenu(m windows.Handle) {
	procDestroyMenu.Call(uintptr(m))
}

func CreateRoundRectRgn(x1, y1, x2, y2, w, h int32) windows.Handle {
	r, _, _ := procCreateRoundRectRgn.Call(uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2), uintptr(w), uintptr(h))
	return windows.Handle(r)
}

func CreateEllipticRgn(x1, y1, x2, y2 int32) windows.Handle {
	r, _, _ := procCreateEllipticRgn.Call(uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2))
	return windows.Handle(r)
}

func Ellipse(hdc windows.Handle, x1, y1, x2, y2 int32) {
	procEllipse.Call(uintptr(hdc), uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2))
}

func MoveTo(hdc windows.Handle, x, y int32) {
	procMoveToEx.Call(uintptr(hdc), uintptr(x), uintptr(y), 0)
}

func LineTo(hdc windows.Handle, x, y int32) {
	procLineTo.Call(uintptr(hdc), uintptr(x), uintptr(y))
}

func CreatePen(width int32, rgb uint32) windows.Handle {
	r, _, _ := procCreatePen.Call(0, uintptr(width), uintptr(rgb))
	return windows.Handle(r)
}

func SetWindowRgn(h, rgn windows.Handle, redraw bool) {
	v := uintptr(0)
	if redraw {
		v = 1
	}
	procSetWindowRgn.Call(uintptr(h), uintptr(rgn), v)
}

func ProcessImageName(pid uint32) string {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)
	var buf [260]uint16
	n := uint32(len(buf))
	err = windows.QueryFullProcessImageName(h, 0, &buf[0], &n)
	if err != nil {
		return ""
	}
	s := syscall.UTF16ToString(buf[:n])
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '\\' || s[i] == '/' {
			return s[i+1:]
		}
	}
	return s
}

func DestroyIcon(h windows.Handle) {
	procDestroyIcon.Call(uintptr(h))
}

func CreateIconIndirect(ii *ICONINFO) windows.Handle {
	r, _, _ := procCreateIconIndirect.Call(uintptr(unsafe.Pointer(ii)))
	return windows.Handle(r)
}

func CreateDIBSection(hdc windows.Handle, w, h int32) (windows.Handle, []byte, error) {
	var bi BITMAPINFO
	bi.Header.Size = uint32(unsafe.Sizeof(bi.Header))
	bi.Header.Width = w
	bi.Header.Height = -h
	bi.Header.Planes = 1
	bi.Header.BitCount = 32
	bi.Header.Compression = BI_RGB
	var bits uintptr
	r, _, e := procCreateDIBSection.Call(uintptr(hdc), uintptr(unsafe.Pointer(&bi)), DIB_RGB_COLORS, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if r == 0 {
		return 0, nil, e
	}
	n := int(w * h * 4)
	slice := unsafe.Slice((*byte)(unsafe.Pointer(bits)), n)
	return windows.Handle(r), slice, nil
}

func CreateBitmap(w, h int32, bits uintptr) windows.Handle {
	r, _, _ := procCreateBitmap.Call(uintptr(w), uintptr(h), 1, 1, bits)
	return windows.Handle(r)
}

func GetDC(h windows.Handle) windows.Handle {
	r, _, _ := procGetDC.Call(uintptr(h))
	return windows.Handle(r)
}

func ReleaseDC(hwnd, hdc windows.Handle) {
	procReleaseDC.Call(uintptr(hwnd), uintptr(hdc))
}

func CreateCompatibleDC(hdc windows.Handle) windows.Handle {
	r, _, _ := procCreateCompatibleDC.Call(uintptr(hdc))
	return windows.Handle(r)
}

func BitBlt(dst windows.Handle, x, y, w, h int32, src windows.Handle) {
	procBitBlt.Call(uintptr(dst), uintptr(x), uintptr(y), uintptr(w), uintptr(h), uintptr(src), 0, 0, SRCCOPY)
}

func StretchBlt(dst windows.Handle, dx, dy, dw, dh int32, src windows.Handle, sx, sy, sw, sh int32) {
	procStretchBlt.Call(uintptr(dst), uintptr(dx), uintptr(dy), uintptr(dw), uintptr(dh), uintptr(src), uintptr(sx), uintptr(sy), uintptr(sw), uintptr(sh), SRCCOPY)
}

func SetStretchBltMode(hdc windows.Handle, mode int) {
	procSetStretchBltMode.Call(uintptr(hdc), uintptr(mode))
}

func DeleteDC(hdc windows.Handle) {
	procDeleteDC.Call(uintptr(hdc))
}

func PrimaryWorkRECT() RECT {
	type acc struct {
		found bool
		rc    RECT
	}
	a := &acc{}
	cb := syscall.NewCallback(func(hMonitor, hdc, lprc, dwData uintptr) uintptr {
		var mi MONITORINFO
		mi.Size = uint32(unsafe.Sizeof(mi))
		procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&mi)))
		if mi.Flags&MONITORINFOF_PRIMARY != 0 {
			a.found = true
			a.rc = mi.Work
			return 0
		}
		return 1
	})
	procEnumDisplayMonitors.Call(0, 0, cb, 0)
	if a.found {
		return a.rc
	}
	pt := POINT{0, 0}
	mon, _, _ := procMonitorFromPoint.Call(uintptr(*(*uint64)(unsafe.Pointer(&pt))), MONITOR_DEFAULTTOPRIMARY)
	var mi MONITORINFO
	mi.Size = uint32(unsafe.Sizeof(mi))
	procGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
	return mi.Work
}

func SetCheck(h windows.Handle, on bool) {
	v := uintptr(BST_UNCHECKED)
	if on {
		v = BST_CHECKED
	}
	Send(h, BM_SETCHECK, v, 0)
}

func GetCheck(h windows.Handle) bool {
	return Send(h, BM_GETCHECK, 0, 0) == BST_CHECKED
}

func ComboReset(h windows.Handle) { Send(h, CB_RESETCONTENT, 0, 0) }

func ComboAdd(h windows.Handle, s string) {
	p, _ := syscall.UTF16PtrFromString(s)
	Send(h, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(p)))
}

func ComboSet(h windows.Handle, idx int) { Send(h, CB_SETCURSEL, uintptr(idx), 0) }

func ComboGet(h windows.Handle) int { return int(int32(Send(h, CB_GETCURSEL, 0, 0))) }

func ComboText(h windows.Handle, idx int) string {
	buf := make([]uint16, 256)
	Send(h, CB_GETLBTEXT, uintptr(idx), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}

func SetWindowText(h windows.Handle, s string) {
	p, _ := syscall.UTF16PtrFromString(s)
	procSetWindowTextW.Call(uintptr(h), uintptr(unsafe.Pointer(p)))
}
