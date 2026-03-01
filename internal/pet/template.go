package pet

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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
var AllTokens = []string{"pet", "mood", "snacks", "bar", "model", "ctx", "cost", "changes", "cwd", "branch"}

// SampleSegmentData returns example values for preview rendering.
func SampleSegmentData(species Species, size Size) *SegmentData {
	emoji := SizeEmoji(species, size)
	return &SegmentData{
		Pet:     emoji,
		Mood:    "bored",
		Snacks:  "5",
		Bar:     strings.Repeat("\u2500", 3) + emoji + strings.Repeat("\u2500", 12) + " Ctx: 53.1%",
		Model:   "Opus 4",
		Ctx:     "53%",
		Cost:    "$0.42",
		Changes: "(+12/-3)",
		Cwd:     "~/project",
		Branch:  "main",
	}
}

// SegmentsToTemplate serializes segments into a template string.
func SegmentsToTemplate(segs []Segment) string {
	var b strings.Builder
	for _, seg := range segs {
		switch seg.Kind {
		case KindToken:
			b.WriteString("{" + seg.Value + "}")
		case KindSeparator:
			b.WriteString(seg.Value)
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
		// literal text before this match
		if loc[0] > last {
			segs = append(segs, Segment{Kind: KindSeparator, Value: tmpl[last:loc[0]]})
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
		segs = append(segs, Segment{Kind: KindSeparator, Value: tmpl[last:]})
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
	d.Mood = s.Mood.String()

	// {snacks}
	d.Snacks = fmt.Sprintf("%d", s.Snacks)

	// {bar}
	d.Bar = FormatSeparator(s, barWidth)

	// Fields from claudeJSON
	if claudeJSON == nil {
		return d
	}

	// {model}
	if model, ok := claudeJSON["model"].(map[string]any); ok {
		if name, ok := model["display_name"].(string); ok && name != "" {
			d.Model = name
		} else if id, ok := model["id"].(string); ok && id != "" {
			d.Model = id
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

	// {changes}
	if lines, ok := claudeJSON["lines_changed"].(map[string]any); ok {
		added, _ := lines["added"].(float64)
		removed, _ := lines["removed"].(float64)
		if added > 0 || removed > 0 {
			d.Changes = fmt.Sprintf("(+%.0f/-%.0f)", added, removed)
		}
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
		case "snacks":
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
