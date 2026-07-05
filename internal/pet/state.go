package pet

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var stateDir = os.TempDir()

// StatePath returns the state file path for a given session ID.
// Falls back to a default path if sessionID is empty.
func StatePath(sessionID string) string {
	if sessionID == "" {
		return filepath.Join(stateDir, "ccpetline-state.json")
	}
	return filepath.Join(stateDir, fmt.Sprintf("ccpetline-state-%s.json", sessionID))
}

type Species string

const (
	SpeciesGoose  Species = "goose"
	SpeciesCat    Species = "cat"
	SpeciesOcean  Species = "ocean"
	SpeciesDragon Species = "dragon"
	SpeciesDino   Species = "dino"
)

var AllSpecies = []Species{SpeciesGoose, SpeciesCat, SpeciesOcean, SpeciesDragon, SpeciesDino}

type ContextMode string

const (
	ContextModeCtx  ContextMode = "ctx"
	ContextModeCtxU ContextMode = "ctx_u"
)

var AllContextModes = []ContextMode{ContextModeCtx, ContextModeCtxU}

func ParseContextMode(s string) ContextMode {
	if ContextMode(s) == ContextModeCtxU {
		return ContextModeCtxU
	}
	return ContextModeCtx
}

func ParseSpecies(s string) Species {
	switch Species(s) {
	case SpeciesGoose, SpeciesCat, SpeciesOcean, SpeciesDragon, SpeciesDino:
		return Species(s)
	default:
		return SpeciesGoose
	}
}

type Mood int

const (
	MoodEating Mood = iota
	MoodChasing
	MoodDigging
	MoodFetching
	MoodPouncing
	MoodBored
	MoodNapping
	MoodGrooming
	MoodWandering
	MoodSleeping
)

var ActiveMoods = []Mood{MoodEating, MoodChasing, MoodDigging, MoodFetching, MoodPouncing}
var IdleMoods = []Mood{MoodBored, MoodNapping, MoodGrooming, MoodWandering}

func (m Mood) String() string {
	switch m {
	case MoodEating:
		return "eating"
	case MoodChasing:
		return "chasing"
	case MoodDigging:
		return "digging"
	case MoodFetching:
		return "fetching"
	case MoodPouncing:
		return "pouncing"
	case MoodBored:
		return "bored"
	case MoodNapping:
		return "napping"
	case MoodGrooming:
		return "grooming"
	case MoodWandering:
		return "wandering"
	case MoodSleeping:
		return "sleeping"
	default:
		return "bored"
	}
}

// moodLabels maps (species, mood) to species-flavored display labels.
var moodLabels = map[Species]map[Mood]string{
	SpeciesGoose: {
		MoodEating: "gobbling", MoodChasing: "honk-chasing", MoodDigging: "pecking",
		MoodFetching: "waddling", MoodPouncing: "flapping",
		MoodBored: "bored", MoodNapping: "nesting", MoodGrooming: "preening",
		MoodWandering: "waddling about", MoodSleeping: "sleeping",
	},
	SpeciesCat: {
		MoodEating: "nibbling", MoodChasing: "pouncing", MoodDigging: "scratching",
		MoodFetching: "batting", MoodPouncing: "leaping",
		MoodBored: "bored", MoodNapping: "curling up", MoodGrooming: "grooming",
		MoodWandering: "prowling", MoodSleeping: "sleeping",
	},
	SpeciesOcean: {
		MoodEating: "gulping", MoodChasing: "darting", MoodDigging: "diving",
		MoodFetching: "surfacing", MoodPouncing: "breaching",
		MoodBored: "bored", MoodNapping: "floating", MoodGrooming: "gliding",
		MoodWandering: "exploring", MoodSleeping: "sleeping",
	},
	SpeciesDragon: {
		MoodEating: "devouring", MoodChasing: "swooping", MoodDigging: "burrowing",
		MoodFetching: "hoarding", MoodPouncing: "striking",
		MoodBored: "bored", MoodNapping: "smoldering", MoodGrooming: "polishing scales",
		MoodWandering: "surveying", MoodSleeping: "sleeping",
	},
	SpeciesDino: {
		MoodEating: "chomping", MoodChasing: "stomping", MoodDigging: "excavating",
		MoodFetching: "lumbering", MoodPouncing: "charging",
		MoodBored: "bored", MoodNapping: "dozing", MoodGrooming: "sunbathing",
		MoodWandering: "roaming", MoodSleeping: "sleeping",
	},
}

// MoodLabel returns a species-flavored display label for a mood.
func MoodLabel(species Species, mood Mood) string {
	if labels, ok := moodLabels[species]; ok {
		if label, ok := labels[mood]; ok {
			return label
		}
	}
	return mood.String()
}

type Size int

const (
	SizeTiny Size = iota
	SizeNormal
	SizeChonky
	SizeMegaChonk
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
	default:
		return "unknown"
	}
}

