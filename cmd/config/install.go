package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func installToClaudeCode() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")

	var settings map[string]interface{}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Sprintf("Error reading settings: %v", err)
		}
		settings = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Sprintf("Error parsing settings: %v", err)
		}
	}

	// Status line
	settings["statusLine"] = map[string]interface{}{
		"type":    "command",
		"command": "ccpetline",
	}

	// Hooks — append to existing entries, skip if ccpetline-hook already present
	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	petHookEntry := map[string]interface{}{"type": "command", "command": "ccpetline-hook", "async": true}

	// PostToolUse needs a matcher
	appendHookEntry(hooks, "PostToolUse", map[string]interface{}{
		"matcher": "*",
		"hooks":   []interface{}{petHookEntry},
	})
	// SessionStart / SessionEnd have no matcher
	simpleEntry := map[string]interface{}{
		"hooks": []interface{}{petHookEntry},
	}
	appendHookEntry(hooks, "SessionStart", simpleEntry)
	appendHookEntry(hooks, "SessionEnd", simpleEntry)

	settings["hooks"] = hooks

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error encoding settings: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err)
	}

	if err := os.WriteFile(settingsPath, append(out, '\n'), 0o644); err != nil {
		return fmt.Sprintf("Error writing settings: %v", err)
	}

	return "Installed! Restart Claude Code to activate."
}

// hookEntryHasPetline checks if a hook entry already references ccpetline-hook.
func hookEntryHasPetline(entry interface{}) bool {
	m, ok := entry.(map[string]interface{})
	if !ok {
		return false
	}
	cmds, ok := m["hooks"].([]interface{})
	if !ok {
		return false
	}
	for _, cmd := range cmds {
		if h, ok := cmd.(map[string]interface{}); ok {
			if h["command"] == "ccpetline-hook" {
				return true
			}
		}
	}
	return false
}

// appendHookEntry appends a hook entry to the given event key, unless
// ccpetline-hook is already present in any existing entry.
func appendHookEntry(hooks map[string]interface{}, event string, entry map[string]interface{}) {
	existing, _ := hooks[event].([]interface{})
	for _, e := range existing {
		if hookEntryHasPetline(e) {
			return
		}
	}
	hooks[event] = append(existing, entry)
}
