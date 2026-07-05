package pet

import (
	"strings"
	"testing"
	"time"
)

func TestFormatRateLimit(t *testing.T) {
	now := time.Unix(1781100000, 0)

	tests := []struct {
		name   string
		window map[string]any
		want   string
	}{
		{
			name:   "with future reset",
			window: map[string]any{"used_percentage": 13.0, "resets_at": float64(1781100000 + 2*3600 + 14*60)},
			want:   "5h: 13% (2h 14m)",
		},
		{
			name:   "reset already passed",
			window: map[string]any{"used_percentage": 9.0, "resets_at": 1781000000.0},
			want:   "5h: 9%",
		},
		{
			name:   "no resets_at",
			window: map[string]any{"used_percentage": 9.0},
			want:   "5h: 9%",
		},
		{
			name:   "no used_percentage",
			window: map[string]any{"resets_at": 1781200000.0},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := map[string]any{"five_hour": tt.window}
			got := formatRateLimit(rl, "five_hour", "5h", now)
			if got != tt.want {
				t.Errorf("formatRateLimit = %q, want %q", got, tt.want)
			}
		})
	}

	if got := formatRateLimit(map[string]any{}, "seven_day", "7d", now); got != "" {
		t.Errorf("formatRateLimit for absent window = %q, want empty", got)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{37 * time.Minute, "37m"},
		{2*time.Hour + 14*time.Minute, "2h 14m"},
		{77*time.Hour + 30*time.Minute, "3d 5h"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.d); got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestBuildSegmentDataRateLimits(t *testing.T) {
	s := &State{Species: SpeciesGoose, Size: SizeNormal}

	// No resets_at in fixtures: keeps expectations independent of time.Now().
	tests := []struct {
		name       string
		claudeJSON map[string]any
		want5h     string
		want7d     string
	}{
		{
			name: "both windows present",
			claudeJSON: map[string]any{
				"rate_limits": map[string]any{
					"five_hour": map[string]any{"used_percentage": 9.0},
					"seven_day": map[string]any{"used_percentage": 41.2},
				},
			},
			want5h: "5h: 9%",
			want7d: "7d: 41%",
		},
		{
			name: "one window absent",
			claudeJSON: map[string]any{
				"rate_limits": map[string]any{
					"five_hour": map[string]any{"used_percentage": 9.0},
				},
			},
			want5h: "5h: 9%",
			want7d: "",
		},
		{
			name:       "rate_limits absent",
			claudeJSON: map[string]any{},
			want5h:     "",
			want7d:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := BuildSegmentData(s, tt.claudeJSON)
			if d.Limit5h != tt.want5h {
				t.Errorf("Limit5h = %q, want %q", d.Limit5h, tt.want5h)
			}
			if d.Limit7d != tt.want7d {
				t.Errorf("Limit7d = %q, want %q", d.Limit7d, tt.want7d)
			}
		})
	}
}

func TestBuildSegmentDataRateLimitBars(t *testing.T) {
	s := &State{Species: SpeciesGoose, Size: SizeNormal, BarStyle: BarBlock, BarWidth: 30}
	claudeJSON := map[string]any{
		"rate_limits": map[string]any{
			"five_hour": map[string]any{"used_percentage": 50.0},
		},
	}

	d := BuildSegmentData(s, claudeJSON)
	want := renderBarLine(50, " 5h: 50%", BarBlock, 30)
	if d.Limit5hBar != want {
		t.Errorf("Limit5hBar = %q, want %q", d.Limit5hBar, want)
	}
	if d.Limit7dBar != "" {
		t.Errorf("Limit7dBar = %q, want empty for absent window", d.Limit7dBar)
	}
}

func TestRenderBarLine(t *testing.T) {
	got := renderBarLine(50, " 5h: 50%", BarBlock, 28)
	// 28 total - 8 suffix bytes = 20 bar chars, half filled.
	want := "▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░ 5h: 50%"
	if got != want {
		t.Errorf("renderBarLine = %q, want %q", got, want)
	}
}

