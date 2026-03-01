package pet

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Species     Species     `json:"species"`
	ContextMode ContextMode `json:"context_mode"`
	ShowSnacks  *bool       `json:"show_snacks,omitempty"`
	SingleLine  bool        `json:"single_line,omitempty"`
}

func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude-pet", "config.json")
}

func defaultConfig() *Config {
	t := true
	return &Config{Species: SpeciesGoose, ContextMode: ContextModeCtx, ShowSnacks: &t}
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
	if c.ShowSnacks == nil {
		t := true
		c.ShowSnacks = &t
	}
	return &c
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
		state.ShowSnacks = c.ShowSnacks != nil && *c.ShowSnacks
		state.SingleLine = c.SingleLine
		_ = SaveState(path, state)
	}
}
