package pet

import "testing"

func TestBuildSegmentDataRateLimits(t *testing.T) {
	s := &State{Species: SpeciesGoose, Size: SizeNormal}

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
					"five_hour": map[string]any{"used_percentage": 9.0, "resets_at": 1781180400.0},
					"seven_day": map[string]any{"used_percentage": 41.2, "resets_at": 1781593200.0},
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
