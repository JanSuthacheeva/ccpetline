package pet

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Species Species `json:"species"`
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
		return &Config{Species: SpeciesGoose}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{Species: SpeciesGoose}
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return &Config{Species: SpeciesGoose}
	}
	if c.Species == "" {
		c.Species = SpeciesGoose
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
	return os.Rename(tmp, path)
}
