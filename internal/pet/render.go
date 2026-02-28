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

// FormatPetLine returns the single pet status line.
func FormatPetLine(s *State) string {
	parts := []string{
		RenderEmoji(s) + " " + s.Mood.String(),
		fmt.Sprintf("snacks: %d", s.Snacks),
	}
	if s.LastSnack != "" {
		parts = append(parts, fmt.Sprintf("last: %s", s.LastSnack))
	}
	return strings.Join(parts, " | ")
}

// FormatFallbackStatus returns a basic status line from Claude's JSON
// when ccstatusline is not available.
func FormatFallbackStatus(j map[string]any) string {
	var parts []string

	if model, ok := j["model"].(map[string]any); ok {
		if name, ok := model["display_name"].(string); ok && name != "" {
			parts = append(parts, name)
		} else if id, ok := model["id"].(string); ok && id != "" {
			parts = append(parts, id)
		}
	}
	if cost, ok := j["cost"].(map[string]any); ok {
		if total, ok := cost["total_cost_usd"].(float64); ok && total > 0 {
			parts = append(parts, fmt.Sprintf("$%.2f", total))
		}
	}
	if cw, ok := j["context_window"].(map[string]any); ok {
		if pct, ok := cw["used_percentage"].(float64); ok && pct > 0 {
			parts = append(parts, fmt.Sprintf("ctx: %.0f%%", pct))
		}
	}
	if lines, ok := j["lines_changed"].(map[string]any); ok {
		added, _ := lines["added"].(float64)
		removed, _ := lines["removed"].(float64)
		if added > 0 || removed > 0 {
			parts = append(parts, fmt.Sprintf("+%.0f/-%.0f lines", added, removed))
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " | ")
}
