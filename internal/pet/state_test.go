package pet

import (
	"os"
	"path/filepath"
	"testing"
)

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
