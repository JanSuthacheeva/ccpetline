package pet

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
var AllTokens = []string{"pet", "mood", "joy", "bar", "model", "ctx", "cost", "changes", "cwd", "branch"}

// SampleSegmentData returns example values for preview rendering.
func SampleSegmentData(species Species, size Size) *SegmentData {
	emoji := SizeEmoji(species, size)
	return &SegmentData{
		Pet:     emoji,
		Mood:    "bored",
		Snacks:  "Joy: 5",
		Bar:     strings.Repeat("\u2500", 3) + emoji + strings.Repeat("\u2500", 12) + " Ctx: 53.1%",
		Model:   "Model: Opus 4",
		Ctx:     "53%",
		Cost:    "$0.42",
		Changes: "(+12/-3)",
		Cwd:     "~/project",
		Branch:  "main",
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

// TokensToTemplate joins token names into a template string like "{a} | {b}".
// Kept for backward compatibility.
func TokensToTemplate(tokens []string) string {
	parts := make([]string, len(tokens))
	for i, t := range tokens {
		parts[i] = "{" + t + "}"
	}
	return strings.Join(parts, " | ")
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
		// literal text before this match — skip whitespace-only gaps
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

// TemplateToTokens parses a template string back into token names.
// Kept for backward compatibility.
func TemplateToTokens(tmpl string) []string {
	matches := tokenRe.FindAllStringSubmatch(tmpl, -1)
	var tokens []string
	for _, m := range matches {
		tokens = append(tokens, m[1])
	}
	return tokens
}

// SegmentData holds all resolved token values for template rendering.
type SegmentData struct {
	Cwd     string
	Branch  string
	Pet     string
	Mood    string
	Changes string
	Model   string
	Ctx     string
	Bar     string
	Snacks  string
	Cost    string
}

// BuildSegmentData resolves all token values from state, Claude JSON, and OS.
func BuildSegmentData(s *State, claudeJSON map[string]any, barWidth int) *SegmentData {
	d := &SegmentData{}

	// {cwd}
	if wd, err := os.Getwd(); err == nil {
		home, _ := os.UserHomeDir()
		if home != "" && strings.HasPrefix(wd, home) {
			wd = "~" + wd[len(home):]
		}
		d.Cwd = wd
	}

	// {branch}
	if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		d.Branch = strings.TrimSpace(string(out))
	}

	// {pet}
	d.Pet = RenderEmoji(s)

	// {mood}
	d.Mood = MoodLabel(s.Species, s.Mood)

	// {snacks}
	d.Snacks = fmt.Sprintf("Joy: %d", s.Happiness)

	// {bar}
	d.Bar = FormatSeparator(s, barWidth)

	// Fields from claudeJSON
	if claudeJSON == nil {
		return d
	}

	// {model}
	if model, ok := claudeJSON["model"].(map[string]any); ok {
		if name, ok := model["display_name"].(string); ok && name != "" {
			d.Model = "Model: " + name
		} else if id, ok := model["id"].(string); ok && id != "" {
			d.Model = "Model: " + id
		}
	}

	// {ctx}
	if cw, ok := claudeJSON["context_window"].(map[string]any); ok {
		if pct, ok := cw["used_percentage"].(float64); ok && pct > 0 {
			d.Ctx = fmt.Sprintf("%.0f%%", pct)
		}
	}

	// {cost}
	if cost, ok := claudeJSON["cost"].(map[string]any); ok {
		if total, ok := cost["total_cost_usd"].(float64); ok && total > 0 {
			d.Cost = fmt.Sprintf("$%.2f", total)
		}
	}

	// {changes} — staged + unstaged git line changes
	if added, removed, err := gitChanges(); err == nil {
		d.Changes = fmt.Sprintf("(+%d/-%d)", added, removed)
	}

	return d
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
		cmd := sub[1]
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		out, err := exec.CommandContext(ctx, "sh", "-c", cmd).Output()
		if err != nil {
			return "<err>"
		}
		return strings.TrimSpace(string(out))
	})

	// Then replace {token} placeholders
	result = tokenRe.ReplaceAllStringFunc(result, func(match string) string {
		key := match[1 : len(match)-1]
		switch key {
		case "cwd":
			return data.Cwd
		case "branch":
			return data.Branch
		case "pet":
			return data.Pet
		case "mood":
			return data.Mood
		case "changes":
			return data.Changes
		case "model":
			return data.Model
		case "ctx":
			return data.Ctx
		case "bar":
			return data.Bar
		case "joy":
			return data.Snacks
		case "cost":
			return data.Cost
		default:
			return match
		}
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
func RenderLines(s *State, claudeJSON map[string]any, barWidth int) []string {
	data := BuildSegmentData(s, claudeJSON, barWidth)

	var lines []string
	for _, tmpl := range s.Lines {
		rendered := RenderTemplate(tmpl, data)
		if rendered != "" {
			lines = append(lines, rendered)
		}
	}

	if len(lines) == 0 {
		lines = []string{FormatSeparator(s, barWidth)}
	}

	return lines
}

// gitChanges returns added/removed line counts from both staged and unstaged changes.
func gitChanges() (added, removed int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	for _, args := range [][]string{
		{"git", "diff", "--shortstat"},
		{"git", "diff", "--cached", "--shortstat"},
	} {
		out, err := exec.CommandContext(ctx, args[0], args[1:]...).Output()
		if err != nil {
			continue
		}
		s := string(out)
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
	return added, removed, nil
}
