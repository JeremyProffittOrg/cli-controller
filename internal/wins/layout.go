package wins

import (
	"image"
	"math"
)

func TileRects(work image.Rectangle, n int) []image.Rectangle {
	if n <= 0 {
		return nil
	}
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	if cols < 1 {
		cols = 1
	}
	rows := int(math.Ceil(float64(n) / float64(cols)))
	cellW := work.Dx() / cols
	cellH := work.Dy() / rows
	out := make([]image.Rectangle, n)
	for i := 0; i < n; i++ {
		r := i / cols
		c := i % cols
		x0 := work.Min.X + c*cellW
		y0 := work.Min.Y + r*cellH
		x1 := x0 + cellW
		y1 := y0 + cellH
		if c == cols-1 {
			x1 = work.Max.X
		}
		if r == rows-1 {
			y1 = work.Max.Y
		}
		out[i] = image.Rect(x0, y0, x1, y1)
	}
	return out
}

func StackRects(work image.Rectangle, n int) []image.Rectangle {
	return StackRectsOffset(work, n, 32)
}

func StackRectsOffset(work image.Rectangle, n, step int) []image.Rectangle {
	if n <= 0 {
		return nil
	}
	if step < 24 {
		step = 24
	}
	maxOff := n - 1
	minW := work.Dx() / 2
	minH := work.Dy() / 2
	if minW < 400 {
		minW = 400
	}
	if minH < 300 {
		minH = 300
	}
	if minW > work.Dx() {
		minW = work.Dx()
	}
	if minH > work.Dy() {
		minH = work.Dy()
	}
	for maxOff > 0 {
		if work.Dx()-maxOff*step >= minW && work.Dy()-maxOff*step >= minH {
			break
		}
		maxOff--
	}
	ww := work.Dx() - maxOff*step
	hh := work.Dy() - maxOff*step
	if ww < 1 {
		ww = 1
	}
	if hh < 1 {
		hh = 1
	}
	span := maxOff + 1
	out := make([]image.Rectangle, n)
	for i := 0; i < n; i++ {
		k := 0
		if span > 1 {
			k = i % span
		}
		x0 := work.Min.X + k*step
		y0 := work.Min.Y + k*step
		out[i] = image.Rect(x0, y0, x0+ww, y0+hh)
	}
	return out
}

func PerScreenRects(works []image.Rectangle, layout func(image.Rectangle, int) []image.Rectangle) []image.Rectangle {
	n := len(works)
	if n == 0 {
		return nil
	}
	type key struct{ x0, y0, x1, y1 int }
	order := make([]key, 0)
	groups := map[key][]int{}
	for i, w := range works {
		k := key{w.Min.X, w.Min.Y, w.Max.X, w.Max.Y}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], i)
	}
	out := make([]image.Rectangle, n)
	for _, k := range order {
		idxs := groups[k]
		work := image.Rect(k.x0, k.y0, k.x1, k.y1)
		rects := layout(work, len(idxs))
		for j, i := range idxs {
			if j < len(rects) {
				out[i] = rects[j]
			}
		}
	}
	return out
}

func RectsInside(work image.Rectangle, rects []image.Rectangle) bool {
	for _, r := range rects {
		if r.Min.X < work.Min.X || r.Min.Y < work.Min.Y || r.Max.X > work.Max.X || r.Max.Y > work.Max.Y {
			return false
		}
		if r.Dx() <= 0 || r.Dy() <= 0 {
			return false
		}
	}
	return true
}
