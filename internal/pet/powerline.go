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
const powerlineDefaultBg Color = "238"

// RenderPowerlineLine renders segments as filled colored blocks joined by
// powerline separators of the given style. Each segment's configured color is
// used as its background; the foreground auto-contrasts. Literal separators in
// the template are dropped (the glyphs replace them) and empty tokens are
// skipped.
func RenderPowerlineLine(segs []Segment, colors []Color, data *SegmentData, sepStyle PowerlineSepStyle) string {
	sep := PowerlineSepGlyph(sepStyle)
	type block struct {
		text string
		bg   Color
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
		var c Color
		if i < len(colors) {
			c = colors[i]
		}
		if c.IsNone() {
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
		fmt.Fprintf(&b, "\x1b[%s;%sm %s ", fg.fgParams(), blk.bg.bgParams(), blk.text)
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
			fmt.Fprintf(&b, "\x1b[%s;%sm%s", blk.bg.fgParams(), blocks[i+1].bg.bgParams(), sep)
		default:
			// Trailing glyph fades the last block into the default background.
			fmt.Fprintf(&b, "\x1b[0m\x1b[%sm%s\x1b[0m", blk.bg.fgParams(), sep)
		}
	}
	return b.String()
}

// contrastFg returns a readable foreground color (near-black or near-white) for
// text drawn on the given background color.
func contrastFg(bg Color) Color {
	r, g, b := bg.RGB()
	// Rec. 709 luma; threshold picked empirically across the palette.
	lum := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	if lum > 140 {
		return "16" // near-black on light backgrounds
	}
	return "231" // near-white on dark backgrounds
}
