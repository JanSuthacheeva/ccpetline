package pet

import (
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
			want:   "5h: 13% (reset in 2h 14m)",
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
