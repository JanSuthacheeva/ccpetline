package pet

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SegmentKind identifies the type of a template segment.
type SegmentKind int

const (
	KindToken     SegmentKind = iota // {pet}, {mood}, etc.
	KindSeparator                    // literal text like " | " or " • "
	KindCommand                      // [cmd: echo "${PWD##*/}"]
)

// Segment is a single piece of a status line template.
type Segment struct {
	Kind  SegmentKind
	Value string // token name, separator text, or shell command
}

// AllTokens is the ordered list of available template tokens.
var AllTokens = []string{"pet", "mood", "joy", "ctx_bar", "model", "ctx", "cost", "changes", "cwd", "dir", "branch", "5h", "7d", "5h_bar", "7d_bar"}

// SampleSegmentData returns example values for preview rendering. Values are
// raw (undecorated); the icon theme's labels/glyphs are applied at resolve time
// via SegmentData.IconTheme, so the preview matches real output.
func SampleSegmentData(species Species, size Size, barStyle BarStyle, barShowPet bool, barWidth int, iconTheme IconTheme) *SegmentData {
	// Build a sample bar using the given style settings.
	sampleState := &State{
		Species:    species,
		Size:       size,
		Mood:       MoodBored,
		ContextPct: 53.1,
		BarStyle:   barStyle,
		BarShowPet: barShowPet,
		BarWidth:   barWidth,
		IconTheme:  iconTheme,
	}
	return &SegmentData{
		IconTheme:  iconTheme,
		Pet:        SizeEmoji(species, size),
		Mood:       "bored",
		Snacks:     "5",
		Bar:        RenderContextBar(sampleState),
		Model:      "Opus 4",
		Ctx:        "53%",
		Cost:       "0.42",
		Changes:    "+12/-3",
		Cwd:        "~/project",
		Dir:        "project",
		Branch:     "main",
		Limit5h:    "5h: 24% (2h 14m)",
		Limit7d:    "7d: 41% (3d 5h)",
		Limit5hBar: renderBarLine(24, " 5h: 24%", barStyle, barWidth),
		Limit7dBar: renderBarLine(41, " 7d: 41%", barStyle, barWidth),
	}
}

// SegmentsToTemplate serializes segments into a template string.
// Adjacent non-separator segments get a space between them.
// KindSeparator segments render as the configured separator.
func SegmentsToTemplate(segs []Segment, separator string) string {
	var b strings.Builder
	for i, seg := range segs {
		// Auto-space between adjacent non-separator segments.
		if i > 0 && seg.Kind != KindSeparator && segs[i-1].Kind != KindSeparator {
			b.WriteByte(' ')
		}
		switch seg.Kind {
		case KindToken:
			b.WriteString("{" + seg.Value + "}")
		case KindSeparator:
			b.WriteString(separator)
		case KindCommand:
			b.WriteString("[cmd: " + seg.Value + "]")
		}
	}
	return b.String()
}

var (
	tokenRe = regexp.MustCompile(`\{(\w+)\}`)
	cmdRe   = regexp.MustCompile(`\[cmd:\s*(.+?)\]`)
	segRe   = regexp.MustCompile(`\{(\w+)\}|\[cmd:\s*(.+?)\]`)
)

// TemplateToSegments parses a template string into segments.
func TemplateToSegments(tmpl string) []Segment {
	var segs []Segment
	last := 0
	for _, loc := range segRe.FindAllStringSubmatchIndex(tmpl, -1) {
		// literal text before this match - skip whitespace-only gaps
		// (adjacent tokens get auto-spaced)
		if loc[0] > last {
			lit := tmpl[last:loc[0]]
			if strings.TrimSpace(lit) != "" {
				segs = append(segs, Segment{Kind: KindSeparator, Value: lit})
			}
		}
		full := tmpl[loc[0]:loc[1]]
		if full[0] == '{' {
			// token: group 1
			segs = append(segs, Segment{Kind: KindToken, Value: tmpl[loc[2]:loc[3]]})
		} else {
			// command: group 2
			segs = append(segs, Segment{Kind: KindCommand, Value: tmpl[loc[4]:loc[5]]})
		}
		last = loc[1]
	}
	// trailing literal
	if last < len(tmpl) {
		lit := tmpl[last:]
		if strings.TrimSpace(lit) != "" {
			segs = append(segs, Segment{Kind: KindSeparator, Value: lit})
		}
	}
	return segs
}

