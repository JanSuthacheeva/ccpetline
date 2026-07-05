package pet

import "testing"

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"0.0.8", "0.0.7", 1},
		{"0.0.7", "0.0.8", -1},
		{"0.0.7", "0.0.7", 0},
		{"1.0.0", "0.9.9", 1},
		{"0.10.0", "0.9.0", 1},
		{"0.0.8-rc1", "0.0.7", 1},
		{"0.0.8-rc1", "0.0.8", 0},
		{"0.0.8+build5", "0.0.8", 0},
		{"1.0", "1.0.0", 0},
	}
	for _, tt := range tests {
		if got := compareSemver(tt.a, tt.b); got != tt.want {
			t.Errorf("compareSemver(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
