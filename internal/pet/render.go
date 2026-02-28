package pet

import (
	"fmt"
	"strings"
)

// Render returns the full frame as a string to print.
func Render(s *State, width int) string {
	var b strings.Builder

	art := getArt(s.Size, s.Mood, s.Frame)
	status := statusLine(s)

	// Pad left for wandering position
	pad := strings.Repeat(" ", s.PosX)

	for _, line := range art {
		b.WriteString(pad)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(status)
	b.WriteString("\n")

	// Context bar
	if s.ContextPct > 0 {
		b.WriteString(contextBar(s.ContextPct, width))
		b.WriteString("\n")
	}

	return b.String()
}

func statusLine(s *State) string {
	parts := []string{
		fmt.Sprintf("mood: %s", s.Mood),
		fmt.Sprintf("size: %s", s.Size),
		fmt.Sprintf("snacks: %d", s.Snacks),
	}
	if s.ContextPct > 0 {
		parts = append(parts, fmt.Sprintf("ctx: %.0f%%", s.ContextPct))
	}
	if s.LastSnack != "" {
		parts = append(parts, fmt.Sprintf("last: %s", s.LastSnack))
	}
	return strings.Join(parts, " | ")
}

func contextBar(pct float64, width int) string {
	if width < 10 {
		width = 40
	}
	barWidth := width - 8
	if barWidth < 10 {
		barWidth = 10
	}
	filled := int(pct / 100 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled
	return fmt.Sprintf("[%s%s] %3.0f%%",
		strings.Repeat("#", filled),
		strings.Repeat(".", empty),
		pct)
}

func getArt(size Size, mood Mood, frame int) []string {
	if mood == MoodSleeping {
		return sleepingGoose(size, frame)
	}
	if mood == MoodEating {
		return eatingGoose(size, frame)
	}
	return normalGoose(size, mood, frame)
}

// --- Tiny goose ---
//
//    ,_
//   (o>
//   //\
//   V_/_
//

// --- Normal goose ---
//
//     ,___
//    (o  >
//    |\ \
//    | \_)
//    V  V
//

// --- Chonky goose ---
//
//      ,____
//     (o   >
//     /|   |
//    / |   |
//    |  \_ )
//    VV  VV
//

// --- Mega chonk goose ---
//
//       ,_____
//      (o    >
//      /|    |
//     / |    |
//    /  |    |
//    |   \__ )
//    VV   VV
//

// --- Absolute unit goose ---
//
//        ,______
//       (o     >
//       /|     |
//      / |     |
//     /  |     |
//    /   |     |
//    |    \___ )
//    VV    VV
//

func normalGoose(size Size, mood Mood, frame int) []string {
	eye := "o"
	if mood == MoodBored {
		if frame%6 < 3 {
			eye = "-"
		}
	}
	if mood == MoodIdle {
		eye = "."
	}

	beak := ">"
	if mood == MoodHappy && frame%4 < 2 {
		beak = ")"
	}
	if mood == MoodBored {
		beak = ">"
	}

	// Foot animation for wandering
	feet := func(l, r string) string {
		if (mood == MoodBored || mood == MoodIdle) && frame%4 < 2 {
			return r + "  " + l // swap feet
		}
		return l + "  " + r
	}

	switch size {
	case SizeTiny:
		return []string{
			"  ,_",
			fmt.Sprintf(" (%s%s", eye, beak),
			" //\\",
			fmt.Sprintf(" %s", feet("V", "V")),
		}
	case SizeNormal:
		return []string{
			"   ,___",
			fmt.Sprintf("  (%s  %s", eye, beak),
			"  |\\ \\",
			"  | \\_)",
			fmt.Sprintf("  %s", feet("V", "V")),
		}
	case SizeChonky:
		return []string{
			"    ,____",
			fmt.Sprintf("   (%s   %s", eye, beak),
			"   /|   |",
			"  / |   |",
			"  |  \\_ )",
			fmt.Sprintf("  %s", feet("VV", "VV")),
		}
	case SizeMegaChonk:
		sweat := " "
		if frame%4 < 2 {
			sweat = "'"
		}
		return []string{
			fmt.Sprintf("     ,_____%s", sweat),
			fmt.Sprintf("    (%s    %s", eye, beak),
			"    /|    |",
			"   / |    |",
			"  /  |    |",
			"  |   \\__ )",
			fmt.Sprintf("  %s", feet("VV", " VV")),
		}
	case SizeAbsoluteUnit:
		sweat := "  "
		if frame%3 == 0 {
			sweat = "' "
		} else if frame%3 == 1 {
			sweat = "' '"
		}
		return []string{
			fmt.Sprintf("      ,______%s", sweat),
			fmt.Sprintf("     (%s     %s", eye, beak),
			"     /|      |",
			"    / |      |",
			"   /  |      |",
			"  /   |      |",
			"  |    \\___ )",
			fmt.Sprintf("  %s", feet("VV", "  VV")),
		}
	}
	return []string{"?"}
}

func eatingGoose(size Size, frame int) []string {
	// Chomp animation
	beak := ">"
	snack := "*"
	if frame%2 == 0 {
		beak = ")"
		snack = "~"
	}

	switch size {
	case SizeTiny:
		return []string{
			"  ,_",
			fmt.Sprintf(" (^%s%s", beak, snack),
			" //\\",
			" V  V",
		}
	case SizeNormal:
		return []string{
			"   ,___",
			fmt.Sprintf("  (^  %s%s", beak, snack),
			"  |\\ \\",
			"  | \\_)",
			"  V  V",
		}
	case SizeChonky:
		return []string{
			"    ,____",
			fmt.Sprintf("   (^   %s%s", beak, snack),
			"   /|   |",
			"  / |   |",
			"  |  \\_ )",
			"  VV  VV",
		}
	case SizeMegaChonk:
		return []string{
			"     ,_____",
			fmt.Sprintf("    (^    %s%s", beak, snack),
			"    /|    |",
			"   / |    |",
			"  /  |    |",
			"  |   \\__ )",
			"  VV   VV",
		}
	case SizeAbsoluteUnit:
		return []string{
			"      ,______",
			fmt.Sprintf("     (^     %s%s", beak, snack),
			"     /|      |",
			"    / |      |",
			"   /  |      |",
			"  /   |      |",
			"  |    \\___ )",
			"  VV    VV",
		}
	}
	return []string{"?"}
}

func sleepingGoose(size Size, frame int) []string {
	zzz := ""
	switch frame % 4 {
	case 0:
		zzz = " z"
	case 1:
		zzz = " zz"
	case 2:
		zzz = " zzZ"
	case 3:
		zzz = " zzZZ"
	}

	switch size {
	case SizeTiny:
		return []string{
			"  ,_",
			fmt.Sprintf(" (- %s", zzz),
			" //\\",
			" V  V",
		}
	case SizeNormal:
		return []string{
			"   ,___",
			fmt.Sprintf("  (-  =%s", zzz),
			"  |\\ \\",
			"  | \\_)",
			"  V  V",
		}
	case SizeChonky:
		return []string{
			"    ,____",
			fmt.Sprintf("   (-   =%s", zzz),
			"   /|   |",
			"  / |   |",
			"  |  \\_ )",
			"  VV  VV",
		}
	case SizeMegaChonk:
		return []string{
			"     ,_____",
			fmt.Sprintf("    (-    =%s", zzz),
			"    /|    |",
			"   / |    |",
			"  /  |    |",
			"  |   \\__ )",
			"  VV   VV",
		}
	case SizeAbsoluteUnit:
		return []string{
			"      ,______",
			fmt.Sprintf("     (-     =%s", zzz),
			"     /|      |",
			"    / |      |",
			"   /  |      |",
			"  /   |      |",
			"  |    \\___ )",
			"  VV    VV",
		}
	}
	return []string{"?"}
}