// SegmentData holds all resolved token values for template rendering. Scalar
// fields (Branch, Model, Snacks, Cost, Changes, ...) hold RAW values; label
// text and Nerd Font glyphs are applied per IconTheme in resolveToken. Pet and
// Bar are pre-themed at build time.
type SegmentData struct {
	IconTheme  IconTheme
	Cwd        string
	Dir        string
	Branch     string
	Pet        string
	Mood       string
	Changes    string
	Model      string
	Ctx        string
	Bar        string
	Snacks     string
	Cost       string
	Limit5h    string
	Limit7d    string
	Limit5hBar string
	Limit7dBar string
}

// BuildSegmentData resolves all token values from state, Claude input, and OS.
func BuildSegmentData(s *State, in *ClaudeInput) *SegmentData {
	d := &SegmentData{IconTheme: s.IconTheme}

	// {cwd}
	if wd, err := os.Getwd(); err == nil {
		home, _ := os.UserHomeDir()
		if home != "" && strings.HasPrefix(wd, home) {
			wd = "~" + wd[len(home):]
		}
		d.Cwd = wd
		d.Dir = filepath.Base(wd)
	}

	// {branch}
	if out, err := runCommand("git", "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		d.Branch = out
	}

	// {pet}
	d.Pet = RenderEmoji(s)

	// {mood}
	d.Mood = MoodLabel(s.Species, s.Mood)

	// {snacks}
	d.Snacks = strconv.Itoa(s.Happiness)

	// {bar}
	d.Bar = RenderContextBar(s)

	// Fields from the Claude payload
	if in == nil {
		return d
	}

	// {model}
	if in.Model.DisplayName != "" {
		d.Model = in.Model.DisplayName
	} else if in.Model.ID != "" {
		d.Model = in.Model.ID
	}

	// {ctx}
	if pct := in.ContextWindow.UsedPercentage; pct != nil && *pct > 0 {
		d.Ctx = fmt.Sprintf("%.0f%%", *pct)
	}

	// {cost}
	if in.Cost.TotalCostUSD > 0 {
		d.Cost = fmt.Sprintf("%.2f", in.Cost.TotalCostUSD)
	}

	// {5h} and {7d} - subscription rate limit usage.
	// Absent until the first API response of the session; each window
	// may be independently missing, so resolve them separately.
	if in.RateLimits != nil {
		now := time.Now()
		d.Limit5h = formatRateLimit(in.RateLimits, "five_hour", "5h", now)
		d.Limit7d = formatRateLimit(in.RateLimits, "seven_day", "7d", now)
		d.Limit5hBar = formatRateLimitBar(in.RateLimits, "five_hour", "5h", s)
		d.Limit7dBar = formatRateLimitBar(in.RateLimits, "seven_day", "7d", s)
	}

	// {changes} - staged + unstaged git line changes
	if added, removed, err := gitChanges(); err == nil {
		d.Changes = fmt.Sprintf("+%d/-%d", added, removed)
	}

	return d
}

// formatRateLimit renders one rate limit window as
// "<label>: <pct>% (<duration>)", or "" when the window is absent.
// The reset part is omitted when resets_at is missing or already passed.
func formatRateLimit(rateLimits map[string]any, window, label string, now time.Time) string {
	pct, ok := rateLimitPct(rateLimits, window)
	if !ok {
		return ""
	}
	out := fmt.Sprintf("%s: %.0f%%", label, pct)
	if w, ok := rateLimits[window].(map[string]any); ok {
		if resetsAt, ok := w["resets_at"].(float64); ok {
			if remaining := time.Unix(int64(resetsAt), 0).Sub(now); remaining > 0 {
				out += fmt.Sprintf(" (%s)", formatDuration(remaining))
			}
		}
	}
	return out
}

