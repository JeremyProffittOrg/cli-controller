package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Brands map[string]bool

type KneeChannel struct {
	Role        string `json:"role"`
	ThresholdMM int    `json:"thresholdMm"`
}

type Config struct {
	PortMode           string        `json:"portMode"`
	Port               string        `json:"port"`
	LastSerial         string        `json:"lastSerial"`
	DwellMs            int           `json:"dwellMs"`
	OverlayView        string        `json:"overlayView"`
	OverlayTheme       string        `json:"overlayTheme"`
	DisplayRotation    int           `json:"displayRotation"`
	Brands             Brands        `json:"brands"`
	KneeMode           string        `json:"kneeMode"`
	KneeLeftRaises     int           `json:"kneeLeftRaises"`
	KneeRightDirection int           `json:"kneeRightDirection"`
	KneeChannels       []KneeChannel `json:"kneeChannels"`
	DeskEnabled        bool          `json:"deskEnabled"`
	DeskSensitivityMg  int           `json:"deskSensitivityMg"`
	DeskOrientation    int           `json:"deskOrientation"`
	DeskLeft           string        `json:"deskLeft"`
	DeskRight          string        `json:"deskRight"`
	DeskForward        string        `json:"deskForward"`
	DeskBack           string        `json:"deskBack"`
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
		PortMode:           "auto",
		Port:               "",
		LastSerial:         "",
		DwellMs:            2000,
		OverlayView:        "classic",
		OverlayTheme:       "neon-core",
		DisplayRotation:    0,
		Brands:             DefaultBrands(),
		KneeMode:           "arm",
		KneeLeftRaises:     2,
		KneeRightDirection: 1,
		KneeChannels: []KneeChannel{
			{Role: "left", ThresholdMM: 75},
			{Role: "right", ThresholdMM: 75},
			{Role: "off", ThresholdMM: 75},
			{Role: "off", ThresholdMM: 75},
		},
		DeskEnabled:       false,
		DeskSensitivityMg: 350,
		DeskOrientation:   0,
		DeskLeft:          "tile",
		DeskRight:         "stack",
		DeskForward:       "none",
		DeskBack:          "none",
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
	if !validDwell(c.DwellMs) {
		c.DwellMs = 2000
	}
	if c.OverlayView != "graphical" {
		c.OverlayView = "classic"
	}
	if c.OverlayTheme == "" {
		c.OverlayTheme = "neon-core"
	}
	c.DisplayRotation %= 360
	if c.DisplayRotation < 0 {
		c.DisplayRotation += 360
	}
	if c.KneeMode != "confirm" {
		c.KneeMode = "arm"
	}
	if c.KneeLeftRaises < 1 || c.KneeLeftRaises > 3 {
		c.KneeLeftRaises = 2
	}
	if c.KneeRightDirection != -1 {
		c.KneeRightDirection = 1
	}
	defaults := Default().KneeChannels
	channels := make([]KneeChannel, 4)
	for i := range channels {
		channels[i] = defaults[i]
		if i < len(c.KneeChannels) {
			channels[i] = c.KneeChannels[i]
		}
		if channels[i].Role != "left" && channels[i].Role != "right" && channels[i].Role != "off" {
			channels[i].Role = defaults[i].Role
		}
		if channels[i].ThresholdMM < 10 || channels[i].ThresholdMM > 300 {
			channels[i].ThresholdMM = 75
		}
	}
	c.KneeChannels = channels
	if c.DeskSensitivityMg < 50 || c.DeskSensitivityMg > 2000 {
		c.DeskSensitivityMg = 350
	}
	if c.DeskOrientation != 0 && c.DeskOrientation != 90 && c.DeskOrientation != 180 && c.DeskOrientation != 270 {
		c.DeskOrientation = 0
	}
	c.DeskLeft = normalizeAction(c.DeskLeft, "tile")
	c.DeskRight = normalizeAction(c.DeskRight, "stack")
	c.DeskForward = normalizeAction(c.DeskForward, "none")
	c.DeskBack = normalizeAction(c.DeskBack, "none")
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

func validDwell(ms int) bool {
	for _, allowed := range []int{250, 500, 750, 1000, 1500, 2000} {
		if ms == allowed {
			return true
		}
	}
	return false
}

func normalizeAction(action, fallback string) string {
	if action == "none" || action == "tile" || action == "stack" {
		return action
	}
	return fallback
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
