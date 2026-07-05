package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jansuthacheeva/ccpetline/internal/pet"
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

	petLines := pet.RenderLines(state, claudeJSON)

	if state.DisplayMode == pet.ModePrepend || state.DisplayMode == pet.ModeAppend {
		wrappedLines := pet.RunWrapCommand(state.WrapCommand, data)
		var combined []string
		if state.DisplayMode == pet.ModePrepend {
			combined = append(combined, petLines...)
			combined = append(combined, wrappedLines...)
		} else {
			combined = append(combined, wrappedLines...)
			combined = append(combined, petLines...)
		}
		for _, line := range combined {
			line = strings.ReplaceAll(line, " ", "\u00A0")
			fmt.Fprintf(os.Stdout, "\x1b[0m%s\n", line)
		}
	} else {
		for _, line := range petLines {
			line = strings.ReplaceAll(line, " ", "\u00A0")
			fmt.Fprintf(os.Stdout, "\x1b[0m%s\n", line)
		}
	}
}
