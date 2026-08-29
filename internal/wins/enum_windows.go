package wins

import (
	"image"
	"path/filepath"
	"strings"
	"syscall"
	"sync"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

type Window struct {
	HWND      windows.Handle
	PID       uint32
	Process   string
	Title     string
	Brand     Brand
	Minimized bool
}

var (
	enumMu  sync.Mutex
	enumBuf []Window
	selfPID = windows.GetCurrentProcessId()
	enumCB  = syscall.NewCallback(enumProc)
)

func enumProc(hwnd, lparam uintptr) uintptr {
	h := windows.Handle(hwnd)
	if !win32.IsWindowVisible(h) {
		return 1
	}
	cls := win32.GetClassName(h)
	if cls == "CLIDialOverlay" || cls == "CLIDialSettings" || cls == "CLIDialHost" {
		return 1
	}
	title := win32.GetWindowText(h)
	if strings.TrimSpace(title) == "" {
		return 1
	}
	_, pid := win32.GetWindowThreadProcessId(h)
	if pid == selfPID {
		return 1
	}
	proc := win32.ProcessImageName(pid)
	proc = strings.TrimSuffix(filepath.Base(proc), ".exe")
	brand := Classify(proc, title)
	if brand == BrandNone {
		return 1
	}
	enumBuf = append(enumBuf, Window{
		HWND:      h,
		PID:       pid,
		Process:   proc,
		Title:     title,
		Brand:     brand,
		Minimized: win32.IsIconic(h),
	})
	return 1
}

func Enumerate(cfg config.Config) []Window {
	enumMu.Lock()
	defer enumMu.Unlock()
	enumBuf = enumBuf[:0]
	win32.EnumWindows(enumCB)
	out := make([]Window, 0, len(enumBuf))
	for _, w := range enumBuf {
		if !cfg.Enabled(string(w.Brand)) {
			continue
		}
		if !win32.IsWindow(w.HWND) {
			continue
		}
		out = append(out, w)
	}
	return out
}

func PrimaryWorkArea() image.Rectangle {
	rc := win32.PrimaryWorkRECT()
	return image.Rect(int(rc.Left), int(rc.Top), int(rc.Right), int(rc.Bottom))
}

func Focus(hwnd windows.Handle) error {
	if !win32.IsWindow(hwnd) {
		return syscall.EINVAL
	}
	if win32.IsIconic(hwnd) {
		win32.ShowWindow(hwnd, win32.SW_RESTORE)
	}
	fg := win32.GetForegroundWindow()
	if fg == hwnd {
		win32.BringWindowToTop(hwnd)
		return nil
	}
	thisTid := win32.GetCurrentThreadId()
	fgTid, _ := win32.GetWindowThreadProcessId(fg)
	tgtTid, pid := win32.GetWindowThreadProcessId(hwnd)
	win32.AllowSetForegroundWindow(pid)
	if fgTid != 0 && fgTid != thisTid {
		win32.AttachThreadInput(thisTid, fgTid, true)
		defer win32.AttachThreadInput(thisTid, fgTid, false)
	}
	if tgtTid != 0 && tgtTid != thisTid && tgtTid != fgTid {
		win32.AttachThreadInput(thisTid, tgtTid, true)
		defer win32.AttachThreadInput(thisTid, tgtTid, false)
	}
	win32.BringWindowToTop(hwnd)
	win32.SetForegroundWindow(hwnd)
	win32.ShowWindow(hwnd, win32.SW_SHOW)
	if win32.GetForegroundWindow() != hwnd {
		win32.SetWindowPos(hwnd, win32.HWND_TOP, 0, 0, 0, 0, win32.SWP_NOMOVE|win32.SWP_NOSIZE|win32.SWP_SHOWWINDOW)
		win32.SetForegroundWindow(hwnd)
	}
	return nil
}

func ApplyRects(windows []Window, rects []image.Rectangle) {
	n := len(windows)
	if len(rects) < n {
		n = len(rects)
	}
	for i := 0; i < n; i++ {
		h := windows[i].HWND
		if !win32.IsWindow(h) {
			continue
		}
		if win32.IsIconic(h) {
			win32.ShowWindow(h, win32.SW_RESTORE)
		}
		r := rects[i]
		win32.SetWindowPos(h, win32.HWND_TOP,
			int32(r.Min.X), int32(r.Min.Y), int32(r.Dx()), int32(r.Dy()),
			win32.SWP_SHOWWINDOW)
	}
}

func Tile(list []Window) {
	if len(list) == 0 {
		return
	}
	ApplyRects(list, TileRects(PrimaryWorkArea(), len(list)))
}

func Stack(list []Window) {
	if len(list) == 0 {
		return
	}
	ApplyRects(list, StackRects(PrimaryWorkArea(), len(list)))
}
