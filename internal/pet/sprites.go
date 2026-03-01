package pet

import "embed"

//go:embed sprites/*.png
var spriteFS embed.FS

// SpritePNG returns the raw PNG bytes for a given size and mood.
func SpritePNG(size Size, mood Mood) []byte {
	name := spriteName(size, mood)
	data, err := spriteFS.ReadFile("sprites/" + name + ".png")
	if err != nil {
		return nil
	}
	return data
}

var sizeNames = [...]string{"tiny", "normal", "chonky", "megachonk", "unit"}
func spriteMood(mood Mood) string {
	switch mood {
	case MoodEating, MoodChasing, MoodDigging, MoodFetching, MoodPouncing:
		return "eating"
	case MoodSleeping:
		return "sleeping"
	default:
		return "bored"
	}
}

func spriteName(size Size, mood Mood) string {
	s := "normal"
	if int(size) < len(sizeNames) {
		s = sizeNames[size]
	}
	return s + "_" + spriteMood(mood)
}
