package pet

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultLines is the default 2-line template layout.
var DefaultLines = []string{
	"{pet} {mood} | snacks: {snacks}",
	"{bar}",
}

// DefaultSeparator is the default token separator.
const DefaultSeparator = " | "

type Config struct {
	Species     Species     `json:"species"`
	ContextMode ContextMode `json:"context_mode"`
	Separator   string      `json:"separator"`
	Lines       []string    `json:"lines,omitempty"`

	// Deprecated fields kept for migration only.
	ShowSnacks *bool `json:"show_snacks,omitempty"`
	SingleLine bool  `json:"single_line,omitempty"`
	PetOnTop   *bool `json:"pet_on_top,omitempty"`
}

func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude-pet", "config.json")
}

func defaultConfig() *Config {
	return &Config{
		Species:     SpeciesGoose,
		ContextMode: ContextModeCtx,
		Separator:   DefaultSeparator,
		Lines:       DefaultLines,
	}
}

func LoadConfig() *Config {
	path := ConfigPath()
	if path == "" {
		return defaultConfig()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig()
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return defaultConfig()
	}
	if c.Species == "" {
		c.Species = SpeciesGoose
	}
	if c.ContextMode == "" {
		c.ContextMode = ContextModeCtx
	}
	if c.Separator == "" {
		c.Separator = DefaultSeparator
	}
	migrateConfig(&c)
	return &c
}

// migrateConfig converts old ShowSnacks/SingleLine/PetOnTop fields to Lines.
func migrateConfig(c *Config) {
	if len(c.Lines) > 0 {
		// Already migrated, clear old fields.
		c.ShowSnacks = nil
		c.SingleLine = false
		c.PetOnTop = nil
		return
	}

	// No Lines set — derive from old fields.
	showSnacks := c.ShowSnacks == nil || *c.ShowSnacks
	petOnTop := c.PetOnTop == nil || *c.PetOnTop

	petLine := "{pet} {mood}"
	if showSnacks {
		petLine += " | snacks: {snacks}"
	}

	if c.SingleLine {
		c.Lines = []string{"{bar}"}
	} else if petOnTop {
		c.Lines = []string{petLine, "{bar}"}
	} else {
		c.Lines = []string{"{bar}", petLine}
	}

	// Clear old fields.
	c.ShowSnacks = nil
	c.SingleLine = false
	c.PetOnTop = nil
}

func SaveConfig(c *Config) error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	updateActiveSessions(c)
	return nil
}

// updateActiveSessions patches all /tmp/claude-pet-state-*.json files
// with the new config values so running sessions pick up changes immediately.
func updateActiveSessions(c *Config) {
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || len(name) < 18 || name[:16] != "claude-pet-state" || name[len(name)-5:] != ".json" {
			continue
		}
		path := filepath.Join(stateDir, name)
		state := LoadState(path)
		state.Species = c.Species
		state.ContextMode = c.ContextMode
		state.Lines = c.Lines
		_ = SaveState(path, state)
	}
}
