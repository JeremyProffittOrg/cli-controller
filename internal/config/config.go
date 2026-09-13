package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/JeremyProffittOrg/cli-controller/internal/protocol"
)

type Brands map[string]bool

type KneeChannel struct {
	Role        string `json:"role"`
	ThresholdMM int    `json:"thresholdMm"`
}

type SensorControl struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Role          string `json:"role,omitempty"`
	ThresholdMM   int    `json:"thresholdMm,omitempty"`
	Enabled       bool   `json:"enabled,omitempty"`
	SensitivityMg int    `json:"sensitivityMg,omitempty"`
	Orientation   int    `json:"orientation,omitempty"`
	Left          string `json:"left,omitempty"`
	Right         string `json:"right,omitempty"`
	Forward       string `json:"forward,omitempty"`
	Back          string `json:"back,omitempty"`
}

type Config struct {
	PortMode           string          `json:"portMode"`
	Port               string          `json:"port"`
	LastSerial         string          `json:"lastSerial"`
	DwellMs            int             `json:"dwellMs"`
	OverlayView        string          `json:"overlayView"`
	OverlayTheme       string          `json:"overlayTheme"`
	DisplayRotation    int             `json:"displayRotation"`
	Brands             Brands          `json:"brands"`
	KneeMode           string          `json:"kneeMode"`
	KneeLeftRaises     int             `json:"kneeLeftRaises"`
	KneeRightDirection int             `json:"kneeRightDirection"`
	KneeChannels       []KneeChannel   `json:"kneeChannels"`
	Sensors            []SensorControl `json:"sensors"`
	DeskEnabled        bool            `json:"deskEnabled"`
	DeskSensitivityMg  int             `json:"deskSensitivityMg"`
	DeskOrientation    int             `json:"deskOrientation"`
	DeskLeft           string          `json:"deskLeft"`
	DeskRight          string          `json:"deskRight"`
	DeskForward        string          `json:"deskForward"`
	DeskBack           string          `json:"deskBack"`
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

func defaultTof(ch int, role string) SensorControl {
	return SensorControl{
		ID:          protocol.FormatSensorID(0x70, ch, "tof"),
		Kind:        "tof",
		Role:        role,
		ThresholdMM: 75,
	}
}

func defaultAccel() SensorControl {
	return SensorControl{
		ID:            protocol.FormatSensorID(0x70, 4, "accel"),
		Kind:          "accel",
		Enabled:       false,
		SensitivityMg: 350,
		Orientation:   0,
		Left:          "tile",
		Right:         "stack",
		Forward:       "none",
		Back:          "none",
	}
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
		Sensors: []SensorControl{
			defaultTof(0, "left"),
			defaultTof(1, "right"),
			defaultTof(2, "off"),
			defaultTof(3, "off"),
			defaultAccel(),
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
	c.normalizeSensors()
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

func (c *Config) normalizeSensors() {
	if len(c.Sensors) == 0 {
		c.Sensors = make([]SensorControl, 0, 5)
		for i, ch := range c.KneeChannels {
			c.Sensors = append(c.Sensors, SensorControl{
				ID:          protocol.FormatSensorID(0x70, i, "tof"),
				Kind:        "tof",
				Role:        ch.Role,
				ThresholdMM: ch.ThresholdMM,
			})
		}
		accel := defaultAccel()
		accel.Enabled = c.DeskEnabled
		accel.SensitivityMg = c.DeskSensitivityMg
		accel.Orientation = c.DeskOrientation
		accel.Left = c.DeskLeft
		accel.Right = c.DeskRight
		accel.Forward = c.DeskForward
		accel.Back = c.DeskBack
		c.Sensors = append(c.Sensors, accel)
	}
	for i, ch := range c.KneeChannels {
		id := protocol.FormatSensorID(0x70, i, "tof")
		for j := range c.Sensors {
			if c.Sensors[j].ID == id {
				c.Sensors[j].Kind = "tof"
				c.Sensors[j].Role = ch.Role
				c.Sensors[j].ThresholdMM = ch.ThresholdMM
			}
		}
	}
	deskID := protocol.FormatSensorID(0x70, 4, "accel")
	for j := range c.Sensors {
		if c.Sensors[j].ID == deskID {
			c.Sensors[j].Kind = "accel"
			c.Sensors[j].Enabled = c.DeskEnabled
			c.Sensors[j].SensitivityMg = c.DeskSensitivityMg
			c.Sensors[j].Orientation = c.DeskOrientation
			c.Sensors[j].Left = c.DeskLeft
			c.Sensors[j].Right = c.DeskRight
			c.Sensors[j].Forward = c.DeskForward
			c.Sensors[j].Back = c.DeskBack
		}
	}
	seen := map[string]bool{}
	out := make([]SensorControl, 0, len(c.Sensors))
	for _, s := range c.Sensors {
		s = normalizeSensor(s, *c)
		if s.ID == "" || seen[s.ID] {
			continue
		}
		seen[s.ID] = true
		out = append(out, s)
	}
	c.Sensors = out
}

func normalizeSensor(s SensorControl, cfg Config) SensorControl {
	s.ID = strings.TrimSpace(s.ID)
	s.Kind = strings.ToLower(strings.TrimSpace(s.Kind))
	if s.Kind != "accel" {
		s.Kind = "tof"
	}
	if s.Kind == "tof" {
		if s.Role != "left" && s.Role != "right" && s.Role != "off" {
			s.Role = "off"
		}
		if s.ThresholdMM < 10 || s.ThresholdMM > 300 {
			s.ThresholdMM = 75
		}
		s.Enabled = false
		s.SensitivityMg = 0
		s.Orientation = 0
		s.Left, s.Right, s.Forward, s.Back = "", "", "", ""
		return s
	}
	s.Role = ""
	s.ThresholdMM = 0
	if s.SensitivityMg < 50 || s.SensitivityMg > 2000 {
		s.SensitivityMg = cfg.DeskSensitivityMg
		if s.SensitivityMg < 50 || s.SensitivityMg > 2000 {
			s.SensitivityMg = 350
		}
	}
	if s.Orientation != 0 && s.Orientation != 90 && s.Orientation != 180 && s.Orientation != 270 {
		s.Orientation = cfg.DeskOrientation
	}
	s.Left = normalizeAction(s.Left, cfg.DeskLeft)
	s.Right = normalizeAction(s.Right, cfg.DeskRight)
	s.Forward = normalizeAction(s.Forward, cfg.DeskForward)
	s.Back = normalizeAction(s.Back, cfg.DeskBack)
	return s
}

func (c Config) FindSensor(id string) (SensorControl, bool) {
	for _, s := range c.Sensors {
		if s.ID == id {
			return s, true
		}
	}
	return SensorControl{}, false
}

func (c *Config) UpsertSensor(s SensorControl) {
	s = normalizeSensor(s, *c)
	if s.ID == "" {
		return
	}
	for i := range c.Sensors {
		if c.Sensors[i].ID == s.ID {
			c.Sensors[i] = s
			return
		}
	}
	c.Sensors = append(c.Sensors, s)
}

func (c *Config) RemoveSensor(id string) {
	out := c.Sensors[:0]
	for _, s := range c.Sensors {
		if s.ID != id {
			out = append(out, s)
		}
	}
	c.Sensors = out
}

func NewControlFromStatus(st protocol.SensorStatus, cfg Config) SensorControl {
	s := SensorControl{ID: st.ID, Kind: st.Kind}
	if s.ID == "" {
		s.ID = protocol.FormatSensorID(st.Mux, st.Ch, st.Kind)
	}
	if s.Kind == "accel" {
		s.Enabled = cfg.DeskEnabled
		s.SensitivityMg = cfg.DeskSensitivityMg
		s.Orientation = cfg.DeskOrientation
		s.Left = cfg.DeskLeft
		s.Right = cfg.DeskRight
		s.Forward = cfg.DeskForward
		s.Back = cfg.DeskBack
	} else {
		s.Kind = "tof"
		s.Role = "off"
		s.ThresholdMM = 75
	}
	return normalizeSensor(s, cfg)
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
