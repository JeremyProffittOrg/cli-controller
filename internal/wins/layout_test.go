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
	if stacked[2] != secondary {
		t.Fatalf("single on secondary %v", stacked[2])
	}
	if stacked[0] == stacked[1] {
		t.Fatalf("primary cascade not offset %v", stacked)
	}
	if !RectsInside(primary, []image.Rectangle{stacked[0], stacked[1]}) {
		t.Fatalf("primary cascade %v", stacked)
	}
}

func TestStackRectsCascadeShowsTitles(t *testing.T) {
	work := image.Rect(0, 0, 1920, 1032)
	one := StackRects(work, 1)
	if len(one) != 1 || one[0] != work {
		t.Fatalf("n=1 %v", one)
	}
	rects := StackRects(work, 4)
	if len(rects) != 4 {
		t.Fatal(len(rects))
	}
	if !RectsInside(work, rects) {
		t.Fatalf("outside %v", rects)
	}
	step := 32
	if rects[1].Min.X-rects[0].Min.X != step || rects[1].Min.Y-rects[0].Min.Y != step {
		t.Fatalf("offset %v", rects)
	}
	wantW := work.Dx() - 3*step
	wantH := work.Dy() - 3*step
	if rects[0].Dx() != wantW || rects[0].Dy() != wantH {
		t.Fatalf("size %dx%d want %dx%d", rects[0].Dx(), rects[0].Dy(), wantW, wantH)
	}
	last := rects[3]
	if last.Max != work.Max {
		t.Fatalf("cascade should reach work max, last=%v work=%v", last, work)
	}
}
