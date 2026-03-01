package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jan/claude-pet/internal/pet"
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	// Load pet state, update context from Claude's JSON, compute mood
	var claudeJSON map[string]any
	var sessionID string
	if json.Unmarshal(data, &claudeJSON) == nil {
		if sid, ok := claudeJSON["session_id"].(string); ok {
			sessionID = sid
		}
	}
	statePath := pet.StatePath(sessionID)
	state := pet.LoadState(statePath)
	if claudeJSON != nil {
		if cw, ok := claudeJSON["context_window"].(map[string]any); ok {
			if pct, ok := cw["used_percentage"].(float64); ok {
				state.SetContext(pct)
			}
		}
	}
	state.ComputeMood()
	_ = pet.SaveState(statePath, state)

	lines := pet.RenderLines(state, claudeJSON, 50)
	for _, line := range lines {
		line = strings.ReplaceAll(line, " ", "\u00A0")
		fmt.Fprintf(os.Stdout, "\x1b[0m%s\n", line)
	}
}
