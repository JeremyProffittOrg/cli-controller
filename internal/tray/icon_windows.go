package tray

import (
	"image"
	"image/color"
	"math"
	"unsafe"

	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"golang.org/x/sys/windows"
)

func DialFrame(size, frame int, connected bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := size/2, size/2
	r := size/2 - 2
	ring := color.RGBA{100, 116, 139, 255}
	face := color.RGBA{30, 41, 59, 255}
	needle := color.RGBA{56, 189, 248, 255}
	if !connected {
		ring = color.RGBA{71, 85, 105, 255}
		face = color.RGBA{51, 65, 85, 255}
		needle = color.RGBA{148, 163, 184, 255}
	}
	fillCircle(img, cx, cy, r, ring)
	fillCircle(img, cx, cy, r-2, face)
	for i := 0; i < 12; i++ {
		a := float64(i) * math.Pi / 6
		x0 := cx + int(float64(r-4)*math.Sin(a))
		y0 := cy - int(float64(r-4)*math.Cos(a))
		x1 := cx + int(float64(r-7)*math.Sin(a))
		y1 := cy - int(float64(r-7)*math.Cos(a))
		line(img, x0, y0, x1, y1, ring)
	}
	ang := float64(frame) * math.Pi / 6
	if !connected {
		ang = 0
	}
	nx := cx + int(float64(r-8)*math.Sin(ang))
	ny := cy - int(float64(r-8)*math.Cos(ang))
	line(img, cx, cy, nx, ny, needle)
	fillCircle(img, cx, cy, 2, needle)
	return img
}

func fillCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	rr := r * r
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= rr {
				set(img, x, y, c)
			}
		}
	}
}

func set(img *image.RGBA, x, y int, c color.RGBA) {
	if !image.Pt(x, y).In(img.Bounds()) {
		return
	}
	img.SetRGBA(x, y, c)
}

func line(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		set(img, x0, y0, c)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func rgbaToIcon(img *image.RGBA) windows.Handle {
	w := int32(img.Bounds().Dx())
	h := int32(img.Bounds().Dy())
	hdc := win32.GetDC(0)
	defer win32.ReleaseDC(0, hdc)
	colorBmp, bits, err := win32.CreateDIBSection(hdc, w, h)
	if err != nil || colorBmp == 0 {
		return 0
	}
	for y := 0; y < int(h); y++ {
		for x := 0; x < int(w); x++ {
			c := img.RGBAAt(x, y)
			i := (y*int(w) + x) * 4
			bits[i+0] = c.B
			bits[i+1] = c.G
			bits[i+2] = c.R
			bits[i+3] = c.A
		}
	}
	maskStride := ((int(w) + 15) / 16) * 2
	mask := make([]byte, maskStride*int(h))
	for y := 0; y < int(h); y++ {
		for x := 0; x < int(w); x++ {
			if img.RGBAAt(x, y).A < 128 {
				mask[y*maskStride+x/8] |= 1 << (7 - uint(x%8))
			}
		}
	}
	maskBmp := win32.CreateBitmap(w, h, uintptr(unsafe.Pointer(&mask[0])))
	ii := win32.ICONINFO{
		FIcon:    1,
		HbmMask:  maskBmp,
		HbmColor: colorBmp,
	}
	icon := win32.CreateIconIndirect(&ii)
	win32.DeleteObject(colorBmp)
	if maskBmp != 0 {
		win32.DeleteObject(maskBmp)
	}
	return icon
}

func makeIcons() (connected, disconnected []windows.Handle) {
	connected = make([]windows.Handle, 12)
	disconnected = make([]windows.Handle, 1)
	for i := 0; i < 12; i++ {
		connected[i] = rgbaToIcon(DialFrame(32, i, true))
	}
	disconnected[0] = rgbaToIcon(DialFrame(32, 0, false))
	return
}
