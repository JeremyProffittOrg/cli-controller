package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Brands map[string]bool

type Config struct {
	PortMode    string `json:"portMode"`
	Port        string `json:"port"`
	LastSerial  string `json:"lastSerial"`
	DwellMs     int    `json:"dwellMs"`
	OverlayView      string `json:"overlayView"`
	DisplayRotation  int    `json:"displayRotation"`
	Brands           Brands `json:"brands"`
}

func BrandNames() []string {
	return []string{"cmd", "powershell", "claude", "grok", "antigravity", "opencode", "codex", "unknown"}
}

func DefaultBrands() Brands {
	b := Brands{}
	for _, n := range BrandNames() {
		b[n] = true
	}
	return b
}

func Default() Config {
	return Config{
		PortMode:    "auto",
		Port:        "",
		LastSerial:  "",
		DwellMs:     2000,
		OverlayView:     "classic",
		DisplayRotation: 0,
		Brands:          DefaultBrands(),
	}
}

func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "cli-controller")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LogPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "cli-controller.log"), nil
}

func (c Config) Enabled(brand string) bool {
	if c.Brands == nil {
		return true
	}
	on, ok := c.Brands[brand]
	if !ok {
		return true
	}
	return on
}

func (c *Config) Normalize() {
	if c.PortMode != "manual" {
		c.PortMode = "auto"
	}
	if c.DwellMs <= 0 {
		c.DwellMs = 2000
	}
	if c.OverlayView != "graphical" {
		c.OverlayView = "classic"
	}
	c.DisplayRotation %= 360
	if c.DisplayRotation < 0 {
		c.DisplayRotation += 360
	}
	if c.Brands == nil {
		c.Brands = DefaultBrands()
		return
	}
	for _, n := range BrandNames() {
		if _, ok := c.Brands[n]; !ok {
			c.Brands[n] = true
		}
	}
}

func Load() (Config, error) {
	cfg := Default()
	p, err := Path()
	if err != nil {
		return cfg, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Default(), err
	}
	cfg.Normalize()
	return cfg, nil
}

func Save(cfg Config) error {
	cfg.Normalize()
	p, err := Path()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, append(b, '\n'), 0o644)
}
