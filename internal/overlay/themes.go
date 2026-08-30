package overlay

import (
	"bytes"
	"embed"
	"image"
	"image/jpeg"
	"image/png"
	"strings"
	"sync"
)

//go:embed themes/* slots/* icons/*
var themeFS embed.FS

type Theme struct {
	ID   string
	Name string
	File string
}

func Catalog() []Theme {
	return []Theme{
		{ID: "neon-core", Name: "Neon Core", File: "themes/neon-core.jpg"},
		{ID: "plasma-blue", Name: "Plasma Blue", File: "themes/plasma-blue.jpg"},
		{ID: "crimson-orbit", Name: "Crimson Orbit", File: "themes/crimson-orbit.jpg"},
		{ID: "void-ember", Name: "Void Ember", File: "themes/void-ember.jpg"},
		{ID: "quantum-violet", Name: "Quantum Violet", File: "themes/quantum-violet.jpg"},
		{ID: "solar-flare", Name: "Solar Flare", File: "themes/solar-flare.jpg"},
		{ID: "arctic-ice", Name: "Arctic Ice", File: "themes/arctic-ice.jpg"},
		{ID: "toxic-lime", Name: "Toxic Lime", File: "themes/toxic-lime.jpg"},
		{ID: "cobalt-pulse", Name: "Cobalt Pulse", File: "themes/cobalt-pulse.jpg"},
		{ID: "midnight-iris", Name: "Midnight Iris", File: "themes/midnight-iris.jpg"},
		{ID: "matrix-green", Name: "Matrix Green", File: "themes/matrix-green.jpg"},
		{ID: "rose-circuit", Name: "Rose Circuit", File: "themes/rose-circuit.jpg"},
		{ID: "ion-teal", Name: "Ion Teal", File: "themes/ion-teal.jpg"},
		{ID: "nebula-drift", Name: "Nebula Drift", File: "themes/nebula-drift.jpg"},
		{ID: "amber-nexus", Name: "Amber Nexus", File: "themes/amber-nexus.jpg"},
		{ID: "aqua-sentinel", Name: "Aqua Sentinel", File: "themes/aqua-sentinel.jpg"},
		{ID: "ember-forge", Name: "Ember Forge", File: "themes/ember-forge.jpg"},
		{ID: "tungsten-halo", Name: "Tungsten Halo", File: "themes/tungsten-halo.jpg"},
		{ID: "ultraviolet-gate", Name: "Ultraviolet Gate", File: "themes/ultraviolet-gate.jpg"},
		{ID: "carbon-blade", Name: "Carbon Blade", File: "themes/carbon-blade.jpg"},
	}
}

func NormalizeTheme(id string) string {
	for _, t := range Catalog() {
		if t.ID == id {
			return id
		}
	}
	return Catalog()[0].ID
}

func ThemeIndex(id string) int {
	id = NormalizeTheme(id)
	for i, t := range Catalog() {
		if t.ID == id {
			return i
		}
	}
	return 0
}

func ThemeByID(id string) Theme {
	id = NormalizeTheme(id)
	for _, t := range Catalog() {
		if t.ID == id {
			return t
		}
	}
	return Catalog()[0]
}

var (
	themeOnce sync.Once
	themeImgs map[string]image.Image
)

func ThemeImage(id string) image.Image {
	themeOnce.Do(func() {
		themeImgs = map[string]image.Image{}
		for _, t := range Catalog() {
			if im := EmbeddedImage(t.File); im != nil {
				themeImgs[t.ID] = im
			}
		}
	})
	if im, ok := themeImgs[NormalizeTheme(id)]; ok {
		return im
	}
	for _, im := range themeImgs {
		return im
	}
	return nil
}

var embedImgs sync.Map

func EmbeddedImage(path string) image.Image {
	if v, ok := embedImgs.Load(path); ok {
		if im, ok := v.(image.Image); ok {
			return im
		}
	}
	b, err := themeFS.ReadFile(path)
	if err != nil {
		return nil
	}
	im, err := decodeTheme(b)
	if err != nil {
		return nil
	}
	embedImgs.Store(path, im)
	return im
}

func SlotFile(selected bool) string {
	if selected {
		return "slots/selected.jpg"
	}
	return "slots/idle.jpg"
}

func IconFile(brand, title string) string {
	if strings.EqualFold(strings.TrimSpace(title), "Settings") {
		return "icons/settings.jpg"
	}
	switch strings.ToLower(strings.TrimSpace(brand)) {
	case "cmd":
		return "icons/cmd.jpg"
	case "powershell":
		return "icons/powershell.jpg"
	case "claude":
		return "icons/claude.jpg"
	case "grok":
		return "icons/grok.jpg"
	case "antigravity":
		return "icons/antigravity.jpg"
	case "opencode":
		return "icons/opencode.jpg"
	case "codex":
		return "icons/codex.jpg"
	case "":
		return "icons/empty.jpg"
	default:
		return "icons/unknown.jpg"
	}
}

func decodeTheme(b []byte) (image.Image, error) {
	if im, err := jpeg.Decode(bytes.NewReader(b)); err == nil {
		return im, nil
	}
	return png.Decode(bytes.NewReader(b))
}