// rateLimitPct extracts used_percentage for one rate limit window, reporting
// whether the window is present.
func rateLimitPct(rateLimits map[string]any, window string) (float64, bool) {
	w, ok := rateLimits[window].(map[string]any)
	if !ok {
		return 0, false
	}
	pct, ok := w["used_percentage"].(float64)
	return pct, ok
}

// formatDuration renders a duration compactly: "37m", "2h 14m", "3d 5h".
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, mins)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}

// formatRateLimitBar renders one rate limit window as a progress bar using
// the configured bar style and width, or "" when the window is absent.
func formatRateLimitBar(rateLimits map[string]any, window, label string, s *State) string {
	pct, ok := rateLimitPct(rateLimits, window)
	if !ok {
		return ""
	}
	suffix := fmt.Sprintf(" %s: %.0f%%", label, pct)
	return renderBarLine(pct, suffix, s.BarStyle, s.BarWidth)
}

// ColorSegment wraps text in ANSI color escape codes: 256-color for palette
// indices, 24-bit truecolor for hex codes. The zero color returns text
// unchanged.
func ColorSegment(text string, color Color) string {
	if color.IsNone() || text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", color.fgParams(), text)
}

// resolveToken resolves a single token name to its display string. Scalar
// tokens are decorated with the icon theme's label/glyph; pet, mood, and the
// bars are returned as-is (already themed or intentionally undecorated).
func resolveToken(key string, data *SegmentData) string {
	switch key {
	case "pet":
		return data.Pet
	case "mood":
		return data.Mood
	case "ctx_bar", "bar": // "bar" kept as alias for configs written before the rename
		return data.Bar
	case "5h_bar":
		return data.Limit5hBar
	case "7d_bar":
		return data.Limit7dBar
	}

	var raw string
	switch key {
	case "cwd":
		raw = data.Cwd
	case "dir":
		raw = data.Dir
	case "branch":
		raw = data.Branch
	case "changes":
		raw = data.Changes
	case "model":
		raw = data.Model
	case "ctx":
		raw = data.Ctx
	case "joy":
		raw = data.Snacks
	case "cost":
		raw = data.Cost
	case "5h":
		raw = data.Limit5h
	case "7d":
		raw = data.Limit7d
	default:
		return "{" + key + "}"
	}
	return decorateToken(data.IconTheme, key, raw)
}

// execCommand runs a shell command with the standard timeout and returns its
// output, or "<err>" so a broken command is visible in the status line.
func execCommand(cmd string) string {
	out, err := runCommand("sh", "-c", cmd)
	if err != nil {
		return "<err>"
	}
	return out
}

// ResolvedSegment is a segment whose value has been resolved to display text.
type ResolvedSegment struct {
	Text  string
	Kind  SegmentKind
	Color Color
}

// ResolveSegments pairs each segment with its color and resolves it to
// display text using resolve.
func ResolveSegments(segs []Segment, colors []Color, resolve func(Segment) string) []ResolvedSegment {
	items := make([]ResolvedSegment, len(segs))
	for i, seg := range segs {
		var color Color
		if i < len(colors) {
			color = colors[i]
		}
		items[i] = ResolvedSegment{Text: resolve(seg), Kind: seg.Kind, Color: color}
	}
	return items
}

// AssembleColoredLine filters empty tokens and dangling separators, auto-spaces
// between non-separator segments, and joins the results, coloring each segment
// with colorize. It is the single assembly pipeline shared by the statusline
// renderer and the config TUI preview so the two cannot drift.
func AssembleColoredLine(items []ResolvedSegment, colorize func(text string, color Color) string) string {
	// Filter empty tokens and dangling separators.
	var filtered []ResolvedSegment
	for _, r := range items {
		if r.Kind != KindSeparator && r.Text == "" {
			continue
		}
		filtered = append(filtered, r)
	}
	// Remove leading/trailing separators and collapse adjacent separators.
	var cleaned []ResolvedSegment
	for i, r := range filtered {
		if r.Kind == KindSeparator {
			if len(cleaned) == 0 {
				continue // leading separator
			}
			if i == len(filtered)-1 {
				continue // trailing separator
			}
			if cleaned[len(cleaned)-1].Kind == KindSeparator {
				continue // consecutive separator
			}
		}
		cleaned = append(cleaned, r)
	}
	// Remove trailing separator that might remain.
	if len(cleaned) > 0 && cleaned[len(cleaned)-1].Kind == KindSeparator {
		cleaned = cleaned[:len(cleaned)-1]
	}

	// Build output with auto-spacing and colors.
	var b strings.Builder
	for i, r := range cleaned {
		if i > 0 && r.Kind != KindSeparator && cleaned[i-1].Kind != KindSeparator {
			b.WriteByte(' ')
		}
		b.WriteString(colorize(r.Text, r.Color))
	}
	return b.String()
}

