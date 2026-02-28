package pet

import (
	"math/rand/v2"
	"time"
)

type Mood int

const (
	MoodHappy Mood = iota
	MoodEating
	MoodIdle
	MoodBored
	MoodSleeping
)

func (m Mood) String() string {
	switch m {
	case MoodHappy:
		return "happy"
	case MoodEating:
		return "eating"
	case MoodIdle:
		return "idle"
	case MoodBored:
		return "bored"
	case MoodSleeping:
		return "sleeping"
	default:
		return "unknown"
	}
}

type Size int

const (
	SizeTiny Size = iota
	SizeNormal
	SizeChonky
	SizeMegaChonk
	SizeAbsoluteUnit
)

func (s Size) String() string {
	switch s {
	case SizeTiny:
		return "tiny"
	case SizeNormal:
		return "normal"
	case SizeChonky:
		return "chonky"
	case SizeMegaChonk:
		return "mega chonk"
	case SizeAbsoluteUnit:
		return "ABSOLUTE UNIT"
	default:
		return "unknown"
	}
}

func SizeFromContext(pct float64) Size {
	switch {
	case pct <= 20:
		return SizeTiny
	case pct <= 45:
		return SizeNormal
	case pct <= 70:
		return SizeChonky
	case pct <= 90:
		return SizeMegaChonk
	default:
		return SizeAbsoluteUnit
	}
}

type State struct {
	Mood       Mood
	Size       Size
	ContextPct float64
	Snacks     int
	LastSnack  string
	PosX       int
	Frame      int
	LastEvent  time.Time
	MaxX       int // render width for wandering
}

func NewState() *State {
	return &State{
		Mood:      MoodSleeping,
		Size:      SizeTiny,
		LastEvent: time.Now(),
		MaxX:      60,
	}
}

// Feed processes a snack event.
func (s *State) Feed(toolName string) {
	s.Snacks++
	s.LastSnack = SnackFlavor(toolName)
	s.Mood = MoodEating
	s.LastEvent = time.Now()
}

// SetContext updates the context usage percentage and recalculates size.
func (s *State) SetContext(pct float64) {
	s.ContextPct = pct
	s.Size = SizeFromContext(pct)
	s.LastEvent = time.Now()
}

// Wake transitions from sleeping to happy.
func (s *State) Wake() {
	s.Mood = MoodHappy
	s.LastEvent = time.Now()
}

// Sleep transitions to sleeping mood.
func (s *State) Sleep() {
	s.Mood = MoodSleeping
	s.LastEvent = time.Now()
}

// Tick advances the animation state. Called every 500ms.
func (s *State) Tick() {
	s.Frame++
	elapsed := time.Since(s.LastEvent)

	if s.Mood == MoodSleeping {
		return
	}

	// Eating lasts ~2 seconds (4 ticks)
	if s.Mood == MoodEating && elapsed > 2*time.Second {
		s.Mood = MoodHappy
	}

	// Idle after 10s
	if s.Mood == MoodHappy && elapsed > 10*time.Second {
		s.Mood = MoodIdle
	}

	// Bored after 30s
	if s.Mood == MoodIdle && elapsed > 30*time.Second {
		s.Mood = MoodBored
	}

	// Wander when bored
	if s.Mood == MoodBored || s.Mood == MoodIdle {
		if s.Frame%4 == 0 {
			dir := rand.IntN(3) - 1 // -1, 0, or 1
			s.PosX += dir
			if s.PosX < 0 {
				s.PosX = 0
			}
			if s.PosX > s.MaxX {
				s.PosX = s.MaxX
			}
		}
	}
}

// SnackFlavor maps a tool name to a fun snack name.
func SnackFlavor(tool string) string {
	switch tool {
	case "Bash":
		return "spicy taco"
	case "Read":
		return "mild salad"
	case "Edit", "Write":
		return "crunchy cookie"
	case "Grep", "Glob":
		return "popcorn"
	case "Agent":
		return "whole pizza"
	case "WebFetch", "WebSearch":
		return "sushi roll"
	default:
		return "mystery snack"
	}
}
