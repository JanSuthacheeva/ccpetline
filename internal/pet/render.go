package pet

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// cellWidth returns the terminal display width of s in columns, correctly
// accounting for wide runes (emoji) and stripping any ANSI. Bar math must use
// this rather than len(), which counts bytes.
func cellWidth(s string) int {
	return lipgloss.Width(s)
}

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

// RenderEmoji returns the pet emoji string based on mood and size. The pet is
// always an emoji, including in the Nerd Font theme, since monochrome animal
// glyphs read poorly at status-bar size.
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

// renderBarLine renders a plain progress bar filled to pct with the suffix
// appended, so that bar plus suffix occupy the given total width.
func renderBarLine(pct float64, suffix string, style BarStyle, width int) string {
	width = clampBarWidth(width)

	chars, ok := barChars[style]
	if !ok {
		chars = barChars[BarClassic]
	}
	filled, empty := chars[0], chars[1]

	barWidth := width - cellWidth(suffix)
	if barWidth < 2 {
		barWidth = 2
	}

	filledLen := int(pct / 100 * float64(barWidth))
	if filledLen < 0 {
		filledLen = 0
	}
	if filledLen > barWidth {
		filledLen = barWidth
	}
	return strings.Repeat(filled, filledLen) + strings.Repeat(empty, barWidth-filledLen) + suffix
}

// FormatSeparator returns a separator line with the pet emoji positioned by context %.
func FormatSeparator(s *State) string {
	width := clampBarWidth(s.BarWidth)

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
	barWidth := width - cellWidth(suffix)
	if barWidth < 2 {
		barWidth = 2
	}

	if s.BarShowPet {
		pet := SizeEmoji(s.Species, s.Size)
		// The pet occupies its own display width on the track (emoji are 2
		// cells wide), so reserve that many columns instead of assuming 1 -
		// otherwise the bar overruns the configured width.
		petW := cellWidth(pet)
		if petW < 1 {
			petW = 1
		}
		track := barWidth - petW
		if track < 0 {
			track = 0
		}
		pos := int(displayPct / 100 * float64(track))
		if pos < 0 {
			pos = 0
		}
		if pos > track {
			pos = track
		}
		left := strings.Repeat(filled, pos)
		rightLen := track - pos
		if rightLen < 0 {
			rightLen = 0
		}
		right := strings.Repeat(empty, rightLen)
		return left + pet + right + suffix
	}

	return renderBarLine(displayPct, suffix, s.BarStyle, width)
}
