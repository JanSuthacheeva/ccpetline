package pet

import (
	"os"
	"path/filepath"
	"testing"
)

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
