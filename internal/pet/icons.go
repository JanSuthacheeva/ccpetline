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

// ParseIconTheme normalizes a raw string to a known theme, defaulting to text.
func ParseIconTheme(s string) IconTheme {
	if IconTheme(s) == IconThemeNerd {
		return IconThemeNerd
	}
	return IconThemeText
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

// tokenEmojiFallbacks maps a token key to the emoji shown for it in UI
// pickers when no Nerd Font glyph applies. Kept next to nerdTokenGlyphs so a
// new token needs exactly one edit per theme.
var tokenEmojiFallbacks = map[string]string{
	"pet":     "\U0001F43E",
	"mood":    "\U0001F60A",
	"joy":     "\U0001F496",
	"bar":     "\U0001F4CA",
	"ctx_bar": "\U0001F4CA",
	"model":   "\U0001F916",
	"ctx":     "\U0001F4D0",
	"cost":    "\U0001F4B0",
	"changes": "\U0001F4DD",
	"cwd":     "\U0001F4C2",
	"dir":     "\U0001F4C1",
	"branch":  "\U0001F33F",
	"5h":      "\u23F3",
	"7d":      "\U0001F4C5",
	"5h_bar":  "\U0001F4CA",
	"7d_bar":  "\U0001F4CA",
}

// TokenIcon returns the icon representing a token in UI pickers: the Nerd
// Font glyph in the nerd theme (when one exists), otherwise the emoji
// fallback. The config UI uses this so its per-token icons match the
// rendered status line.
func TokenIcon(theme IconTheme, key string) string {
	if theme == IconThemeNerd {
		if g, ok := nerdTokenGlyphs[key]; ok {
			return g
		}
	}
	return tokenEmojiFallbacks[key]
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
var DefaultTokenColors = map[string]Color{
	"cwd":     "39",  // blue
	"dir":     "39",  // blue
	"branch":  "99",  // purple
	"changes": "214", // gold
	"model":   "51",  // cyan
	"ctx":     "208", // amber
	"ctx_bar": "39",  // blue
	"cost":    "118", // green
	"joy":     "212", // pink
	"mood":    "245", // gray
	"5h":      "208", // amber
	"7d":      "208", // amber
	"5h_bar":  "208", // amber
	"7d_bar":  "208", // amber
}

// DefaultLineColors builds per-segment color arrays for the given line
// templates using DefaultTokenColors. Separators and commands stay uncolored.
// The result aligns positionally with TemplateToSegments, matching how the
// TUI stores per-segment colors.
func DefaultLineColors(lines []string) [][]Color {
	out := make([][]Color, len(lines))
	for i, tmpl := range lines {
		segs := TemplateToSegments(tmpl)
		colors := make([]Color, len(segs))
		for j, seg := range segs {
			if seg.Kind == KindToken {
				colors[j] = DefaultTokenColors[seg.Value]
			}
		}
		out[i] = colors
	}
	return out
}