// RenderColoredLine resolves segments, filters empties and dangling separators,
// auto-spaces between non-separator segments, and applies per-segment colors.
func RenderColoredLine(segs []Segment, colors []Color, data *SegmentData) string {
	items := ResolveSegments(segs, colors, func(seg Segment) string {
		switch seg.Kind {
		case KindToken:
			return resolveToken(seg.Value, data)
		case KindCommand:
			return execCommand(seg.Value)
		default:
			return seg.Value
		}
	})
	return AssembleColoredLine(items, ColorSegment)
}

// RenderTemplate replaces {token} placeholders and [cmd: ...] commands,
// then cleans up dangling separators.
func RenderTemplate(tmpl string, data *SegmentData) string {
	// First replace [cmd: ...] blocks
	result := cmdRe.ReplaceAllStringFunc(tmpl, func(match string) string {
		sub := cmdRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		return execCommand(sub[1])
	})

	// Then replace {token} placeholders
	result = tokenRe.ReplaceAllStringFunc(result, func(match string) string {
		return resolveToken(match[1:len(match)-1], data)
	})

	// Clean dangling separators: " | " at start/end, or doubled " | | "
	result = strings.ReplaceAll(result, " |  | ", " | ")
	result = strings.TrimPrefix(result, " | ")
	result = strings.TrimSuffix(result, " | ")
	result = strings.TrimPrefix(result, "| ")
	result = strings.TrimSuffix(result, " |")
	result = strings.TrimSpace(result)

	return result
}

// RenderLines renders all configured line templates, skipping empty results.
// Falls back to a single {bar} line if everything is empty.
func RenderLines(s *State, in *ClaudeInput) []string {
	data := BuildSegmentData(s, in)

	var lines []string
	for i, tmpl := range s.Lines {
		var colors []Color
		if i < len(s.LineColors) {
			colors = s.LineColors[i]
		}
		var rendered string
		switch {
		case s.Powerline:
			rendered = RenderPowerlineLine(TemplateToSegments(tmpl), colors, data, s.PowerlineSep)
		case len(colors) > 0:
			rendered = RenderColoredLine(TemplateToSegments(tmpl), colors, data)
		default:
			rendered = RenderTemplate(tmpl, data)
		}
		if rendered != "" {
			lines = append(lines, rendered)
		}
	}

	if len(lines) == 0 {
		lines = []string{RenderContextBar(s)}
	}

	return lines
}

// gitChanges returns added/removed line counts from both staged and unstaged changes.
func gitChanges() (added, removed int, err error) {
	cmds := [][]string{
		{"git", "diff", "--shortstat"},
		{"git", "diff", "--cached", "--shortstat"},
	}
	failures := 0
	for _, args := range cmds {
		s, cmdErr := runCommand(args[0], args[1:]...)
		if cmdErr != nil {
			failures++
			err = cmdErr
			continue
		}
		if m := regexp.MustCompile(`(\d+) insertion`).FindStringSubmatch(s); len(m) > 1 {
			if n, err := strconv.Atoi(m[1]); err == nil {
				added += n
			}
		}
		if m := regexp.MustCompile(`(\d+) deletion`).FindStringSubmatch(s); len(m) > 1 {
			if n, err := strconv.Atoi(m[1]); err == nil {
				removed += n
			}
		}
	}
	// Outside a git repository both invocations fail; report that instead
	// of a fake +0/-0.
	if failures == len(cmds) {
		return 0, 0, err
	}
	return added, removed, nil
}
