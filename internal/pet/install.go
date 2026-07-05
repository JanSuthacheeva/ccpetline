package pet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ClaudeSettingsPath returns the path of the user's Claude Code settings file.
func ClaudeSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// InstallToClaudeCode registers the ccpetline statusline and hook binaries
// in the user's Claude Code settings.json, preserving all existing settings.
func InstallToClaudeCode() error {
	path, err := ClaudeSettingsPath()
	if err != nil {
		return err
	}
	return installToSettingsFile(path)
}

func installToSettingsFile(path string) error {
	settings := make(map[string]any)
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parsing settings: %w", err)
		}
	case !os.IsNotExist(err):
		return fmt.Errorf("reading settings: %w", err)
	}

	// Status line
	settings["statusLine"] = map[string]any{
		"type":    "command",
		"command": "ccpetline",
	}

	// Hooks — append to existing entries, skip if ccpetline-hook already present
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = make(map[string]any)
	}

	petHookEntry := map[string]any{"type": "command", "command": "ccpetline-hook", "async": true}

	// PostToolUse needs a matcher
	appendHookEntry(hooks, "PostToolUse", map[string]any{
		"matcher": "*",
		"hooks":   []any{petHookEntry},
	})
	// SessionStart / SessionEnd have no matcher
	simpleEntry := map[string]any{
		"hooks": []any{petHookEntry},
	}
	appendHookEntry(hooks, "SessionStart", simpleEntry)
	appendHookEntry(hooks, "SessionEnd", simpleEntry)

	settings["hooks"] = hooks

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding settings: %w", err)
	}

	// Claude Code itself reads and rewrites this file, so never leave it torn.
	if err := writeFileAtomic(path, append(out, '\n'), 0644); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}
	return nil
}

// hookEntryHasPetline checks if a hook entry already references ccpetline-hook.
func hookEntryHasPetline(entry any) bool {
	m, ok := entry.(map[string]any)
	if !ok {
		return false
	}
	cmds, ok := m["hooks"].([]any)
	if !ok {
		return false
	}
	for _, cmd := range cmds {
		if h, ok := cmd.(map[string]any); ok {
			if h["command"] == "ccpetline-hook" {
				return true
			}
		}
	}
	return false
}

// appendHookEntry appends a hook entry to the given event key, unless
// ccpetline-hook is already present in any existing entry.
func appendHookEntry(hooks map[string]any, event string, entry map[string]any) {
	existing, _ := hooks[event].([]any)
	for _, e := range existing {
		if hookEntryHasPetline(e) {
			return
		}
	}
	hooks[event] = append(existing, entry)
}
