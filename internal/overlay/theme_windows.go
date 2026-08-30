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
	return DrawEmbedded(hdc, ThemeByID(id).File, dest, false)
}

func DrawEmbedded(hdc windows.Handle, path string, dest win32.RECT, alpha bool) bool {
	w := dest.Right - dest.Left
	h := dest.Bottom - dest.Top
	if w <= 0 || h <= 0 {
		return false
	}
	tb := bitmapFor(path, w, h, alpha)
	if tb.bmp == 0 {
		return false
	}
	if alpha {
		win32.AlphaBlend(hdc, dest.Left, dest.Top, w, h, tb.hdc, w, h)
		return true
	}
	win32.BitBlt(hdc, dest.Left, dest.Top, w, h, tb.hdc)
	return true
}

func bitmapFor(path string, w, h int32, alpha bool) themeBitmap {
	key := fmt.Sprintf("%s:%d:%d:%t", path, w, h, alpha)
	themeBmpMu.Lock()
	defer themeBmpMu.Unlock()
	if tb, ok := themeBmps[key]; ok && tb.bmp != 0 {
		return tb
	}
	im := EmbeddedImage(path)
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
	scaleInto(im, bits, int(w), int(h), alpha)
	tb := themeBitmap{id: path, w: w, h: h, hdc: hdc, bmp: bmp}
	themeBmps[key] = tb
	return tb
}

func scaleInto(im image.Image, bits []byte, w, h int, keyBlack bool) {
	b := im.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw <= 0 || sh <= 0 {
		return
	}
	for y := 0; y < h; y++ {
		sy := b.Min.Y + y*sh/h
		for x := 0; x < w; x++ {
			sx := b.Min.X + x*sw/w
			r16, g16, b16, _ := im.At(sx, sy).RGBA()
			r := byte(r16 >> 8)
			g := byte(g16 >> 8)
			bl := byte(b16 >> 8)
			a := byte(255)
			if keyBlack && r < 18 && g < 18 && bl < 18 {
				a = 0
				r, g, bl = 0, 0, 0
			} else if keyBlack {
				r = byte(int(r) * int(a) / 255)
				g = byte(int(g) * int(a) / 255)
				bl = byte(int(bl) * int(a) / 255)
			}
			i := (y*w + x) * 4
			bits[i+0] = bl
			bits[i+1] = g
			bits[i+2] = r
			bits[i+3] = a
		}
	}
}
