// Package config loads the user's tsui configuration from disk.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/neuralinkcorp/tsui/ui"
)

// Config is the on-disk config file shape.
type Config struct {
	// Name of a built-in theme ("default", "tokyo-night"). Empty = default.
	Theme string `json:"theme,omitempty"`
	// Optional per-token overrides applied on top of the selected theme.
	// Keys are lowercase theme field names (e.g. "primary", "fg_on_success").
	ThemeOverrides map[string]string `json:"theme_overrides,omitempty"`
}

// Dir returns the config directory: $XDG_CONFIG_HOME/tsui, or
// ~/.config/tsui if XDG_CONFIG_HOME is unset.
func Dir() (string, error) {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "tsui"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "tsui"), nil
}

// Path returns the config file path.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// ThemesDir returns the user's themes directory ($CONFIG_DIR/themes).
func ThemesDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "themes"), nil
}

// Load reads and parses the config file. Missing file is not an error — an
// empty Config is returned instead.
func Load() (Config, error) {
	var cfg Config

	path, err := Path()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read %s: %w", path, err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", path, err)
	}

	return cfg, nil
}

// Apply applies the loaded config to global UI state (currently: the theme).
func (c Config) Apply() error {
	return ui.ApplyTheme(c.Theme, c.ThemeOverrides)
}

// LoadThemes scans the user's themes directory and registers each *.json
// file as a theme. The theme name is the filename without extension. Each
// file is a JSON object of token-name -> color-string; see ui.Theme for the
// available tokens. User themes shadow built-ins on name collision. Missing
// directory is not an error.
func LoadThemes() error {
	dir, err := ThemesDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}

		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		var tokens map[string]string
		if err := json.Unmarshal(data, &tokens); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		theme, err := ui.ParseThemeTokens(tokens)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		themeName := name[:len(name)-len(".json")]
		ui.RegisterTheme(themeName, theme)
	}

	return nil
}

// Save writes the config back to disk at Path(), creating the directory if
// needed.
func (c Config) Save() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	path, err := Path()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o644)
}

// SetTheme updates the persisted theme selection, preserving any existing
// ThemeOverrides, and applies the change to the running UI.
func SetTheme(name string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.Theme = name
	if err := cfg.Save(); err != nil {
		return err
	}
	return ui.ApplyTheme(name, cfg.ThemeOverrides)
}
