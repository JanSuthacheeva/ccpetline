package pet

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Species     Species     `json:"species"`
	ContextMode ContextMode `json:"context_mode"`
}

func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude-pet", "config.json")
}

func LoadConfig() *Config {
	path := ConfigPath()
	if path == "" {
		return &Config{Species: SpeciesGoose, ContextMode: ContextModeCtx}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{Species: SpeciesGoose, ContextMode: ContextModeCtx}
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return &Config{Species: SpeciesGoose, ContextMode: ContextModeCtx}
	}
	if c.Species == "" {
		c.Species = SpeciesGoose
	}
	if c.ContextMode == "" {
		c.ContextMode = ContextModeCtx
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
		_ = SaveState(path, state)
	}
}
