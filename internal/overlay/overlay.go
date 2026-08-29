package overlay

import (
	"image"
	"strings"
	"unicode/utf8"
)

const (
	Width     = 520
	MaxHeight = 720
	Margin    = 80
	RowH      = 36
	Pad       = 16
)

type Item struct {
	Brand string
	Title string
}

func Bounds(work image.Rectangle) image.Rectangle {
	h := MaxHeight
	if max := work.Dy() - Margin; max < h {
		h = max
	}
	if h < 120 {
		h = 120
	}
	w := Width
	if w > work.Dx()-40 {
		w = work.Dx() - 40
		if w < 200 {
			w = work.Dx()
		}
	}
	x := work.Min.X + (work.Dx()-w)/2
	y := work.Min.Y + (work.Dy()-h)/2
	return image.Rect(x, y, x+w, y+h)
}

func Truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n])
}

func FormatRow(selected bool, brand, title string) string {
	title = Truncate(title, 64)
	b := strings.ToUpper(strings.TrimSpace(brand))
	var line string
	if b == "" || b == "UNKNOWN" {
		line = title
	} else {
		line = b + "  " + title
	}
	if selected {
		return "> " + line
	}
	return "  " + line
}

func Step(n, sel, delta int) int {
	if n <= 0 {
		return 0
	}
	sel = (sel + delta) % n
	if sel < 0 {
		sel += n
	}
	return sel
}

func VisibleCount(box image.Rectangle) int {
	inner := box.Dy() - Pad*2 - 24
	if inner < RowH {
		return 1
	}
	return inner / RowH
}
