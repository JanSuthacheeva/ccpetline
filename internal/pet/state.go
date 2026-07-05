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

// Species identifies which pet the user picked.
type Species string

const (
	SpeciesGoose  Species = "goose"
	SpeciesCat    Species = "cat"
	SpeciesOcean  Species = "ocean"
	SpeciesDragon Species = "dragon"
	SpeciesDino   Species = "dino"
)

// AllSpecies is the ordered list of selectable species.
var AllSpecies = []Species{SpeciesGoose, SpeciesCat, SpeciesOcean, SpeciesDragon, SpeciesDino}

// ContextMode selects how context usage is displayed: raw ({ctx}) or scaled
// to the usable window before auto-compact ({ctx_u}).
type ContextMode string

const (
	ContextModeCtx  ContextMode = "ctx"
	ContextModeCtxU ContextMode = "ctx_u"
)

// ParseContextMode normalizes a raw string to a known context mode,
// defaulting to ctx.
func ParseContextMode(s string) ContextMode {
	if ContextMode(s) == ContextModeCtxU {
		return ContextModeCtxU
	}
	return ContextModeCtx
}

// ParseSpecies normalizes a raw string to a known species, defaulting to the
// same cat that defaultConfig uses.
func ParseSpecies(s string) Species {
	switch Species(s) {
	case SpeciesGoose, SpeciesCat, SpeciesOcean, SpeciesDragon, SpeciesDino:
		return Species(s)
	default:
		return SpeciesCat
	}
}

// Mood is the pet's current activity, derived from recent hook events.
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

var (
	// ActiveMoods are picked randomly while tools are running.
	ActiveMoods = []Mood{MoodEating, MoodChasing, MoodDigging, MoodFetching, MoodPouncing}
	// IdleMoods are picked randomly once activity dies down.
	IdleMoods = []Mood{MoodBored, MoodNapping, MoodGrooming, MoodWandering}
)

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

// Size is the pet's growth stage, derived from context usage.
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

// State is the per-session pet state persisted between hook and statusline
// invocations. It duplicates the display settings from Config as a snapshot;
// ApplyConfig is the single place that copy happens.
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

// NewState returns a fresh sleeping pet configured from the user's config.
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
func (s *State) Feed() {
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
// This replaces the old Tick() loop - mood is computed on-read.
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

// isStateFile reports whether name is a ccpetline session state file.
func isStateFile(name string) bool {
	return strings.HasPrefix(name, "ccpetline-state") && strings.HasSuffix(name, ".json")
}

// isStateTempFile reports whether name is a temp file orphaned by a crash
// between CreateTemp and rename in SaveState.
func isStateTempFile(name string) bool {
	return strings.HasPrefix(name, "ccpetline-state") && strings.Contains(name, ".json.tmp-")
}

// statePaths returns the full paths of all session state files in stateDir.
func statePaths() []string {
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return nil
	}
	var paths []string
	for _, e := range entries {
		if !e.IsDir() && isStateFile(e.Name()) {
			paths = append(paths, filepath.Join(stateDir, e.Name()))
		}
	}
	return paths
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
		if !isStateFile(name) && !isStateTempFile(name) {
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
	normalizeState(&s)
	return &s
}

// normalizeState applies the same validation to a loaded state that
// LoadConfig applies to configs, so a hand-edited or corrupted state file
// cannot render outside sane bounds.
func normalizeState(s *State) {
	s.Species = ParseSpecies(string(s.Species))
	s.ContextMode = ParseContextMode(string(s.ContextMode))
	s.BarWidth = clampBarWidth(s.BarWidth)
}

// SaveState writes pet state to a JSON file atomically. The write goes
// through writeFileAtomic because the hook, statusline, and config binaries
// can all save the same state file concurrently.
func SaveState(path string, s *State) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	if err := writeFileAtomic(path, data, 0o644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}
