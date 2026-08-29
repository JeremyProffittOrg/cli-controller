//go:build live

package wins

import (
	"image"
	"strings"
	"testing"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
)

func TestLiveEnumerateNoChrome(t *testing.T) {
	list := Enumerate(config.Default())
	for _, w := range list {
		p := strings.ToLower(w.Process)
		if p == "chrome" || p == "msedge" || p == "firefox" {
			t.Fatalf("browser included: %+v", w)
		}
		if w.Brand == BrandNone {
			t.Fatalf("none brand %+v", w)
		}
	}
}

func TestLiveTileRectsInsideWork(t *testing.T) {
	work := PrimaryWorkArea()
	if work.Dx() < 800 || work.Dy() < 500 {
		t.Fatalf("work %v", work)
	}
	list := Enumerate(config.Default())
	n := len(list)
	if n == 0 {
		n = 4
	}
	rects := TileRects(work, n)
	if !RectsInside(work, rects) {
		t.Fatalf("tile outside %v work %v", rects, work)
	}
	if !RectsInside(work, StackRects(work, n)) {
		t.Fatal("stack outside")
	}
	want := image.Rect(0, 0, 1920, 1032)
	_ = want
}

func TestLiveFocus(t *testing.T) {
	list := Enumerate(config.Default())
	if len(list) == 0 {
		t.Skip("no CLI windows")
	}
	if err := Focus(list[0].HWND); err != nil {
		t.Fatal(err)
	}
}