func TestDecorateTokenText(t *testing.T) {
	tests := []struct {
		key, val string
		want     string
	}{
		{"model", "Opus 4", "Model: Opus 4"},
		{"joy", "5", "Joy: 5"},
		{"cost", "0.42", "$0.42"},
		{"changes", "+12/-3", "(+12/-3)"},
		{"ctx", "53%", "53%"},
		{"cwd", "~/p", "~/p"},
		// Branch marker is fixed even in the text theme (U+2387, not U+2325).
		{"branch", "main", "⎇ main"},
		// Empty values stay empty so absent tokens drop cleanly.
		{"model", "", ""},
	}
	for _, tt := range tests {
		// Empty theme must behave as text (back-compat with pre-field configs).
		for _, theme := range []IconTheme{IconThemeText, IconTheme("")} {
			if got := decorateToken(theme, tt.key, tt.val); got != tt.want {
				t.Errorf("decorateToken(%q, %q, %q) = %q, want %q", theme, tt.key, tt.val, got, tt.want)
			}
		}
	}
}

func TestDecorateTokenNerd(t *testing.T) {
	// Expectations derive from the glyph map so they can't drift from it.
	for _, key := range []string{"model", "branch", "joy", "cost", "cwd", "ctx"} {
		want := nerdTokenGlyphs[key] + " val"
		if got := decorateToken(IconThemeNerd, key, "val"); got != want {
			t.Errorf("decorateToken(nerd, %q, val) = %q, want %q", key, got, want)
		}
	}
	// Empty value stays empty regardless of theme.
	if got := decorateToken(IconThemeNerd, "changes", ""); got != "" {
		t.Errorf("empty nerd value = %q, want empty", got)
	}
}

func TestRenderTemplateIconThemes(t *testing.T) {
	tmpl := "{model} | {joy} | {branch}"
	text := &SegmentData{IconTheme: IconThemeText, Model: "Opus 4", Snacks: "5", Branch: "main"}
	if got, want := RenderTemplate(tmpl, text), "Model: Opus 4 | Joy: 5 | ⎇ main"; got != want {
		t.Errorf("text theme = %q, want %q", got, want)
	}
	nerd := &SegmentData{IconTheme: IconThemeNerd, Model: "Opus 4", Snacks: "5", Branch: "main"}
	want := nerdTokenGlyphs["model"] + " Opus 4 | " + nerdTokenGlyphs["joy"] + " 5 | " + nerdTokenGlyphs["branch"] + " main"
	if got := RenderTemplate(tmpl, nerd); got != want {
		t.Errorf("nerd theme = %q, want %q", got, want)
	}
}

func TestDefaultLineColors(t *testing.T) {
	// Colors align positionally with segments; separators stay 0.
	lines := []string{"{cwd} | {branch}", "{pet} {mood}"}
	got := DefaultLineColors(lines)
	// Line 0: [cwd, sep, branch]
	if len(got[0]) != 3 || got[0][0] != DefaultTokenColors["cwd"] || got[0][1] != 0 || got[0][2] != DefaultTokenColors["branch"] {
		t.Errorf("line 0 colors = %v", got[0])
	}
	// Line 1: [pet, mood] - pet uncolored (0), mood gray.
	if len(got[1]) != 2 || got[1][0] != 0 || got[1][1] != DefaultTokenColors["mood"] {
		t.Errorf("line 1 colors = %v", got[1])
	}
}

// TestFormatSeparatorWidth guards the bar-width fix: the rendered line must
// occupy exactly BarWidth display cells regardless of the pet's cell width
// (emoji are 2 cells, Nerd glyphs 1).
func TestFormatSeparatorWidth(t *testing.T) {
	for _, tc := range []struct {
		name  string
		theme IconTheme
		pet   bool
	}{
		{"emoji pet", IconThemeText, true},
		{"nerd pet", IconThemeNerd, true},
		{"no pet", IconThemeText, false},
	} {
		s := &State{
			Species: SpeciesCat, Size: SizeNormal, ContextPct: 53.1,
			BarStyle: BarThin, BarShowPet: tc.pet, BarWidth: 50, IconTheme: tc.theme,
		}
		out := FormatSeparator(s)
		if w := cellWidth(out); w != 50 {
			t.Errorf("%s: width = %d, want 50 (%q)", tc.name, w, out)
		}
	}
}

