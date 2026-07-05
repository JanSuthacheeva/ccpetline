package pet

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestSizeFromContext(t *testing.T) {
	tests := []struct {
		pct  float64
		want Size
	}{
		{0, SizeTiny},
		{20, SizeTiny},
		{20.1, SizeNormal},
		{35, SizeNormal},
		{35.1, SizeChonky},
		{60, SizeChonky},
		{60.1, SizeMegaChonk},
		{100, SizeMegaChonk},
	}
	for _, tt := range tests {
		if got := SizeFromContext(tt.pct); got != tt.want {
			t.Errorf("SizeFromContext(%v) = %v, want %v", tt.pct, got, tt.want)
		}
	}
}

func TestComputeMood(t *testing.T) {
	now := time.Now()

	t.Run("sleeping never wakes on its own", func(t *testing.T) {
		s := &State{Mood: MoodSleeping, LastEvent: now.Add(-time.Hour), LastMoodChange: now.Add(-time.Hour)}
		s.ComputeMood()
		if s.Mood != MoodSleeping {
			t.Errorf("mood = %v", s.Mood)
		}
	})

	t.Run("cooldown blocks any change", func(t *testing.T) {
		s := &State{Mood: MoodEating, LastEvent: now.Add(-10 * time.Second), LastMoodChange: now}
		s.ComputeMood()
		if s.Mood != MoodEating {
			t.Errorf("mood = %v", s.Mood)
		}
	})

	t.Run("active mood goes idle after inactivity", func(t *testing.T) {
		s := &State{Mood: MoodEating, LastEvent: now.Add(-10 * time.Second), LastMoodChange: now.Add(-2 * moodCooldown)}
		s.ComputeMood()
		if !slices.Contains(IdleMoods, s.Mood) {
			t.Errorf("mood = %v, want an idle mood", s.Mood)
		}
		if time.Since(s.LastMoodChange) > time.Second {
			t.Error("LastMoodChange not updated")
		}
	})

	t.Run("recent activity keeps the active mood", func(t *testing.T) {
		s := &State{Mood: MoodChasing, LastEvent: now.Add(-time.Second), LastMoodChange: now.Add(-2 * moodCooldown)}
		s.ComputeMood()
		if s.Mood != MoodChasing {
			t.Errorf("mood = %v", s.Mood)
		}
	})

	t.Run("idle mood falls asleep after a minute", func(t *testing.T) {
		s := &State{Mood: MoodBored, LastEvent: now.Add(-2 * time.Minute), LastMoodChange: now.Add(-2 * moodCooldown)}
		s.ComputeMood()
		if s.Mood != MoodSleeping {
			t.Errorf("mood = %v", s.Mood)
		}
	})
}

func TestStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ccpetline-state-rt.json")
	in := &State{
		Species:     SpeciesDino,
		ContextMode: ContextModeCtxU,
		Lines:       []string{"{model}"},
		BarStyle:    BarBlock,
		BarShowPet:  true,
		BarWidth:    42,
		Mood:        MoodGrooming,
		Size:        SizeChonky,
		ContextPct:  61.5,
		Happiness:   7,
		LastEvent:   time.Now().Round(time.Second),
	}
	if err := SaveState(path, in); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	out := LoadState(path)
	if out.Species != in.Species || out.ContextMode != in.ContextMode ||
		out.BarStyle != in.BarStyle || out.BarWidth != in.BarWidth ||
		out.Mood != in.Mood || out.Size != in.Size ||
		out.ContextPct != in.ContextPct || out.Happiness != in.Happiness {
		t.Errorf("round trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
	if !out.LastEvent.Equal(in.LastEvent) {
		t.Errorf("LastEvent = %v, want %v", out.LastEvent, in.LastEvent)
	}
}

func TestLoadStateFallsBackToNewState(t *testing.T) {
	writeConfig(t, "") // redirect HOME so NewState sees pure defaults
	missing := LoadState(filepath.Join(t.TempDir(), "nope.json"))
	if missing.Mood != MoodSleeping || missing.Species != SpeciesCat {
		t.Errorf("missing file should yield a fresh default state: %+v", missing)
	}

	path := filepath.Join(t.TempDir(), "corrupt.json")
	if err := os.WriteFile(path, []byte("{torn"), 0644); err != nil {
		t.Fatal(err)
	}
	corrupt := LoadState(path)
	if corrupt.Mood != MoodSleeping || corrupt.Species != SpeciesCat {
		t.Errorf("corrupt file should yield a fresh default state: %+v", corrupt)
	}
}

func TestApplyConfig(t *testing.T) {
	show := false
	cfg := &Config{
		Species:      SpeciesDragon,
		ContextMode:  ContextModeCtxU,
		IconTheme:    IconThemeNerd,
		Lines:        []string{"{model}"},
		LineColors:   [][]uint8{{42}},
		DisplayMode:  ModeAppend,
		WrapCommand:  "echo hi",
		BarStyle:     BarDot,
		BarShowPet:   &show,
		BarWidth:     33,
		Powerline:    true,
		PowerlineSep: SepRound,
	}
	s := &State{BarShowPet: true, Mood: MoodBored}
	s.ApplyConfig(cfg)

	if s.Species != SpeciesDragon || s.ContextMode != ContextModeCtxU || s.IconTheme != IconThemeNerd {
		t.Errorf("species/context/icon not applied: %+v", s)
	}
	if len(s.Lines) != 1 || s.Lines[0] != "{model}" || len(s.LineColors) != 1 {
		t.Errorf("lines/colors not applied: %+v", s)
	}
	if s.DisplayMode != ModeAppend || s.WrapCommand != "echo hi" {
		t.Errorf("display mode/wrap command not applied: %+v", s)
	}
	if s.BarStyle != BarDot || s.BarShowPet || s.BarWidth != 33 {
		t.Errorf("bar settings not applied: %+v", s)
	}
	if !s.Powerline || s.PowerlineSep != SepRound {
		t.Errorf("powerline settings not applied: %+v", s)
	}
	if s.Mood != MoodBored {
		t.Errorf("ApplyConfig must not touch pet state, mood changed to %v", s.Mood)
	}

	// A nil BarShowPet keeps the state's existing value.
	cfg.BarShowPet = nil
	s.BarShowPet = true
	s.ApplyConfig(cfg)
	if !s.BarShowPet {
		t.Error("nil BarShowPet should preserve the existing state value")
	}
}

func TestSaveStateCreatesParentDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "ccpetline-state-test.json")
	if err := SaveState(path, NewState()); err != nil {
		t.Fatalf("SaveState into missing parent dir: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("state file not written: %v", err)
	}
}

func TestSaveStateLeavesNoTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ccpetline-state-test.json")
	if err := SaveState(path, NewState()); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "ccpetline-state-test.json" {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("expected only the state file, got %v", names)
	}
}
