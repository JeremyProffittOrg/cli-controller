package overlay

import (
	"bytes"
	"embed"
	"image"
	"image/jpeg"
	"image/png"
	"sync"
)

//go:embed themes/*
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
			b, err := themeFS.ReadFile(t.File)
			if err != nil {
				continue
			}
			im, err := decodeTheme(b)
			if err != nil {
				continue
			}
			themeImgs[t.ID] = im
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

func decodeTheme(b []byte) (image.Image, error) {
	if im, err := jpeg.Decode(bytes.NewReader(b)); err == nil {
		return im, nil
	}
	return png.Decode(bytes.NewReader(b))
}
