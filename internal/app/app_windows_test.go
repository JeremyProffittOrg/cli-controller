package app

import (
	"testing"

	"github.com/JeremyProffittOrg/cli-controller/internal/wins"
	"golang.org/x/sys/windows"
)

func TestDialItemsAppendSettings(t *testing.T) {
	list := []wins.Window{
		{HWND: windows.Handle(11), Brand: wins.BrandCodex, Title: "First"},
		{HWND: windows.Handle(22), Brand: wins.BrandUnknown, Title: "Second"},
	}
	items := DialItems(list)
	if len(items) != 3 {
		t.Fatalf("len %d", len(items))
	}
	if items[0].Brand != "codex" || items[0].Title != "First" {
		t.Fatalf("first %#v", items[0])
	}
	if items[2].Brand != "" || items[2].Title != SettingsItemTitle {
		t.Fatalf("settings %#v", items[2])
	}
}

func TestSettingsSelectionSurvivesRefresh(t *testing.T) {
	old := []wins.Window{
		{HWND: windows.Handle(11)},
		{HWND: windows.Handle(22)},
	}
	next := []wins.Window{{HWND: windows.Handle(22)}}
	if got := selectionAfterRefresh(old, len(old), next); got != len(next) {
		t.Fatalf("settings selection %d", got)
	}
	if got := selectionAfterRefresh(old, 1, next); got != 0 {
		t.Fatalf("window selection %d", got)
	}
}

func TestDialStateIncludesSettings(t *testing.T) {
	list := []wins.Window{{Brand: wins.BrandCodex, Title: "Session"}}
	n, brand, title := stateForSelection(list, len(list))
	if n != 2 || brand != "" || title != SettingsItemTitle {
		t.Fatalf("settings state n=%d brand=%q title=%q", n, brand, title)
	}
	if !isSettingsSelection(len(list), len(list)) {
		t.Fatal("settings selection not recognized")
	}
}
