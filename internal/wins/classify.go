package wins

import (
	"regexp"
	"strings"
)

type Brand string

const (
	BrandAntigravity Brand = "antigravity"
	BrandClaude      Brand = "claude"
	BrandGrok        Brand = "grok"
	BrandCodex       Brand = "codex"
	BrandOpenCode    Brand = "opencode"
	BrandPowerShell  Brand = "powershell"
	BrandCmd         Brand = "cmd"
	BrandUnknown     Brand = "unknown"
	BrandNone        Brand = ""
)

var includeProcesses = map[string]bool{
	"windowsterminal": true,
	"cmd":             true,
	"powershell":      true,
	"pwsh":            true,
	"wezterm-gui":     true,
	"wezterm":         true,
	"alacritty":       true,
	"openconsole":     true,
	"claude":          true,
	"grok":            true,
	"codex":           true,
	"opencode":        true,
	"agy":             true,
	"antigravity":     true,
}

var excludeProcesses = map[string]bool{
	"chrome":               true,
	"msedge":               true,
	"firefox":              true,
	"explorer":             true,
	"applicationframehost": true,
	"systemsettings":       true,
	"textinputhost":        true,
	"nvidia overlay":       true,
	"cli-controller":       true,
}

var (
	reAntigravity = regexp.MustCompile(`(?i)(^|[^A-Za-z0-9])(agy|antigravity)([^A-Za-z0-9]|$)`)
	reClaude      = regexp.MustCompile(`(?i)\bclaude\b`)
	reGrok        = regexp.MustCompile(`(?i)\bgrok\b`)
	reCodex       = regexp.MustCompile(`(?i)\bcodex\b`)
	reOpenCode    = regexp.MustCompile(`(?i)\bopencode\b`)
	rePowerShell  = regexp.MustCompile(`(?i)\b(powershell|pwsh)\b`)
	reCmd         = regexp.MustCompile(`(?i)(command prompt|cmd\.exe)`)
)

func NormalizeProcess(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.TrimSuffix(s, ".exe")
	return s
}

func ShouldInclude(process, title string) bool {
	_ = title
	p := NormalizeProcess(process)
	if excludeProcesses[p] {
		return false
	}
	return includeProcesses[p]
}

func Classify(process, title string) Brand {
	if !ShouldInclude(process, title) {
		return BrandNone
	}
	p := NormalizeProcess(process)
	switch p {
	case "agy", "antigravity":
		return BrandAntigravity
	case "claude":
		return BrandClaude
	case "grok":
		return BrandGrok
	case "codex":
		return BrandCodex
	case "opencode":
		return BrandOpenCode
	case "powershell", "pwsh":
		return BrandPowerShell
	case "cmd":
		return BrandCmd
	}
	switch {
	case reAntigravity.MatchString(title):
		return BrandAntigravity
	case reClaude.MatchString(title):
		return BrandClaude
	case reGrok.MatchString(title):
		return BrandGrok
	case reCodex.MatchString(title):
		return BrandCodex
	case reOpenCode.MatchString(title):
		return BrandOpenCode
	case rePowerShell.MatchString(title):
		return BrandPowerShell
	case reCmd.MatchString(title):
		return BrandCmd
	default:
		return BrandUnknown
	}
}
