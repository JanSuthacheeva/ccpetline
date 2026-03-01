package pet

import "fmt"

// FormatPetLineKitty returns the pet status line with an inline sprite.
func FormatPetLineKitty(s *State) string {
	png := SpritePNG(s.Size, s.Mood)
	if png == nil {
		return FormatPetLine(s)
	}
	sprite := KittyInlinePNG(png)
	suffix := fmt.Sprintf(" %s | snacks: %d", s.Mood.String(), s.Snacks)
	if s.Mood == MoodEating {
		suffix = SnackEmoji(s.LastTool) + suffix
	}
	return sprite + suffix
}

// FormatSeparatorKitty returns the separator with a sprite pet positioned by context %.
func FormatSeparatorKitty(s *State, width int) string {
	png := SpritePNG(s.Size, MoodBored)
	if png == nil {
		return FormatSeparator(s, width)
	}
	pos := int(s.ContextPct / 100 * float64(width-1))
	if pos < 0 {
		pos = 0
	}
	if pos > width-1 {
		pos = width - 1
	}
	suffix := fmt.Sprintf(" Ctx: %.1f%%", s.ContextPct)
	left := repeatDash(pos)
	rightLen := width - 1 - pos - len(suffix)
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
