package overlay

import (
	"image"
	"strings"
	"testing"
)

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
	u := FormatRow(false, "", "Current plan")
	if !strings.HasPrefix(u, "  UNKNOWN") {
		t.Fatalf("%q", u)
	}
}
