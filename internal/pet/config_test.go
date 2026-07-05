package pet

import (
	"os"
	"path/filepath"
	"testing"
)

// writeConfig points HOME at a temp dir and writes raw config JSON there,
// so LoadConfig can be exercised without touching the real user config.
func writeConfig(t *testing.T, raw string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	if raw == "" {
		return
	}
	dir := filepath.Join(home, ".ccpetline")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfigDefaultsWhenMissing(t *testing.T) {
	writeConfig(t, "")
	c := LoadConfig()
	if c.Species != SpeciesCat || c.ContextMode != ContextModeCtx {
		t.Errorf("default species/context wrong: %v %v", c.Species, c.ContextMode)
	}
	if c.BarStyle != BarThin || c.BarWidth != 50 || c.BarShowPet == nil || !*c.BarShowPet {
		t.Errorf("default bar settings wrong: %+v", c)
	}
	if c.Separator != DefaultSeparator || len(c.Lines) == 0 || len(c.LineColors) == 0 {
		t.Errorf("default lines/separator wrong: %+v", c)
	}
}

func TestLoadConfigNormalization(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		check func(t *testing.T, c *Config)
	}{
		{
			name: "empty species and context mode get defaults",
			raw:  `{"species": "", "context_mode": ""}`,
			check: func(t *testing.T, c *Config) {
				if c.Species != SpeciesCat || c.ContextMode != ContextModeCtx {
					t.Errorf("got %v %v", c.Species, c.ContextMode)
				}
			},
		},
		{
			name: "invalid bar style falls back to thin",
			raw:  `{"bar_style": "zigzag"}`,
			check: func(t *testing.T, c *Config) {
				if c.BarStyle != BarThin {
					t.Errorf("got %v", c.BarStyle)
				}
			},
		},
		{
			name: "invalid powerline separator falls back to arrow",
			raw:  `{"powerline_sep": "wavy"}`,
			check: func(t *testing.T, c *Config) {
				if c.PowerlineSep != SepArrow {
					t.Errorf("got %v", c.PowerlineSep)
				}
			},
		},
		{
			name: "bar width out of range resets to 50",
			raw:  `{"bar_width": 500}`,
			check: func(t *testing.T, c *Config) {
				if c.BarWidth != 50 {
					t.Errorf("got %d", c.BarWidth)
				}
			},
		},
		{
			name: "corrupt json falls back to defaults and backs up the file",
			raw:  `{not json`,
			check: func(t *testing.T, c *Config) {
				if c.Species != SpeciesCat || c.BarWidth != 50 {
					t.Errorf("got %+v", c)
				}
				if _, err := os.Stat(ConfigPath() + ".bad"); err != nil {
					t.Errorf("malformed config not preserved as .bad: %v", err)
				}
				if _, err := os.Stat(ConfigPath()); !os.IsNotExist(err) {
					t.Errorf("malformed config still in place: %v", err)
				}
			},
		},
		{
			name: "invalid species and context mode fall back to defaults",
			raw:  `{"species": "doge", "context_mode": "wat"}`,
			check: func(t *testing.T, c *Config) {
				if c.Species != SpeciesCat || c.ContextMode != ContextModeCtx {
					t.Errorf("got %v %v", c.Species, c.ContextMode)
				}
			},
		},
		{
			name: "nerd font inferred from nerd icon theme when key absent",
			raw:  `{"icon_theme": "nerd"}`,
			check: func(t *testing.T, c *Config) {
				if !c.NerdFont || c.IconTheme != IconThemeNerd {
					t.Errorf("nerd font not inferred: %+v", c)
				}
			},
		},
		{
			name: "nerd font inferred from powerline when key absent",
			raw:  `{"powerline": true}`,
			check: func(t *testing.T, c *Config) {
				if !c.NerdFont || !c.Powerline {
					t.Errorf("nerd font not inferred: %+v", c)
				}
			},
		},
		{
			name: "explicit nerd_font false disables glyphs and powerline",
			raw:  `{"nerd_font": false, "icon_theme": "nerd", "powerline": true}`,
			check: func(t *testing.T, c *Config) {
				if c.NerdFont || c.IconTheme != IconThemeText || c.Powerline {
					t.Errorf("nerd font capabilities not disabled: %+v", c)
				}
			},
		},
		{
			name: "missing line colors get defaults for the configured lines",
			raw:  `{"lines": ["{model}", "{bar}"]}`,
			check: func(t *testing.T, c *Config) {
				if len(c.LineColors) != len(c.Lines) {
					t.Errorf("line colors %d for %d lines", len(c.LineColors), len(c.Lines))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeConfig(t, tt.raw)
			tt.check(t, LoadConfig())
		})
	}
}

func TestMigrateConfig(t *testing.T) {
	boolPtr := func(v bool) *bool { return &v }

	tests := []struct {
		name string
		in   Config
		want []string
	}{
		{
			name: "defaults migrate to pet line with joy on top",
			in:   Config{},
			want: []string{"{pet} {mood} | {joy}", "{bar}"},
		},
		{
			name: "single line keeps only the bar",
			in:   Config{SingleLine: true},
			want: []string{"{bar}"},
		},
		{
			name: "pet on bottom",
			in:   Config{PetOnTop: boolPtr(false)},
			want: []string{"{bar}", "{pet} {mood} | {joy}"},
		},
		{
			name: "snacks disabled drops joy",
			in:   Config{ShowSnacks: boolPtr(false)},
			want: []string{"{pet} {mood}", "{bar}"},
		},
		{
			name: "existing lines win over legacy fields",
			in:   Config{Lines: []string{"{model}"}, SingleLine: true, ShowSnacks: boolPtr(false)},
			want: []string{"{model}"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.in
			migrateConfig(&c)
			if len(c.Lines) != len(tt.want) {
				t.Fatalf("lines = %v, want %v", c.Lines, tt.want)
			}
			for i := range tt.want {
				if c.Lines[i] != tt.want[i] {
					t.Fatalf("lines = %v, want %v", c.Lines, tt.want)
				}
			}
			if c.ShowSnacks != nil || c.SingleLine || c.PetOnTop != nil {
				t.Errorf("legacy fields not cleared: %+v", c)
			}
		})
	}
}
