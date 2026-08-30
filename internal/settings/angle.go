package settings

import "math"

func NormalizeDeg(deg int) int {
	deg %= 360
	if deg < 0 {
		deg += 360
	}
	return deg
}

func AngleFromPoint(cx, cy, x, y int) int {
	dx := float64(x - cx)
	dy := float64(y - cy)
	if dx == 0 && dy == 0 {
		return 0
	}
	ang := math.Atan2(dx, -dy) * 180 / math.Pi
	return NormalizeDeg(int(math.Round(ang)))
}

func NeedleEnd(cx, cy, r, deg int) (int, int) {
	rad := float64(NormalizeDeg(deg)) * math.Pi / 180
	x := cx + int(math.Round(float64(r)*math.Sin(rad)))
	y := cy - int(math.Round(float64(r)*math.Cos(rad)))
	return x, y
}
