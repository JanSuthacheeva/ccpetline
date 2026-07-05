package pet

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Color is a single segment color. Its string form is either an ANSI-256
// index in decimal ("39") or a 24-bit hex code ("#ff8800"); the empty string
// means "no color". In JSON it round-trips as a number for ANSI indices
// (the format of configs written before hex support existed) and as a string
// for hex codes.
type Color string

// ColorNone is the zero Color, rendering text uncolored.
const ColorNone Color = ""

// AnsiColor returns the Color for an ANSI-256 index. Index 0 maps to
// ColorNone, preserving the historical "0 means uncolored" convention.
func AnsiColor(n uint8) Color {
	if n == 0 {
		return ColorNone
	}
	return Color(strconv.Itoa(int(n)))
}

// ParseColor normalizes a user-entered color: empty means none, decimal
// digits are an ANSI-256 index, and "#rgb" / "#rrggbb" are hex codes
// (normalized to lowercase "#rrggbb").
func ParseColor(s string) (Color, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ColorNone, nil
	}
	if s[0] == '#' {
		hex := s[1:]
		if len(hex) == 3 {
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if len(hex) != 6 {
			return ColorNone, fmt.Errorf("hex color must be #rgb or #rrggbb: %q", s)
		}
		if _, err := strconv.ParseUint(hex, 16, 32); err != nil {
			return ColorNone, fmt.Errorf("invalid hex color %q", s)
		}
		return Color("#" + hex), nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n > 255 {
		return ColorNone, fmt.Errorf("ANSI color must be 0-255: %q", s)
	}
	return AnsiColor(uint8(n)), nil
}

// IsNone reports whether the color is the zero "no color" value.
func (c Color) IsNone() bool { return c == "" }

// IsHex reports whether the color is a 24-bit hex code.
func (c Color) IsHex() bool { return strings.HasPrefix(string(c), "#") }

// RGB returns the color's 8-bit RGB components. ANSI indices are approximated
// via the standard 256-color cube; the zero Color returns black.
func (c Color) RGB() (int, int, int) {
	if c.IsNone() {
		return 0, 0, 0
	}
	if c.IsHex() {
		v, _ := strconv.ParseUint(string(c[1:]), 16, 32)
		return int(v >> 16 & 0xff), int(v >> 8 & 0xff), int(v & 0xff)
	}
	n, _ := strconv.Atoi(string(c))
	return ansi256RGB(uint8(n))
}

// fgParams returns the SGR parameters selecting c as the foreground color
// ("38;5;39" or "38;2;255;136;0"), or "" for the zero Color.
func (c Color) fgParams() string { return c.sgrParams(38) }

// bgParams returns the SGR parameters selecting c as the background color
// ("48;5;39" or "48;2;255;136;0"), or "" for the zero Color.
func (c Color) bgParams() string { return c.sgrParams(48) }

func (c Color) sgrParams(base int) string {
	if c.IsNone() {
		return ""
	}
	if c.IsHex() {
		r, g, b := c.RGB()
		return fmt.Sprintf("%d;2;%d;%d;%d", base, r, g, b)
	}
	return fmt.Sprintf("%d;5;%s", base, string(c))
}

// MarshalJSON writes ANSI indices as numbers so existing config files keep
// their shape, and hex codes as strings. The zero Color marshals as 0, the
// historical "no color" value. A Color that is neither (constructed by hand
// from a bad literal) also marshals as 0 rather than emitting invalid JSON.
func (c Color) MarshalJSON() ([]byte, error) {
	if c.IsHex() {
		if _, err := ParseColor(string(c)); err == nil {
			return json.Marshal(string(c))
		}
		return []byte("0"), nil
	}
	if n, err := strconv.Atoi(string(c)); err == nil && n >= 0 && n <= 255 {
		return []byte(strconv.Itoa(n)), nil
	}
	return []byte("0"), nil
}

// UnmarshalJSON accepts a number (ANSI-256 index, the historical format) or a
// string (hex code or decimal index). Invalid values load as ColorNone so one
// bad entry cannot take down the whole config, matching how LoadConfig
// normalizes other fields.
func (c *Color) UnmarshalJSON(data []byte) error {
	var n uint8
	if err := json.Unmarshal(data, &n); err == nil {
		*c = AnsiColor(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		*c = ColorNone
		return nil
	}
	parsed, err := ParseColor(s)
	if err != nil {
		*c = ColorNone
		return nil
	}
	*c = parsed
	return nil
}

// ansi256RGB converts an ANSI-256 color index to approximate 8-bit RGB.
func ansi256RGB(c uint8) (int, int, int) {
	switch {
	case c < 16:
		base := [16][3]int{
			{0, 0, 0}, {128, 0, 0}, {0, 128, 0}, {128, 128, 0},
			{0, 0, 128}, {128, 0, 128}, {0, 128, 128}, {192, 192, 192},
			{128, 128, 128}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
			{0, 0, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
		}
		return base[c][0], base[c][1], base[c][2]
	case c >= 232:
		v := int(c-232)*10 + 8
		return v, v, v
	default:
		i := int(c) - 16
		conv := func(x int) int {
			if x == 0 {
				return 0
			}
			return x*40 + 55
		}
		return conv(i / 36), conv((i % 36) / 6), conv(i % 6)
	}
}
