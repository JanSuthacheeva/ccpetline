package pet

import (
	"fmt"
	"math/rand"
	"strings"
)

var foodEmojis = []string{
	"\U0001F32E", // 🌮
	"\U0001F957", // 🥗
	"\U0001F36A", // 🍪
	"\U0001F37F", // 🍿
	"\U0001F355", // 🍕
	"\U0001F363", // 🍣
	"\U0001F36C", // 🍬
	"\U0001F34E", // 🍎
	"\U0001F953", // 🥓
	"\U0001F96F", // 🥯
}

// RandomFoodEmoji returns a random food emoji.
func RandomFoodEmoji() string {
	return foodEmojis[rand.Intn(len(foodEmojis))]
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
		return base + RandomFoodEmoji()
	case MoodChasing:
		return base + "\U0001F4A8" // 💨
	case MoodDigging:
		return base + "\U0001F573\uFE0F" // 🕳️
	case MoodFetching:
		return base + "\U0001F4E6" // 📦
	case MoodPouncing:
		return base + "\U0001F4A5" // 💥
	case MoodBored:
		return base + "\U0001F4AD" // 💭
	case MoodNapping:
		return base + "\U0001F4A4" // 💤
	case MoodGrooming:
		return base + "\u2728" // ✨
	case MoodWandering:
		return base + "\U0001F440" // 👀
	case MoodSleeping:
		return base + "\U0001F4A4" // 💤
	default:
		return base
	}
}

// barChars maps each bar style to its (filled, empty) character pair.
var barChars = map[BarStyle][2]string{
	BarClassic: {"\u2500", "\u2500"},
	BarBlock:   {"\u2593", "\u2591"},
	BarThin:    {"\u2501", "\u254C"},
	BarDot:     {"\u25CF", "\u25CB"},
}

// FormatSeparator returns a separator line with the pet emoji positioned by context %.
func FormatSeparator(s *State) string {
	width := s.BarWidth
	if width < 20 {
		width = 50
	}

	chars, ok := barChars[s.BarStyle]
	if !ok {
		chars = barChars[BarClassic]
	}
	filled, empty := chars[0], chars[1]

	displayPct := s.ContextPct
	label := "Ctx"
	if s.ContextMode == ContextModeCtxU {
		displayPct = s.ContextPct / 0.8
		if displayPct > 100 {
			displayPct = 100
		}
		label = "Ctx(u)"
	}

	suffix := fmt.Sprintf(" %s: %.1f%%", label, displayPct)
	barWidth := width - len(suffix)
	if barWidth < 2 {
		barWidth = 2
	}

	if s.BarShowPet {
		emoji := SizeEmoji(s.Species, s.Size)
		pos := int(displayPct / 100 * float64(barWidth-1))
		if pos < 0 {
			pos = 0
		}
		if pos > barWidth-1 {
			pos = barWidth - 1
		}
		left := strings.Repeat(filled, pos)
		rightLen := barWidth - 1 - pos
		if rightLen < 0 {
			rightLen = 0
		}
		right := strings.Repeat(empty, rightLen)
		return left + emoji + right + suffix
	}

	filledLen := int(displayPct / 100 * float64(barWidth))
	if filledLen < 0 {
		filledLen = 0
	}
	if filledLen > barWidth {
		filledLen = barWidth
	}
	return strings.Repeat(filled, filledLen) + strings.Repeat(empty, barWidth-filledLen) + suffix
}