// TestRenderBarLineWideSuffix guards against byte-vs-cell miscounting: a suffix
// containing a 2-cell emoji must not push the total past the target width.
func TestRenderBarLineWideSuffix(t *testing.T) {
	out := renderBarLine(50, " 🔥 50%", BarBlock, 40)
	if w := cellWidth(out); w != 40 {
		t.Errorf("renderBarLine width = %d, want 40 (%q)", w, out)
	}
}

func TestRenderPowerlineLine(t *testing.T) {
	segs := TemplateToSegments("{model} | {joy}")
	colors := []uint8{51, 0, 212} // model=cyan, sep=0, joy=pink
	data := &SegmentData{IconTheme: IconThemeText, Model: "Opus 4", Snacks: "5"}
	out := RenderPowerlineLine(segs, colors, data, SepArrow)

	// Two arrows (between blocks + trailing), no literal " | ".
	arrow := PowerlineSepGlyph(SepArrow)
	if strings.Count(out, arrow) != 2 {
		t.Errorf("want 2 separators, got %d in %q", strings.Count(out, arrow), out)
	}
	if strings.Contains(out, "|") {
		t.Errorf("powerline line should not contain the plain separator: %q", out)
	}
	// Backgrounds come from the token colors; text is decorated + padded.
	if !strings.Contains(out, "48;5;51m Model: Opus 4 ") {
		t.Errorf("model block missing cyan background: %q", out)
	}
	if !strings.Contains(out, "48;5;212m Joy: 5 ") {
		t.Errorf("joy block missing pink background: %q", out)
	}

	// Empty input renders nothing.
	if RenderPowerlineLine(TemplateToSegments("{cost}"), nil, &SegmentData{}, SepArrow) != "" {
		t.Error("powerline line with only-empty tokens should be empty")
	}
}

func TestRenderPowerlineLineSepStyles(t *testing.T) {
	segs := TemplateToSegments("{model} | {joy}")
	data := &SegmentData{IconTheme: IconThemeText, Model: "Opus 4", Snacks: "5"}

	for _, style := range AllPowerlineSepStyles {
		if style == SepNone {
			continue
		}
		glyph := PowerlineSepGlyph(style)
		out := RenderPowerlineLine(segs, nil, data, style)
		if strings.Count(out, glyph) != 2 {
			t.Errorf("style %q: want 2 %q separators, got %d in %q",
				style, glyph, strings.Count(out, glyph), out)
		}
	}

	// None: blocks sit flush with no glyph between them, one trailing reset.
	out := RenderPowerlineLine(segs, nil, data, SepNone)
	for _, style := range AllPowerlineSepStyles {
		if g := PowerlineSepGlyph(style); g != "" && strings.Contains(out, g) {
			t.Errorf("style none should contain no separator glyph, got %q in %q", g, out)
		}
	}
	if !strings.HasSuffix(out, "\x1b[0m") {
		t.Errorf("style none should end with a reset: %q", out)
	}

	// Unknown or empty styles (old configs) fall back to the arrow.
	arrow := PowerlineSepGlyph(SepArrow)
	if PowerlineSepGlyph("") != arrow {
		t.Error("empty style should fall back to arrow glyph")
	}
	out = RenderPowerlineLine(segs, nil, data, "")
	if strings.Count(out, arrow) != 2 {
		t.Errorf("empty style: want 2 arrow separators in %q", out)
	}
}

func TestContrastFg(t *testing.T) {
	if got := contrastFg(51); got != 16 { // bright cyan -> dark text
		t.Errorf("contrastFg(51) = %d, want 16", got)
	}
	if got := contrastFg(21); got != 231 { // pure blue -> light text
		t.Errorf("contrastFg(21) = %d, want 231", got)
	}
}

func TestRenderTemplateRateLimitTokens(t *testing.T) {
	data := &SegmentData{Limit5h: "5h: 9%", Limit7d: "7d: 41%"}

	got := RenderTemplate("{5h} | {7d}", data)
	want := "5h: 9% | 7d: 41%"
	if got != want {
		t.Errorf("RenderTemplate = %q, want %q", got, want)
	}

	// Absent windows leave no dangling separator.
	got = RenderTemplate("{5h} | {7d}", &SegmentData{})
	if got != "" {
		t.Errorf("RenderTemplate with empty data = %q, want empty", got)
	}
}
