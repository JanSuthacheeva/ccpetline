package pet

import (
	"fmt"
	"strings"
)

// PowerlineSepStyle names the Nerd Font glyph drawn between powerline
// segments. All glyphs come from the Powerline-extra range (nf-ple), which
// every Nerd Font includes.
type PowerlineSepStyle string

const (
	SepArrow     PowerlineSepStyle = "arrow"     // left hard divider
	SepRound     PowerlineSepStyle = "round"     // right half circle thick
	SepSlant     PowerlineSepStyle = "slant"     // upper left triangle
	SepBackslant PowerlineSepStyle = "backslant" // lower left triangle
	SepFlame     PowerlineSepStyle = "flame"     // flame thick
	SepPixels    PowerlineSepStyle = "pixels"    // pixelated squares big
	SepNone      PowerlineSepStyle = "none"      // no glyph, straight edge
)

// AllPowerlineSepStyles is the ordered list of selectable separator styles.
var AllPowerlineSepStyles = []PowerlineSepStyle{
	SepArrow, SepRound, SepSlant, SepBackslant, SepFlame, SepPixels, SepNone,
}

var powerlineSepGlyphs = map[PowerlineSepStyle]string{
	SepArrow:     "",
	SepRound:     "",
	SepSlant:     "",
	SepBackslant: "",
	SepFlame:     "",
	SepPixels:    "",
	SepNone:      "",
}

// PowerlineSepGlyph returns the glyph for a separator style, falling back to
// the arrow for unknown or empty styles (pre-existing configs).
func PowerlineSepGlyph(s PowerlineSepStyle) string {
	if g, ok := powerlineSepGlyphs[s]; ok {
		return g
	}
	return powerlineSepGlyphs[SepArrow]
}

// PowerlineSepLabel returns a human-readable name for a separator style.
func PowerlineSepLabel(s PowerlineSepStyle) string {
	switch s {
	case SepRound:
		return "Round"
	case SepSlant:
		return "Slant"
	case SepBackslant:
		return "Backslant"
	case SepFlame:
		return "Flame"
	case SepPixels:
		return "Pixels"
	case SepNone:
		return "None"
	default:
		return "Arrow"
	}
}

// powerlineDefaultBg is the background applied to a segment that has no color
// assigned, so every segment still renders as a filled block.
const powerlineDefaultBg uint8 = 238

// RenderPowerlineLine renders segments as filled colored blocks joined by
// powerline separators of the given style. Each segment's configured color is
// used as its background; the foreground auto-contrasts. Literal separators in
// the template are dropped (the glyphs replace them) and empty tokens are
// skipped.
func RenderPowerlineLine(segs []Segment, colors []uint8, data *SegmentData, sepStyle PowerlineSepStyle) string {
	sep := PowerlineSepGlyph(sepStyle)
	type block struct {
		text string
		bg   uint8
	}
	var blocks []block
	for i, seg := range segs {
		if seg.Kind == KindSeparator {
			continue
		}
		var text string
		switch seg.Kind {
		case KindToken:
			text = resolveToken(seg.Value, data)
		case KindCommand:
			text = execCommand(seg.Value)
		}
		if text == "" {
			continue
		}
		var c uint8
		if i < len(colors) {
			c = colors[i]
		}
		if c == 0 {
			c = powerlineDefaultBg
		}
		blocks = append(blocks, block{text: text, bg: c})
	}
	if len(blocks) == 0 {
		return ""
	}

	var b strings.Builder
	for i, blk := range blocks {
		fg := contrastFg(blk.bg)
		// Filled block: " text " on the segment background.
		fmt.Fprintf(&b, "\x1b[38;5;%d;48;5;%dm %s ", fg, blk.bg, blk.text)
		switch {
		case sep == "":
			// No separator: blocks sit flush against each other; the last one
			// just resets. Each block sets its own colors, so no transition
			// escapes are needed.
			if i+1 == len(blocks) {
				b.WriteString("\x1b[0m")
			}
		case i+1 < len(blocks):
			// Transition glyph: its foreground is this block's background and
			// its background is the next block's, so the colors flow together.
			fmt.Fprintf(&b, "\x1b[38;5;%d;48;5;%dm%s", blk.bg, blocks[i+1].bg, sep)
		default:
			// Trailing glyph fades the last block into the default background.
			fmt.Fprintf(&b, "\x1b[0m\x1b[38;5;%dm%s\x1b[0m", blk.bg, sep)
		}
	}
	return b.String()
}

// contrastFg returns a readable foreground color (near-black or near-white) for
// text drawn on the given ANSI-256 background color.
func contrastFg(bg uint8) uint8 {
	r, g, b := ansi256RGB(bg)
	// Rec. 601 luma; threshold picked empirically across the palette.
	lum := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	if lum > 140 {
		return 16 // near-black on light backgrounds
	}
	return 231 // near-white on dark backgrounds
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
