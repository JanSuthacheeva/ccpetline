package pet

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultLines is the default 2-line template layout.
var DefaultLines = []string{
	"{cwd} | {branch} | {changes} | {model}",
	"{ctx_bar} | {pet} {mood}",
}

// DefaultSeparator is the default token separator.
const DefaultSeparator = " | "

type DisplayMode string

const (
	ModeStandalone DisplayMode = ""
	ModePrepend    DisplayMode = "prepend"
	ModeAppend     DisplayMode = "append"
)

var AllDisplayModes = []DisplayMode{ModeStandalone, ModePrepend, ModeAppend}

type BarStyle string

const (
	BarClassic BarStyle = "classic"
	BarBlock   BarStyle = "block"
	BarThin    BarStyle = "thin"
	BarDot     BarStyle = "dot"
)

var AllBarStyles = []BarStyle{BarClassic, BarBlock, BarThin, BarDot}

func BarStyleLabel(s BarStyle) string {
	switch s {
	case BarBlock:
		return "Block"
	case BarThin:
		return "Thin"
	case BarDot:
		return "Dot"
	default:
		return "Classic"
	}
}

func DisplayModeLabel(m DisplayMode) string {
	switch m {
	case ModePrepend:
		return "Prepend"
	case ModeAppend:
		return "Append"
	default:
		return "Standalone"
	}
}

type Config struct {
	Species     Species     `json:"species"`
	ContextMode ContextMode `json:"context_mode"`
	Separator   string      `json:"separator"`
	Lines      []string    `json:"lines,omitempty"`
	LineColors [][]uint8   `json:"line_colors,omitempty"`
	DisplayMode DisplayMode `json:"display_mode,omitempty"`
	WrapCommand string      `json:"wrap_command,omitempty"`
	BarStyle    BarStyle    `json:"bar_style,omitempty"`
	BarShowPet  *bool       `json:"bar_show_pet,omitempty"`
	BarWidth    int         `json:"bar_width,omitempty"`

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
	return filepath.Join(home, ".ccpetline", "config.json")
}

func barShowPetDefault() *bool {
	v := true
	return &v
}

func defaultConfig() *Config {
	return &Config{
		Species:     SpeciesCat,
		ContextMode: ContextModeCtx,
		Separator:   DefaultSeparator,
		Lines:       DefaultLines,
		BarStyle:    BarThin,
		BarShowPet:  barShowPetDefault(),
		BarWidth:    50,
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
		c.Species = SpeciesCat
	}
	if c.ContextMode == "" {
		c.ContextMode = ContextModeCtx
	}
	if c.Separator == "" {
		c.Separator = DefaultSeparator
	}
	switch c.BarStyle {
	case BarClassic, BarBlock, BarThin, BarDot:
	default:
		c.BarStyle = BarThin
	}
	if c.BarShowPet == nil {
		c.BarShowPet = barShowPetDefault()
	}
	if c.BarWidth < 20 || c.BarWidth > 80 {
		c.BarWidth = 50
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
		petLine += " | {joy}"
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

// updateActiveSessions patches all /tmp/ccpetline-state-*.json files
// with the new config values so running sessions pick up changes immediately.
func updateActiveSessions(c *Config) {
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || len(name) < 20 || name[:15] != "ccpetline-state" || name[len(name)-5:] != ".json" {
			continue
		}
		path := filepath.Join(stateDir, name)
		state := LoadState(path)
		state.Species = c.Species
		state.ContextMode = c.ContextMode
		state.Lines = c.Lines
		state.LineColors = c.LineColors
		state.DisplayMode = c.DisplayMode
		state.WrapCommand = c.WrapCommand
		state.BarStyle = c.BarStyle
		if c.BarShowPet != nil {
			state.BarShowPet = *c.BarShowPet
		}
		state.BarWidth = c.BarWidth
		_ = SaveState(path, state)
	}
}
