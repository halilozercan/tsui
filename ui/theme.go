package ui

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
)

// Theme is the full set of semantic color tokens used by the UI.
// Every color rendered by tsui reads from CurrentTheme.
type Theme struct {
	// Brand / chrome colors.
	Primary   lipgloss.Color // Logo, main menu selection background.
	Secondary lipgloss.Color // Submenu selection background, accent text, active toggle.

	// Semantic state colors.
	Success lipgloss.Color // Online, "Yes", Connected button, success messages.
	Danger  lipgloss.Color // Offline, "No", Stopped button, errors, danger variant.
	Warning lipgloss.Color // NeedsLogin button, locked-out warning, read-only notice, update notice.
	Info    lipgloss.Color // Starting/Loading button, tips, links, "other" setting values.

	// Foreground color to use when drawing text on top of each background color above.
	FgOnPrimary   lipgloss.Color
	FgOnSecondary lipgloss.Color
	FgOnSuccess   lipgloss.Color
	FgOnDanger    lipgloss.Color
	FgOnWarning   lipgloss.Color
	FgOnInfo      lipgloss.Color

	// Dim surface color used for the main menu item selection when a submenu is open,
	// and for the submenu overflow ("...") indicator background.
	Muted lipgloss.Color
}

// CurrentTheme is read by all rendering code. Overwrite it at startup via ApplyTheme.
var CurrentTheme = DefaultTheme

// CurrentThemeName is the name of the theme currently applied to CurrentTheme.
// Empty before the first ApplyTheme call.
var CurrentThemeName = "default"

// DefaultTheme matches tsui's original color palette.
var DefaultTheme = Theme{
	Primary:   lipgloss.Color("207"),
	Secondary: lipgloss.Color("135"),

	Success: lipgloss.Color("040"),
	Danger:  lipgloss.Color("203"),
	Warning: lipgloss.Color("214"),
	Info:    lipgloss.Color("039"),

	FgOnPrimary:   lipgloss.Color("016"),
	FgOnSecondary: lipgloss.Color("016"),
	FgOnSuccess:   lipgloss.Color("016"),
	FgOnDanger:    lipgloss.Color("016"),
	FgOnWarning:   lipgloss.Color("016"),
	FgOnInfo:      lipgloss.Color("231"),

	Muted: lipgloss.Color("237"),
}

// TokyoNightTheme is a Tokyo Night inspired theme.
// Palette roughly from https://github.com/folke/tokyonight.nvim
var TokyoNightTheme = Theme{
	Primary:   lipgloss.Color("#bb9af7"), // purple
	Secondary: lipgloss.Color("#7aa2f7"), // blue

	Success: lipgloss.Color("#9ece6a"), // green
	Danger:  lipgloss.Color("#f7768e"), // red/pink
	Warning: lipgloss.Color("#e0af68"), // orange/yellow
	Info:    lipgloss.Color("#7dcfff"), // cyan

	FgOnPrimary:   lipgloss.Color("#1a1b26"),
	FgOnSecondary: lipgloss.Color("#1a1b26"),
	FgOnSuccess:   lipgloss.Color("#1a1b26"),
	FgOnDanger:    lipgloss.Color("#1a1b26"),
	FgOnWarning:   lipgloss.Color("#1a1b26"),
	FgOnInfo:      lipgloss.Color("#1a1b26"),

	Muted: lipgloss.Color("#414868"),
}

// DraculaTheme is the Dracula palette (https://draculatheme.com).
var DraculaTheme = Theme{
	Primary:   lipgloss.Color("#bd93f9"), // purple
	Secondary: lipgloss.Color("#ff79c6"), // pink

	Success: lipgloss.Color("#50fa7b"), // green
	Danger:  lipgloss.Color("#ff5555"), // red
	Warning: lipgloss.Color("#f1fa8c"), // yellow
	Info:    lipgloss.Color("#8be9fd"), // cyan

	FgOnPrimary:   lipgloss.Color("#282a36"),
	FgOnSecondary: lipgloss.Color("#282a36"),
	FgOnSuccess:   lipgloss.Color("#282a36"),
	FgOnDanger:    lipgloss.Color("#282a36"),
	FgOnWarning:   lipgloss.Color("#282a36"),
	FgOnInfo:      lipgloss.Color("#282a36"),

	Muted: lipgloss.Color("#44475a"),
}

// GruvboxDarkTheme is the Gruvbox Dark palette
// (https://github.com/morhetz/gruvbox).
var GruvboxDarkTheme = Theme{
	Primary:   lipgloss.Color("#d3869b"), // purple
	Secondary: lipgloss.Color("#fe8019"), // orange

	Success: lipgloss.Color("#b8bb26"), // green
	Danger:  lipgloss.Color("#fb4934"), // red
	Warning: lipgloss.Color("#fabd2f"), // yellow
	Info:    lipgloss.Color("#83a598"), // blue

	FgOnPrimary:   lipgloss.Color("#282828"),
	FgOnSecondary: lipgloss.Color("#282828"),
	FgOnSuccess:   lipgloss.Color("#282828"),
	FgOnDanger:    lipgloss.Color("#282828"),
	FgOnWarning:   lipgloss.Color("#282828"),
	FgOnInfo:      lipgloss.Color("#282828"),

	Muted: lipgloss.Color("#3c3836"),
}

