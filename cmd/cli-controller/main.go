package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"runtime"
	"time"

	"github.com/JeremyProffittOrg/cli-controller/internal/app"
	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"github.com/JeremyProffittOrg/cli-controller/internal/overlay"
	"github.com/JeremyProffittOrg/cli-controller/internal/win32"
	"github.com/JeremyProffittOrg/cli-controller/internal/wins"
)

func main() {
	runtime.LockOSThread()
	selftest := flag.Bool("selftest", false, "enumerate windows and layout, then exit")
	preview := flag.Bool("preview", false, "show overlay HUD for 15s then exit")
	flag.Parse()
	if *selftest {
		if err := runSelftest(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *preview {
		if err := runPreview(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runSelftest() error {
	cfg := config.Default()
	list := wins.Enumerate(cfg)
	for _, w := range list {
		p := w.Process
		if p == "chrome" || p == "msedge" || p == "firefox" {
			return fmt.Errorf("browser included: %s %s", p, w.Title)
		}
	}
	work := wins.PrimaryWorkArea()
	n := len(list)
	if n == 0 {
		n = 4
	}
	tr := wins.TileRects(work, n)
	if !wins.RectsInside(work, tr) {
		return fmt.Errorf("tile outside work %v", work)
	}
	sr := wins.StackRects(work, n)
	if !wins.RectsInside(work, sr) {
		return fmt.Errorf("stack outside work %v", work)
	}
	if len(list) > 0 {
		works := make([]image.Rectangle, len(list))
		for i, w := range list {
			works[i] = w.Work
			if works[i].Empty() {
				works[i] = work
			}
		}
		ps := wins.PerScreenRects(works, wins.TileRects)
		for i, r := range ps {
			if !wins.RectsInside(works[i], []image.Rectangle{r}) {
				return fmt.Errorf("per-screen tile %d %v not in %v", i, r, works[i])
			}
		}
	}
	b := overlay.Bounds(work)
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return fmt.Errorf("overlay bounds %v", b)
	}
	if len(list) > 0 {
		if err := wins.Focus(list[0].HWND); err != nil {
			return err
		}
	}
	fmt.Printf("selftest ok windows=%d work=%v overlay=%v\n", len(list), work, b)
	return nil
}

func runPreview() error {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	ov, err := overlay.New()
	if err != nil {
		return err
	}
	list := wins.Enumerate(cfg)
	items := make([]overlay.Item, len(list))
	for i, w := range list {
		items[i] = overlay.Item{Brand: string(w.Brand), Title: w.Title}
	}
	sel := 0
	if len(items) > 2 {
		sel = 2
	} else if len(items) > 1 {
		sel = 1
	}
	ov.Show(wins.PrimaryWorkArea(), items, sel)
	fmt.Printf("preview overlay items=%d sel=%d\n", len(items), sel)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		var msg win32.MSG
		for win32.PeekMessage(&msg) {
			if msg.Message == win32.WM_QUIT {
				return nil
			}
			win32.Translate(&msg)
			win32.Dispatch(&msg)
		}
		time.Sleep(16 * time.Millisecond)
	}
	ov.Hide()
	return nil
}
