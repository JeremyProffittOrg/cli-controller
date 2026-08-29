package wins

import (
	"image"
	"testing"
)

func TestTileRectsFillPrimaryWorkArea(t *testing.T) {
	work := image.Rect(0, 0, 1920, 1032)
	for _, n := range []int{1, 2, 3, 4, 7, 9} {
		rects := TileRects(work, n)
		if len(rects) != n {
			t.Fatalf("n=%d len=%d", n, len(rects))
		}
		if !RectsInside(work, rects) {
			t.Fatalf("n=%d outside work %v", n, rects)
		}
	}
	if TileRects(work, 0) != nil {
		t.Fatal("n=0")
	}
}

func TestPerScreenRectsStayOnOwnScreen(t *testing.T) {
	primary := image.Rect(0, 0, 1920, 1032)
	secondary := image.Rect(-1920, -15, 0, 1017)
	works := []image.Rectangle{primary, primary, secondary}
	got := PerScreenRects(works, TileRects)
	if len(got) != 3 {
		t.Fatal(len(got))
	}
	if !RectsInside(primary, []image.Rectangle{got[0], got[1]}) {
		t.Fatalf("primary %v", got)
	}
	if !RectsInside(secondary, []image.Rectangle{got[2]}) {
		t.Fatalf("secondary %v", got[2])
	}
	stacked := PerScreenRects(works, StackRects)
	if !RectsInside(primary, []image.Rectangle{stacked[0], stacked[1]}) || !RectsInside(secondary, []image.Rectangle{stacked[2]}) {
		t.Fatalf("stack %v", stacked)
	}
}

func TestStackRectsCascade(t *testing.T) {
	work := image.Rect(0, 0, 1920, 1032)
	rects := StackRects(work, 4)
	if len(rects) != 4 {
		t.Fatal(len(rects))
	}
	if !RectsInside(work, rects) {
		t.Fatalf("outside %v", rects)
	}
	if rects[1].Min.X-rects[0].Min.X != 32 || rects[1].Min.Y-rects[0].Min.Y != 32 {
		t.Fatalf("offset %v", rects)
	}
	if rects[0].Dx() != work.Dx()*70/100 {
		t.Fatalf("width %d", rects[0].Dx())
	}
}
