package overlay

import (
	"fmt"
	"image"
	"sync"

	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

type themeBitmap struct {
	id  string
	w   int32
	h   int32
	hdc windows.Handle
	bmp windows.Handle
}

var (
	themeBmpMu sync.Mutex
	themeBmps  = map[string]themeBitmap{}
)

func DrawTheme(hdc windows.Handle, id string, dest win32.RECT) bool {
	w := dest.Right - dest.Left
	h := dest.Bottom - dest.Top
	if w <= 0 || h <= 0 {
		return false
	}
	tb := bitmapFor(id, w, h)
	if tb.bmp == 0 {
		return false
	}
	win32.BitBlt(hdc, dest.Left, dest.Top, w, h, tb.hdc)
	return true
}

func bitmapFor(id string, w, h int32) themeBitmap {
	id = NormalizeTheme(id)
	key := fmt.Sprintf("%s:%d:%d", id, w, h)
	themeBmpMu.Lock()
	defer themeBmpMu.Unlock()
	if tb, ok := themeBmps[key]; ok && tb.bmp != 0 {
		return tb
	}
	im := ThemeImage(id)
	if im == nil {
		return themeBitmap{}
	}
	screen := win32.GetDC(0)
	hdc := win32.CreateCompatibleDC(screen)
	bmp, bits, err := win32.CreateDIBSection(hdc, w, h)
	win32.ReleaseDC(0, screen)
	if err != nil || bmp == 0 {
		return themeBitmap{}
	}
	win32.SelectObject(hdc, bmp)
	scaleInto(im, bits, int(w), int(h))
	tb := themeBitmap{id: id, w: w, h: h, hdc: hdc, bmp: bmp}
	themeBmps[key] = tb
	return tb
}

func scaleInto(im image.Image, bits []byte, w, h int) {
	b := im.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw <= 0 || sh <= 0 {
		return
	}
	for y := 0; y < h; y++ {
		sy := b.Min.Y + y*sh/h
		for x := 0; x < w; x++ {
			sx := b.Min.X + x*sw/w
			r, g, bl, a := im.At(sx, sy).RGBA()
			i := (y*w + x) * 4
			bits[i+0] = byte(bl >> 8)
			bits[i+1] = byte(g >> 8)
			bits[i+2] = byte(r >> 8)
			bits[i+3] = byte(a >> 8)
		}
	}
}
