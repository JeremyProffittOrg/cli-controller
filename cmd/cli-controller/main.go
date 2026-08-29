package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/JeremyProffittOrg/cli-controller/internal/app"
	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"github.com/JeremyProffittOrg/cli-controller/internal/overlay"
	"github.com/JeremyProffittOrg/cli-controller/internal/wins"
)

func main() {
	runtime.LockOSThread()
	selftest := flag.Bool("selftest", false, "enumerate windows and layout, then exit")
	flag.Parse()
	if *selftest {
		if err := runSelftest(); err != nil {
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
