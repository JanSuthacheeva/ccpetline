package pet

// IconTheme selects how token labels are decorated.
type IconTheme string

const (
	// IconThemeText spells out labels ("Model: ", "Joy: ", ...). Works in any
	// font and is the default, so existing configs render unchanged.
	IconThemeText IconTheme = "text"
	// IconThemeNerd prefixes tokens with Nerd Font glyphs. Requires a Nerd Font
	// installed in the terminal; falls back to tofu boxes otherwise.
	IconThemeNerd IconTheme = "nerd"
)

// AllIconThemes is the ordered list of selectable icon themes.
var AllIconThemes = []IconTheme{IconThemeText, IconThemeNerd}

// ParseIconTheme normalizes a raw string to a known theme, defaulting to text.
func ParseIconTheme(s string) IconTheme {
	if IconTheme(s) == IconThemeNerd {
		return IconThemeNerd
	}
	return IconThemeText
}

// IconThemeLabel returns a human-readable name for an icon theme.
func IconThemeLabel(t IconTheme) string {
	switch t {
	case IconThemeNerd:
		return "Nerd Font"
	default:
		return "Text"
	}
}

// nerdTokenGlyphs maps a token key to its Nerd Font prefix glyph. All are drawn
// from long-standing Font Awesome / Octicon / Powerline ranges to minimize the
// chance of a missing glyph, and are single-cell width. The pet itself stays an
// emoji even in this theme - monochrome animal glyphs read poorly at status-bar
// size.
var nerdTokenGlyphs = map[string]string{
	"cwd":     "", // nf-fa-folder
	"dir":     "", // nf-fa-folder
	"branch":  "", // powerline branch
	"model":   "", // nf-fa-microchip
	"ctx":     "", // nf-fa-tachometer (gauge)
	"cost":    "", // nf-fa-dollar
	"changes": "", // nf-oct-diff
	"5h":      "", // nf-fa-clock
	"7d":      "", // nf-fa-clock
	"joy":     "", // nf-fa-heart
}

// TokenIcon returns the Nerd Font glyph representing a token for UI display in
// the nerd theme, or "" when the theme is text or the token has no glyph. The
// config UI uses this so its per-token icons match the rendered status line.
func TokenIcon(theme IconTheme, key string) string {
	if theme == IconThemeNerd {
		return nerdTokenGlyphs[key]
	}
	return ""
}

// textBranchMarker is the non-Nerd branch prefix. It replaces the historical
// "⌥" (the macOS Option key symbol, which is not a VCS glyph) with
// "⎇" (ALTERNATIVE KEY SYMBOL), a fork shape that renders in common fonts.
const textBranchMarker = "⎇"

// decorateToken wraps a raw token value with the theme-appropriate label or
// glyph. An empty value stays empty so downstream separator cleanup still
// drops absent tokens.
func decorateToken(theme IconTheme, key, val string) string {
	if val == "" {
		return ""
	}
	if theme == IconThemeNerd {
		if g, ok := nerdTokenGlyphs[key]; ok {
			return g + " " + val
		}
		return val
	}
	// Text theme: preserve the historical spelled-out labels.
	switch key {
	case "branch":
		return textBranchMarker + " " + val
	case "model":
		return "Model: " + val
	case "joy":
		return "Joy: " + val
	case "cost":
		return "$" + val
	case "changes":
		return "(" + val + ")"
	default:
		return val
	}
}

// DefaultTokenColors maps a token key to an ANSI 256-color for the default
// color scheme. Tokens absent from the map (e.g. pet) stay uncolored so their
// natural emoji color shows through.
var DefaultTokenColors = map[string]uint8{
	"cwd":     39,  // blue
	"dir":     39,  // blue
	"branch":  99,  // purple
	"changes": 214, // gold
	"model":   51,  // cyan
	"ctx":     208, // amber
	"ctx_bar": 39,  // blue
	"cost":    118, // green
	"joy":     212, // pink
	"mood":    245, // gray
	"5h":      208, // amber
	"7d":      208, // amber
	"5h_bar":  208, // amber
	"7d_bar":  208, // amber
}

// DefaultLineColors builds per-segment color arrays for the given line
// templates using DefaultTokenColors. Separators and commands stay uncolored
// (0). The result aligns positionally with TemplateToSegments, matching how the
// TUI stores per-segment colors.
func DefaultLineColors(lines []string) [][]uint8 {
	out := make([][]uint8, len(lines))
	for i, tmpl := range lines {
		segs := TemplateToSegments(tmpl)
		colors := make([]uint8, len(segs))
		for j, seg := range segs {
			if seg.Kind == KindToken {
				colors[j] = DefaultTokenColors[seg.Value]
			}
		}
		out[i] = colors
	}
	return out
}
