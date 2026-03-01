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

// speciesEmojis maps each species to its 4-stage emoji progression [tiny, normal, chonky, mega].
var speciesEmojis = map[Species][4]string{
	SpeciesGoose:  {"\U0001F423", "\U0001F425", "\U0001FABF", "\U0001F9A2"},             // 🐣 🐥 🪿 🦢
	SpeciesCat:    {"\U0001F431", "\U0001F408", "\U0001F408\u200D\u2B1B", "\U0001F981"}, // 🐱 🐈 🐈‍⬛ 🦁
	SpeciesOcean:  {"\U0001F990", "\U0001F41F", "\U0001F42C", "\U0001F433"},             // 🦐 🐟 🐬 🐳
	SpeciesDragon: {"\U0001F95A", "\U0001F98E", "\U0001F409", "\U0001F432"},             // 🥚 🦎 🐉 🐲
	SpeciesDino:   {"\U0001F9B4", "\U0001F43E", "\U0001F996", "\U0001F995"},             // 🦴 🐾 🦖 🦕
}

// SizeEmoji returns the pet emoji based on species and size.
func SizeEmoji(species Species, size Size) string {
	emojis, ok := speciesEmojis[species]
	if !ok {
		emojis = speciesEmojis[SpeciesGoose]
	}
	idx := int(size)
	if idx < 0 || idx > 3 {
		idx = 1
	}
	return emojis[idx]
}

// RenderEmoji returns the pet emoji string based on mood and size.
func RenderEmoji(s *State) string {
	base := SizeEmoji(s.Species, s.Size)
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

// FormatSeparator returns a separator line with the pet emoji positioned by context %.
func FormatSeparator(s *State, width int) string {
	emoji := SizeEmoji(s.Species, s.Size)

	displayPct := s.ContextPct
	label := "Ctx"
	if s.ContextMode == ContextModeCtxU {
		displayPct = s.ContextPct / 0.8
		if displayPct > 100 {
			displayPct = 100
		}
		label = "Ctx(u)"
	}

	pos := int(displayPct / 100 * float64(width-1))
	if pos < 0 {
		pos = 0
	}
	if pos > width-1 {
		pos = width - 1
	}
	suffix := fmt.Sprintf(" %s: %.1f%%", label, displayPct)
	left := strings.Repeat("\u2500", pos)
	rightLen := width - 1 - pos - len(suffix)
	if rightLen < 0 {
		rightLen = 0
	}
	right := strings.Repeat("\u2500", rightLen)
	return left + emoji + right + suffix
}
