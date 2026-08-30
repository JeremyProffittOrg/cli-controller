package settings

import "testing"

func TestAngleFromPointCompass(t *testing.T) {
	if g := AngleFromPoint(100, 100, 100, 20); g != 0 {
		t.Fatalf("up %d", g)
	}
	if g := AngleFromPoint(100, 100, 180, 100); g != 90 {
		t.Fatalf("right %d", g)
	}
	if g := AngleFromPoint(100, 100, 100, 180); g != 180 {
		t.Fatalf("down %d", g)
	}
	if g := AngleFromPoint(100, 100, 20, 100); g != 270 {
		t.Fatalf("left %d", g)
	}
	g := AngleFromPoint(100, 100, 100-40, 100-40)
	if g < 310 || g > 320 {
		t.Fatalf("nw want ~315 got %d", g)
	}
}

func TestNeedleEndRoundTrip(t *testing.T) {
	x, y := NeedleEnd(100, 100, 50, 315)
	got := AngleFromPoint(100, 100, x, y)
	if got < 313 || got > 317 {
		t.Fatalf("roundtrip %d from %d,%d", got, x, y)
	}
}