// CatppuccinMochaTheme is the Catppuccin Mocha palette
// (https://github.com/catppuccin/catppuccin).
var CatppuccinMochaTheme = Theme{
	Primary:   lipgloss.Color("#cba6f7"), // mauve
	Secondary: lipgloss.Color("#89b4fa"), // blue

	Success: lipgloss.Color("#a6e3a1"), // green
	Danger:  lipgloss.Color("#f38ba8"), // red
	Warning: lipgloss.Color("#f9e2af"), // yellow
	Info:    lipgloss.Color("#89dceb"), // sky

	FgOnPrimary:   lipgloss.Color("#1e1e2e"),
	FgOnSecondary: lipgloss.Color("#1e1e2e"),
	FgOnSuccess:   lipgloss.Color("#1e1e2e"),
	FgOnDanger:    lipgloss.Color("#1e1e2e"),
	FgOnWarning:   lipgloss.Color("#1e1e2e"),
	FgOnInfo:      lipgloss.Color("#1e1e2e"),

	Muted: lipgloss.Color("#313244"),
}

// BuiltinThemes are the themes shipped with tsui. They're always selectable
// at runtime; user-registered themes (loaded from ~/.config/tsui/themes/)
// shadow these on name collision.
var BuiltinThemes = map[string]Theme{
	"default":          DefaultTheme,
	"tokyo-night":      TokyoNightTheme,
	"dracula":          DraculaTheme,
	"gruvbox-dark":     GruvboxDarkTheme,
	"catppuccin-mocha": CatppuccinMochaTheme,
}

// registeredThemes holds externally-registered themes (loaded from disk).
// Registered themes shadow built-ins with the same name.
var registeredThemes = map[string]Theme{}

// RegisterTheme adds a theme under the given name, making it selectable by
// ApplyTheme and visible in ThemeNames. Registered themes take precedence
// over built-ins with the same name.
func RegisterTheme(name string, theme Theme) {
	registeredThemes[name] = theme
}

// ThemeNames returns all known theme names (built-in ∪ registered), sorted.
func ThemeNames() []string {
	seen := make(map[string]struct{}, len(BuiltinThemes)+len(registeredThemes))
	for name := range BuiltinThemes {
		seen[name] = struct{}{}
	}
	for name := range registeredThemes {
		seen[name] = struct{}{}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// LookupTheme returns the theme for the given name. User-registered themes
// shadow built-ins on name collision.
func LookupTheme(name string) (Theme, bool) {
	if t, ok := registeredThemes[name]; ok {
		return t, true
	}
	t, ok := BuiltinThemes[name]
	return t, ok
}

// ThemeTokens serializes a Theme to the flat token-name -> color-string map
// format used on disk.
func ThemeTokens(t Theme) map[string]string {
	return map[string]string{
		"primary":         string(t.Primary),
		"secondary":       string(t.Secondary),
		"success":         string(t.Success),
		"danger":          string(t.Danger),
		"warning":         string(t.Warning),
		"info":            string(t.Info),
		"fg_on_primary":   string(t.FgOnPrimary),
		"fg_on_secondary": string(t.FgOnSecondary),
		"fg_on_success":   string(t.FgOnSuccess),
		"fg_on_danger":    string(t.FgOnDanger),
		"fg_on_warning":   string(t.FgOnWarning),
		"fg_on_info":      string(t.FgOnInfo),
		"muted":           string(t.Muted),
	}
}

// ParseThemeTokens applies a map of token-name -> color-string on top of
// DefaultTheme. Unknown tokens return an error. Used when loading external
// theme files so authors only need to specify the tokens they want to change.
func ParseThemeTokens(tokens map[string]string) (Theme, error) {
	t := DefaultTheme
	for key, value := range tokens {
		if err := setThemeToken(&t, key, lipgloss.Color(value)); err != nil {
			return Theme{}, err
		}
	}
	return t, nil
}

// ApplyTheme sets CurrentTheme to the named built-in theme and applies any
// per-token overrides. Overrides are keyed by the lowercase field name (e.g.
// "primary", "fg_on_success"). Values are any string lipgloss.Color accepts
// (256-color index like "207" or hex like "#bb9af7").
func ApplyTheme(name string, overrides map[string]string) error {
	if name == "" {
		name = "default"
	}
	base, ok := LookupTheme(name)
	if !ok {
		return fmt.Errorf("unknown theme %q", name)
	}

	for key, value := range overrides {
		if err := setThemeToken(&base, key, lipgloss.Color(value)); err != nil {
			return err
		}
	}

	CurrentTheme = base
	CurrentThemeName = name
	return nil
}

func setThemeToken(t *Theme, key string, value lipgloss.Color) error {
	switch key {
	case "primary":
		t.Primary = value
	case "secondary":
		t.Secondary = value
	case "success":
		t.Success = value
	case "danger":
		t.Danger = value
	case "warning":
		t.Warning = value
	case "info":
		t.Info = value
	case "fg_on_primary":
		t.FgOnPrimary = value
	case "fg_on_secondary":
		t.FgOnSecondary = value
	case "fg_on_success":
		t.FgOnSuccess = value
	case "fg_on_danger":
		t.FgOnDanger = value
	case "fg_on_warning":
		t.FgOnWarning = value
	case "fg_on_info":
		t.FgOnInfo = value
	case "muted":
		t.Muted = value
	default:
		return fmt.Errorf("unknown theme token %q", key)
	}
	return nil
}
