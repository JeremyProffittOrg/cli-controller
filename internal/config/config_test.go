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
	cfg.KneeMode = "confirm"
	cfg.KneeLeftRaises = 3
	cfg.KneeRightDirection = -1
	cfg.KneeChannels[2] = KneeChannel{Role: "left", ThresholdMM: 123}
	cfg.DeskEnabled = true
	cfg.DeskSensitivityMg = 900
	cfg.DeskOrientation = 270
	cfg.DeskForward = "stack"
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
	if got.KneeMode != "confirm" || got.KneeLeftRaises != 3 || got.KneeRightDirection != -1 || got.KneeChannels[2].ThresholdMM != 123 {
		t.Fatalf("knee config %+v", got)
	}
	if !got.DeskEnabled || got.DeskSensitivityMg != 900 || got.DeskOrientation != 270 || got.DeskForward != "stack" {
		t.Fatalf("desk config %+v", got)
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

func TestNormalizeMotionDefaultsAndInvalidValues(t *testing.T) {
	c := Config{DwellMs: 333, KneeMode: "bad", KneeLeftRaises: 9, KneeRightDirection: 0,
		KneeChannels: []KneeChannel{{Role: "bad", ThresholdMM: 4}}, DeskSensitivityMg: 1,
		DeskOrientation: 45, DeskLeft: "bad", DeskRight: "bad", DeskForward: "bad", DeskBack: "bad"}
	c.Normalize()
	if c.DwellMs != 2000 || c.KneeMode != "arm" || c.KneeLeftRaises != 2 || c.KneeRightDirection != 1 {
		t.Fatalf("normalized %+v", c)
	}
	if len(c.KneeChannels) != 4 || c.KneeChannels[0].Role != "left" || c.KneeChannels[0].ThresholdMM != 75 || c.KneeChannels[1].Role != "right" {
		t.Fatalf("channels %+v", c.KneeChannels)
	}
	if c.DeskSensitivityMg != 350 || c.DeskOrientation != 0 || c.DeskLeft != "tile" || c.DeskRight != "stack" || c.DeskForward != "none" || c.DeskBack != "none" {
		t.Fatalf("desk %+v", c)
	}
}

func TestOldConfigGetsMotionDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(`{"portMode":"auto","dwellMs":500}`), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.DwellMs != 500 || c.KneeMode != "arm" || len(c.KneeChannels) != 4 || c.DeskSensitivityMg != 350 {
		t.Fatalf("old config %+v", c)
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
