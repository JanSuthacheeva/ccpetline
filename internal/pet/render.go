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
