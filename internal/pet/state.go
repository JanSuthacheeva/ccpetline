package pet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const DefaultStatePath = "/tmp/claude-pet-state.json"

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
	Mood       Mood      `json:"mood"`
	Size       Size      `json:"size"`
	ContextPct float64   `json:"context_pct"`
	Snacks     int       `json:"snacks"`
	LastSnack  string    `json:"last_snack"`
	LastTool   string    `json:"last_tool"`
	LastEvent  time.Time `json:"last_event"`
}

func NewState() *State {
	return &State{
		Mood:      MoodSleeping,
		Size:      SizeTiny,
		LastEvent: time.Now(),
	}
}

// Feed processes a snack event.
func (s *State) Feed(toolName string) {
	s.Snacks++
	s.LastSnack = SnackFlavor(toolName)
	s.LastTool = toolName
	s.Mood = MoodEating
	s.LastEvent = time.Now()
}

// SetContext updates the context usage percentage and recalculates size.
func (s *State) SetContext(pct float64) {
	s.ContextPct = pct
	s.Size = SizeFromContext(pct)
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

// ComputeMood derives the current mood from the LastEvent timestamp.
// This replaces the old Tick() loop — mood is computed on-read.
func (s *State) ComputeMood() {
	if s.Mood == MoodSleeping {
		return
	}
	elapsed := time.Since(s.LastEvent)
	switch {
	case s.Mood == MoodEating && elapsed > 2*time.Second:
		s.Mood = MoodHappy
		fallthrough
	case s.Mood == MoodHappy && elapsed > 10*time.Second:
		s.Mood = MoodIdle
		fallthrough
	case s.Mood == MoodIdle && elapsed > 30*time.Second:
		s.Mood = MoodBored
	}
}

// LoadState reads pet state from a JSON file. Returns a new state if the file doesn't exist.
func LoadState(path string) *State {
	data, err := os.ReadFile(path)
	if err != nil {
		return NewState()
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return NewState()
	}
	return &s
}

// SaveState writes pet state to a JSON file atomically (temp + rename).
func SaveState(path string, s *State) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		// Clean up temp file on rename failure
		os.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}
	// Ensure parent dir exists (for first write)
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	return nil
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
