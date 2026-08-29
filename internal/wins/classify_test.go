package wins

import "testing"

func TestSurveyFixtures(t *testing.T) {
	cases := []struct {
		proc, title string
		want        Brand
	}{
		{"WindowsTerminal", "Command Prompt - agy", BrandAntigravity},
		{"WindowsTerminal", "? - Thinking - Grok CLI always skip permissions mode - grok", BrandGrok},
		{"WindowsTerminal", `C:\WINDOWS\system32\cmd.exe`, BrandCmd},
		{"WindowsTerminal", "? Current plan", BrandUnknown},
		{"WindowsTerminal", "? ASUS DX10 non-vision model", BrandUnknown},
		{"chrome", "New chat - Claude - Google Chrome", BrandNone},
		{"msedge", "claude", BrandNone},
		{"pwsh", "PowerShell 7.4", BrandPowerShell},
		{"cmd", "Administrator: C:\\Windows\\system32\\cmd.exe", BrandCmd},
		{"claude", "Claude Code", BrandClaude},
		{"grok", "something", BrandGrok},
		{"WindowsTerminal", "opencode --help", BrandOpenCode},
		{"WindowsTerminal", "codex", BrandCodex},
		{"explorer", "File Explorer", BrandNone},
	}
	for _, tc := range cases {
		got := Classify(tc.proc, tc.title)
		if got != tc.want {
			t.Fatalf("%s %q: got %s want %s", tc.proc, tc.title, got, tc.want)
		}
	}
}

func TestChromeExcludedEvenWithClaudeTitle(t *testing.T) {
	if ShouldInclude("chrome", "New chat - Claude - Google Chrome") {
		t.Fatal("chrome must be excluded")
	}
}
