package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	cfg := Default()
	cfg.PortMode = "manual"
	cfg.Port = "COM10"
	cfg.LastSerial = "B0:81:84:97:1E:54"
	cfg.OverlayView = "graphical"
	cfg.OverlayTheme = "solar-flare"
	cfg.DisplayRotation = 315
	cfg.Brands["unknown"] = false
	cfg.Brands["grok"] = true
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.PortMode != "manual" || got.Port != "COM10" || got.LastSerial != "B0:81:84:97:1E:54" {
		t.Fatalf("got %+v", got)
	}
	if got.OverlayView != "graphical" {
		t.Fatalf("overlayView %s", got.OverlayView)
	}
	if got.OverlayTheme != "solar-flare" {
		t.Fatalf("overlayTheme %s", got.OverlayTheme)
	}
	if got.DisplayRotation != 315 {
		t.Fatalf("displayRotation %d", got.DisplayRotation)
	}
	if got.Enabled("unknown") {
		t.Fatal("unknown should be off")
	}
	if !got.Enabled("grok") {
		t.Fatal("grok should be on")
	}
	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "config.json" {
		t.Fatalf("path %s", p)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultBrandsAllOn(t *testing.T) {
	b := DefaultBrands()
	for _, n := range BrandNames() {
		if !b[n] {
			t.Fatalf("%s default off", n)
		}
	}
}
