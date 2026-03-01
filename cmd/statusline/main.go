package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

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

	petLines := pet.RenderLines(state, claudeJSON)

	if state.DisplayMode == pet.ModePrepend || state.DisplayMode == pet.ModeAppend {
		wrappedLines := runWrapCommand(state.WrapCommand, data)
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

// runWrapCommand executes the wrap command with the given stdin data and returns
// its stdout lines. Returns nil on error or timeout.
func runWrapCommand(command string, stdinData []byte) []string {
	if command == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdin = bytes.NewReader(stdinData)

	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	raw := strings.TrimRight(string(out), "\n")
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}
