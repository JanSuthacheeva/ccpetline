package pet

import (
	"encoding/json"
	"fmt"
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
	Species      Species           `json:"species"`
	ContextMode  ContextMode       `json:"context_mode"`
	NerdFont     bool              `json:"nerd_font,omitempty"`
	IconTheme    IconTheme         `json:"icon_theme,omitempty"`
	Separator    string            `json:"separator"`
	Lines        []string          `json:"lines,omitempty"`
	LineColors   [][]uint8         `json:"line_colors,omitempty"`
	DisplayMode  DisplayMode       `json:"display_mode,omitempty"`
	WrapCommand  string            `json:"wrap_command,omitempty"`
	BarStyle     BarStyle          `json:"bar_style,omitempty"`
	BarShowPet   *bool             `json:"bar_show_pet,omitempty"`
	BarWidth     int               `json:"bar_width,omitempty"`
	Powerline    bool              `json:"powerline,omitempty"`
	PowerlineSep PowerlineSepStyle `json:"powerline_sep,omitempty"`

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

// Bar width bounds. Out-of-range values reset to DefaultBarWidth on load.
const (
	MinBarWidth     = 20
	MaxBarWidth     = 80
	DefaultBarWidth = 50
)

// clampBarWidth normalizes a bar width, resetting out-of-range values to the
// default. LoadConfig and LoadState both apply it, so render code can trust
// the value.
func clampBarWidth(w int) int {
	if w < MinBarWidth || w > MaxBarWidth {
		return DefaultBarWidth
	}
	return w
}

func barShowPetDefault() *bool {
	v := true
	return &v
}

func defaultConfig() *Config {
	return &Config{
		Species:      SpeciesCat,
		ContextMode:  ContextModeCtx,
		IconTheme:    IconThemeText,
		Separator:    DefaultSeparator,
		Lines:        DefaultLines,
		LineColors:   DefaultLineColors(DefaultLines),
		BarStyle:     BarThin,
		BarShowPet:   barShowPetDefault(),
		BarWidth:     DefaultBarWidth,
		PowerlineSep: SepArrow,
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
		// Preserve the malformed file for manual recovery: the next save
		// would otherwise permanently overwrite it with defaults.
		_ = os.Rename(path, path+".bad")
		return defaultConfig()
	}
	c.Species = ParseSpecies(string(c.Species))
	c.ContextMode = ParseContextMode(string(c.ContextMode))
	c.IconTheme = ParseIconTheme(string(c.IconTheme))
	// For configs written before nerd_font existed, infer the capability from
	// usage: any config already using glyphs or the powerline look must have had
	// a Nerd Font. An explicitly present nerd_font is authoritative, so only
	// infer when the key is absent (detected via a nil pointer probe).
	var probe struct {
		NerdFont *bool `json:"nerd_font"`
	}
	_ = json.Unmarshal(data, &probe)
	if probe.NerdFont == nil && (c.IconTheme == IconThemeNerd || c.Powerline) {
		c.NerdFont = true
	}
	// Without a Nerd Font, glyph icons and the powerline look cannot render, so
	// keep the stored config consistent with what the terminal can display.
	if !c.NerdFont {
		c.IconTheme = IconThemeText
		c.Powerline = false
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
	switch c.PowerlineSep {
	case SepRound, SepSlant, SepBackslant, SepFlame, SepPixels, SepNone:
	default:
		c.PowerlineSep = SepArrow
	}
	c.BarWidth = clampBarWidth(c.BarWidth)
	migrateConfig(&c)
	if len(c.LineColors) == 0 {
		c.LineColors = DefaultLineColors(c.Lines)
	}
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
	if path == "" {
		return fmt.Errorf("cannot resolve config path: no home directory")
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(path, data, 0644); err != nil {
		return err
	}
	updateActiveSessions(c)
	return nil
}

// updateActiveSessions patches all session state files with the new config
// values so running sessions pick up changes immediately.
func updateActiveSessions(c *Config) {
	for _, path := range statePaths() {
		state := LoadState(path)
		state.ApplyConfig(c)
		_ = SaveState(path, state)
	}
}
