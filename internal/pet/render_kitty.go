package pet

import "fmt"

// FormatPetLineKitty returns the pet status line with an inline sprite.
func FormatPetLineKitty(s *State) string {
	png := SpritePNG(s.Size, s.Mood)
	if png == nil {
		return RenderEmoji(s) + " " + s.Mood.String()
	}
	sprite := KittyInlinePNG(png)
	suffix := fmt.Sprintf(" %s | Joy: %d", MoodLabel(s.Species, s.Mood), s.Happiness)
	if s.Mood >= MoodEating && s.Mood <= MoodPouncing {
		suffix = RandomFoodEmoji() + suffix
	}
	return sprite + suffix
}

// FormatSeparatorKitty returns the separator with a sprite pet positioned by context %.
func FormatSeparatorKitty(s *State) string {
	png := SpritePNG(s.Size, MoodBored)
	if png == nil {
		return FormatSeparator(s)
	}
	width := s.BarWidth
	if width < 20 {
		width = 50
	}
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
	pos := int(displayPct / 100 * float64(barWidth-1))
	if pos < 0 {
		pos = 0
	}
	if pos > barWidth-1 {
		pos = barWidth - 1
	}
	left := repeatDash(pos)
	rightLen := barWidth - 1 - pos
	if rightLen < 0 {
		rightLen = 0
	}
	right := repeatDash(rightLen)
	return left + KittyInlinePNG(png) + right + suffix
}

func repeatDash(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n*3)
	for i := 0; i < n; i++ {
		b[i*3] = 0xe2
		b[i*3+1] = 0x94
		b[i*3+2] = 0x80
	}
	return string(b)
}
