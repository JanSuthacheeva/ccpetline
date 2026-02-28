package pet

import (
	"fmt"
	"strings"
)

// SnackEmoji returns an emoji for the tool that was just eaten.
func SnackEmoji(toolName string) string {
	switch toolName {
	case "Bash":
		return "\U0001F32E" // taco
	case "Read":
		return "\U0001F957" // salad
	case "Edit", "Write":
		return "\U0001F36A" // cookie
	case "Grep", "Glob":
		return "\U0001F37F" // popcorn
	case "Agent":
		return "\U0001F355" // pizza
	case "WebFetch", "WebSearch":
		return "\U0001F363" // sushi
	default:
		return "\U0001F36C" // candy
	}
}

// SizeEmoji returns the pet emoji based on size.
func SizeEmoji(size Size) string {
	switch size {
	case SizeTiny:
		return "\U0001F423" // hatching chick
	case SizeNormal:
		return "\U0001FABF" // goose
	case SizeChonky:
		return "\U0001FABF\U0001FABF" // double goose
	case SizeMegaChonk:
		return "\U0001F986" // duck
	case SizeAbsoluteUnit:
		return "\U0001F9A2" // swan
	default:
		return "\U0001FABF"
	}
}

// RenderEmoji returns the pet emoji string based on mood and size.
func RenderEmoji(s *State) string {
	base := SizeEmoji(s.Size)
	switch s.Mood {
	case MoodEating:
		return base + SnackEmoji(s.LastTool)
	case MoodBored:
		return base + "\U0001F4AD" // thought bubble
	case MoodSleeping:
		return base + "\U0001F4A4" // zzz
	default:
		return base
	}
}

// FormatStatusLine returns 1-2 lines combining pet info + Claude status info.
func FormatStatusLine(s *State, claudeJSON map[string]any) []string {
	// Line 1: pet status
	parts := []string{
		RenderEmoji(s) + " " + s.Mood.String(),
		fmt.Sprintf("snacks: %d", s.Snacks),
	}
	if s.LastSnack != "" {
		parts = append(parts, fmt.Sprintf("last: %s", s.LastSnack))
	}
	line1 := strings.Join(parts, " | ")

	// Line 2: Claude status info (model, cost, context, lines changed)
	var info []string

	if model, ok := extractModel(claudeJSON); ok {
		info = append(info, model)
	}
if ctx, ok := extractContext(claudeJSON); ok {
		info = append(info, ctx)
	}
	if lines, ok := extractLines(claudeJSON); ok {
		info = append(info, lines)
	}

	if len(info) == 0 {
		return []string{line1}
	}
	line2 := strings.Join(info, " | ")
	return []string{line1, line2}
}

func extractModel(j map[string]any) (string, bool) {
	model, ok := j["model"].(map[string]any)
	if !ok {
		return "", false
	}
	if name, ok := model["display_name"].(string); ok && name != "" {
		return name, true
	}
	if id, ok := model["id"].(string); ok && id != "" {
		return id, true
	}
	return "", false
}

func extractContext(j map[string]any) (string, bool) {
	cw, ok := j["context_window"].(map[string]any)
	if !ok {
		return "", false
	}
	pct, ok := cw["used_percentage"].(float64)
	if !ok || pct == 0 {
		return "", false
	}
	return fmt.Sprintf("ctx: %.0f%%", pct), true
}

func extractLines(j map[string]any) (string, bool) {
	lines, ok := j["lines_changed"].(map[string]any)
	if !ok {
		return "", false
	}
	added, _ := lines["added"].(float64)
	removed, _ := lines["removed"].(float64)
	if added == 0 && removed == 0 {
		return "", false
	}
	return fmt.Sprintf("+%.0f/-%.0f lines", added, removed), true
}
