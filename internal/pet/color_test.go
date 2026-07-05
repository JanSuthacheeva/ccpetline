package pet

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseColor(t *testing.T) {
	tests := []struct {
		in      string
		want    Color
		wantErr bool
	}{
		{"", ColorNone, false},
		{"  ", ColorNone, false},
		{"0", ColorNone, false},
		{"39", "39", false},
		{"255", "255", false},
		{"#FF8800", "#ff8800", false},
		{" #ff8800 ", "#ff8800", false},
		{"#abc", "#aabbcc", false},
		{"256", ColorNone, true},
		{"-1", ColorNone, true},
		{"blue", ColorNone, true},
		{"#ff88", ColorNone, true},
		{"#ff880g", ColorNone, true},
	}
	for _, tt := range tests {
		got, err := ParseColor(tt.in)
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Errorf("ParseColor(%q) = %q, %v; want %q, err=%v", tt.in, got, err, tt.want, tt.wantErr)
		}
	}
}

// TestColorJSONRoundTrip guards the config compatibility contract: ANSI
// indices round-trip as numbers (the pre-hex format), hex codes as strings,
// and the historical 0 loads as "no color".
func TestColorJSONRoundTrip(t *testing.T) {
	var colors []Color
	if err := json.Unmarshal([]byte(`[39, 0, "#FF8800", "212", 300, "bogus", {}]`), &colors); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := []Color{"39", ColorNone, "#ff8800", "212", ColorNone, ColorNone, ColorNone}
	for i, w := range want {
		if colors[i] != w {
			t.Errorf("colors[%d] = %q, want %q", i, colors[i], w)
		}
	}

	out, err := json.Marshal([]Color{"39", ColorNone, "#ff8800", Color("garbage")})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := string(out); got != `[39,0,"#ff8800",0]` {
		t.Errorf("marshal = %s", got)
	}
}

func TestColorRGB(t *testing.T) {
	if r, g, b := Color("#ff8800").RGB(); r != 255 || g != 136 || b != 0 {
		t.Errorf("#ff8800 RGB = %d,%d,%d", r, g, b)
	}
	if r, g, b := Color("196").RGB(); r != 255 || g != 0 || b != 0 {
		t.Errorf("196 RGB = %d,%d,%d", r, g, b)
	}
	if r, g, b := ColorNone.RGB(); r != 0 || g != 0 || b != 0 {
		t.Errorf("none RGB = %d,%d,%d", r, g, b)
	}
}

func TestColorSegmentEscapes(t *testing.T) {
	if got := ColorSegment("x", "39"); got != "\x1b[38;5;39mx\x1b[0m" {
		t.Errorf("ansi = %q", got)
	}
	if got := ColorSegment("x", "#ff8800"); got != "\x1b[38;2;255;136;0mx\x1b[0m" {
		t.Errorf("hex = %q", got)
	}
	if got := ColorSegment("x", ColorNone); got != "x" {
		t.Errorf("none = %q", got)
	}
}

// TestRenderPowerlineLineHex checks that hex colors flow through the
// powerline renderer as truecolor backgrounds with a contrasting foreground.
func TestRenderPowerlineLineHex(t *testing.T) {
	segs := TemplateToSegments("{model}")
	data := &SegmentData{IconTheme: IconThemeText, Model: "Opus 4"}
	out := RenderPowerlineLine(segs, []Color{"#ff8800"}, data, SepArrow)
	if !strings.Contains(out, "48;2;255;136;0m Model: Opus 4 ") {
		t.Errorf("hex background missing: %q", out)
	}
	// Bright orange needs a dark foreground.
	if !strings.Contains(out, "38;5;16;48;2;255;136;0m") {
		t.Errorf("contrast foreground missing: %q", out)
	}
}

// TestLoadConfigMixedLineColors loads a config mixing numeric and hex color
// entries, the shape written after a user picks a custom color in the TUI.
func TestLoadConfigMixedLineColors(t *testing.T) {
	var c Config
	data := []byte(`{"species":"cat","separator":" | ","lines":["{cwd} | {branch}"],"line_colors":[[39,0,"#ff8800"]]}`)
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := []Color{"39", ColorNone, "#ff8800"}
	if len(c.LineColors) != 1 || len(c.LineColors[0]) != 3 {
		t.Fatalf("line colors = %v", c.LineColors)
	}
	for i, w := range want {
		if c.LineColors[0][i] != w {
			t.Errorf("line_colors[0][%d] = %q, want %q", i, c.LineColors[0][i], w)
		}
	}
}
