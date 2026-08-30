package overlay

import "testing"

func TestCatalogHasTwentyThemes(t *testing.T) {
	c := Catalog()
	if len(c) != 20 {
		t.Fatalf("catalog %d", len(c))
	}
	seen := map[string]bool{}
	for _, th := range c {
		if th.ID == "" || th.Name == "" || th.File == "" {
			t.Fatalf("empty %#v", th)
		}
		if seen[th.ID] {
			t.Fatalf("dup %s", th.ID)
		}
		seen[th.ID] = true
	}
}

func TestSlotAndIconImagesDecode(t *testing.T) {
	paths := []string{
		SlotFile(false), SlotFile(true),
		IconFile("cmd", ""), IconFile("powershell", ""), IconFile("claude", ""),
		IconFile("grok", ""), IconFile("antigravity", ""), IconFile("opencode", ""),
		IconFile("codex", ""), IconFile("unknown", ""), IconFile("", "Settings"),
		IconFile("", ""),
	}
	for _, p := range paths {
		if EmbeddedImage(p) == nil {
			t.Fatalf("missing %s", p)
		}
	}
	if IconFile("", "Settings") != "icons/settings.jpg" {
		t.Fatal("settings icon")
	}
}

func TestThemeImagesDecode(t *testing.T) {
	for _, th := range Catalog() {
		im := ThemeImage(th.ID)
		if im == nil {
			t.Fatalf("missing %s", th.ID)
		}
		b := im.Bounds()
		if b.Dx() < 256 || b.Dy() < 256 {
			t.Fatalf("%s size %v", th.ID, b)
		}
	}
}

func TestNormalizeTheme(t *testing.T) {
	if got := NormalizeTheme("solar-flare"); got != "solar-flare" {
		t.Fatalf("known %s", got)
	}
	if got := NormalizeTheme(""); got != "neon-core" {
		t.Fatalf("default %s", got)
	}
	if got := NormalizeTheme("nope"); got != "neon-core" {
		t.Fatalf("unknown %s", got)
	}
	if ThemeIndex("carbon-blade") != 19 {
		t.Fatalf("index %d", ThemeIndex("carbon-blade"))
	}
}
