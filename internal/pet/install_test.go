package pet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readSettings(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading settings: %v", err)
	}
	var s map[string]any
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("parsing settings: %v", err)
	}
	return s
}

func hookEvents(t *testing.T, s map[string]any) map[string][]any {
	t.Helper()
	hooks, ok := s["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("no hooks object in settings: %v", s)
	}
	out := make(map[string][]any)
	for event, v := range hooks {
		entries, ok := v.([]any)
		if !ok {
			t.Fatalf("hooks[%s] is not a list: %v", event, v)
		}
		out[event] = entries
	}
	return out
}

func TestInstallFreshSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".claude", "settings.json")
	if err := installToSettingsFile(path); err != nil {
		t.Fatalf("install into missing file: %v", err)
	}
	s := readSettings(t, path)

	sl, ok := s["statusLine"].(map[string]any)
	if !ok || sl["command"] != "ccpetline" || sl["type"] != "command" {
		t.Errorf("statusLine not installed: %v", s["statusLine"])
	}
	events := hookEvents(t, s)
	for _, event := range []string{"PostToolUse", "SessionStart", "SessionEnd"} {
		if len(events[event]) != 1 || !hookEntryHasPetline(events[event][0]) {
			t.Errorf("%s hook not installed: %v", event, events[event])
		}
	}
	if m, _ := events["PostToolUse"][0].(map[string]any); m["matcher"] != "*" {
		t.Errorf("PostToolUse entry missing matcher: %v", events["PostToolUse"][0])
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	for i := 0; i < 2; i++ {
		if err := installToSettingsFile(path); err != nil {
			t.Fatalf("install run %d: %v", i+1, err)
		}
	}
	events := hookEvents(t, readSettings(t, path))
	for event, entries := range events {
		if len(entries) != 1 {
			t.Errorf("%s has %d entries after two installs, want 1", event, len(entries))
		}
	}
}

func TestInstallPreservesExistingSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	existing := `{
		"model": "opus",
		"hooks": {
			"PostToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "other-tool"}]}]
		}
	}`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}
	if err := installToSettingsFile(path); err != nil {
		t.Fatalf("install over existing settings: %v", err)
	}
	s := readSettings(t, path)
	if s["model"] != "opus" {
		t.Errorf("unrelated setting lost: model = %v", s["model"])
	}
	events := hookEvents(t, s)
	if len(events["PostToolUse"]) != 2 {
		t.Fatalf("PostToolUse should keep the foreign entry and add ours, got %v", events["PostToolUse"])
	}
	if hookEntryHasPetline(events["PostToolUse"][0]) {
		t.Error("foreign hook entry was replaced")
	}
	if !hookEntryHasPetline(events["PostToolUse"][1]) {
		t.Error("ccpetline hook entry not appended")
	}
}

func TestInstallRejectsMalformedSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := installToSettingsFile(path); err == nil {
		t.Fatal("install over malformed settings should fail, not overwrite")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "{not json" {
		t.Errorf("malformed settings file was modified: %q, %v", data, err)
	}
}

func TestInstallLeavesNoTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := installToSettingsFile(path); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "settings.json" {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("expected only settings.json, got %v", names)
	}
}