func SizeFromContext(pct float64) Size {
	switch {
	case pct <= 20:
		return SizeTiny
	case pct <= 35:
		return SizeNormal
	case pct <= 60:
		return SizeChonky
	default:
		return SizeMegaChonk
	}
}

type State struct {
	Species        Species           `json:"species"`
	ContextMode    ContextMode       `json:"context_mode"`
	IconTheme      IconTheme         `json:"icon_theme,omitempty"`
	Lines          []string          `json:"lines,omitempty"`
	LineColors     [][]uint8         `json:"line_colors,omitempty"`
	DisplayMode    DisplayMode       `json:"display_mode,omitempty"`
	WrapCommand    string            `json:"wrap_command,omitempty"`
	BarStyle       BarStyle          `json:"bar_style,omitempty"`
	BarShowPet     bool              `json:"bar_show_pet"`
	BarWidth       int               `json:"bar_width,omitempty"`
	Powerline      bool              `json:"powerline,omitempty"`
	PowerlineSep   PowerlineSepStyle `json:"powerline_sep,omitempty"`
	Mood           Mood              `json:"mood"`
	Size           Size              `json:"size"`
	ContextPct     float64           `json:"context_pct"`
	Happiness      int               `json:"happiness"`
	LastEvent      time.Time         `json:"last_event"`
	LastMoodChange time.Time         `json:"last_mood_change"`
}

func NewState() *State {
	s := &State{
		BarShowPet: true,
		Mood:       MoodSleeping,
		Size:       SizeTiny,
		LastEvent:  time.Now(),
	}
	s.ApplyConfig(LoadConfig())
	return s
}

// ApplyConfig copies the display settings snapshot from cfg into s. This is
// the single place where Config fields map onto their State duplicates; new
// settings only need to be added here.
func (s *State) ApplyConfig(cfg *Config) {
	s.Species = cfg.Species
	s.ContextMode = cfg.ContextMode
	s.IconTheme = cfg.IconTheme
	s.Lines = cfg.Lines
	s.LineColors = cfg.LineColors
	s.DisplayMode = cfg.DisplayMode
	s.WrapCommand = cfg.WrapCommand
	s.BarStyle = cfg.BarStyle
	if cfg.BarShowPet != nil {
		s.BarShowPet = *cfg.BarShowPet
	}
	s.BarWidth = cfg.BarWidth
	s.Powerline = cfg.Powerline
	s.PowerlineSep = cfg.PowerlineSep
}

const moodCooldown = 60 * time.Second

// Feed processes a snack event.
func (s *State) Feed(toolName string) {
	s.Happiness++
	s.LastEvent = time.Now()
	if time.Since(s.LastMoodChange) >= moodCooldown {
		s.Mood = ActiveMoods[rand.Intn(len(ActiveMoods))]
		s.LastMoodChange = time.Now()
	}
}

// SetContext updates the context usage percentage and recalculates size.
func (s *State) SetContext(pct float64) {
	s.ContextPct = pct
	s.Size = SizeFromContext(pct)
}

// Wake transitions from sleeping to bored.
func (s *State) Wake() {
	s.Mood = MoodBored
	now := time.Now()
	s.LastEvent = now
	s.LastMoodChange = now
}

// Sleep transitions to sleeping mood.
func (s *State) Sleep() {
	s.Mood = MoodSleeping
	s.LastEvent = time.Now()
}

// ComputeMood derives the current mood from the LastEvent timestamp.
// This replaces the old Tick() loop — mood is computed on-read.
// Mood changes are rate-limited to once per moodCooldown.
func (s *State) ComputeMood() {
	if s.Mood == MoodSleeping {
		return
	}
	if time.Since(s.LastMoodChange) < moodCooldown {
		return
	}
	elapsed := time.Since(s.LastEvent)
	isActive := s.Mood >= MoodEating && s.Mood <= MoodPouncing
	switch {
	case isActive && elapsed > 3*time.Second:
		s.Mood = IdleMoods[rand.Intn(len(IdleMoods))]
		s.LastMoodChange = time.Now()
	case elapsed > 60*time.Second:
		s.Mood = MoodSleeping
		s.LastMoodChange = time.Now()
	}
}

// CleanStaleStates removes state files not modified in the given duration.
func CleanStaleStates(maxAge time.Duration) {
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "ccpetline-state") {
			continue
		}
		// State files end in .json; orphaned temp files (crash between
		// CreateTemp and rename in SaveState) contain .json.tmp-.
		if !strings.HasSuffix(name, ".json") && !strings.Contains(name, ".json.tmp-") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if time.Since(info.ModTime()) > maxAge {
			os.Remove(filepath.Join(stateDir, name))
		}
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
// The temp file name must be unique: the hook, statusline, and config
// binaries can all write the same state file concurrently.
func SaveState(path string, s *State) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0644); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("chmod temp: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		// Clean up temp file on rename failure
		os.Remove(tmp.Name())
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
