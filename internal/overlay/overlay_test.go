package overlay

import (
	"image"
	"strings"
	"testing"
)

func TestRingAnglePutsSelectionAtTop(t *testing.T) {
	top := RingAngle(3, 3, 8)
	if top > -1.57 || top < -1.58 {
		t.Fatalf("sel at top got %v", top)
	}
	if RingAngle(0, 0, 1) > -1.57 || RingAngle(0, 0, 1) < -1.58 {
		t.Fatalf("n=1 %v", RingAngle(0, 0, 1))
	}
}

func TestBoundsGraphicalIsWideAndCentered(t *testing.T) {
	work := image.Rect(0, 0, 1920, 1032)
	b := BoundsGraphical(work)
	if b.Dx() != GraphWidth || b.Dy() != GraphHeight {
		t.Fatalf("size %v", b)
	}
	if b.Min.X != (work.Dx()-GraphWidth)/2 || b.Min.Y != (work.Dy()-GraphHeight)/2 {
		t.Fatalf("not centered %v", b)
	}
}

func TestVisibleStartKeepsFiveContiguousItems(t *testing.T) {
	if got := VisibleStart(9, 0, 5); got != 0 {
		t.Fatalf("first %d", got)
	}
	if got := VisibleStart(9, 4, 5); got != 2 {
		t.Fatalf("middle %d", got)
	}
	if got := VisibleStart(9, 8, 5); got != 4 {
		t.Fatalf("last %d", got)
	}
	if got := VisibleStart(3, 2, 5); got != 0 {
		t.Fatalf("short %d", got)
	}
}

func TestBoundsCenteredOnPrimaryWork(t *testing.T) {
	work := image.Rect(0, 0, 1920, 1032)
	b := Bounds(work)
	if b.Dx() != Width {
		t.Fatalf("width %d", b.Dx())
	}
	if b.Min.X != (1920-Width)/2 {
		t.Fatalf("x %d", b.Min.X)
	}
	if b.Min.Y < work.Min.Y || b.Max.Y > work.Max.Y {
		t.Fatalf("y %v", b)
	}
}

func TestStepWraps(t *testing.T) {
	if Step(5, 0, -1) != 4 || Step(5, 4, 1) != 0 || Step(0, 0, 1) != 0 {
		t.Fatal(Step(5, 0, -1), Step(5, 4, 1), Step(0, 0, 1))
	}
}

func TestFormatRowTruncateAndSelect(t *testing.T) {
	long := strings.Repeat("a", 80)
	s := FormatRow(true, "grok", long)
	if !strings.HasPrefix(s, "> GROK") {
		t.Fatalf("%q", s)
	}
	if strings.Count(s, "a") != 64 {
		t.Fatalf("trunc %q", s)
	}
	u := FormatRow(false, "unknown", "Current plan")
	if u != "  Current plan" {
		t.Fatalf("%q", u)
	}
	empty := FormatRow(true, "", "Ultracode plan")
	if empty != "> Ultracode plan" {
		t.Fatalf("%q", empty)
	}
}
