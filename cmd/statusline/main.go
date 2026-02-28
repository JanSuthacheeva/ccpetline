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

	var claudeJSON map[string]any
	if err := json.Unmarshal(data, &claudeJSON); err != nil {
		os.Exit(1)
	}

	// Load pet state and update context from Claude's JSON
	state := pet.LoadState(pet.DefaultStatePath)

	if cw, ok := claudeJSON["context_window"].(map[string]any); ok {
		if pct, ok := cw["used_percentage"].(float64); ok && pct > 0 {
			state.SetContext(pct)
		}
	}

	state.ComputeMood()

	lines := pet.FormatStatusLine(state, claudeJSON)
	for _, line := range lines {
		// Replace spaces with non-breaking spaces to prevent trimming
		line = strings.ReplaceAll(line, " ", "\u00A0")
		// Prefix with ANSI reset to undo dim styling
		fmt.Fprintf(os.Stdout, "\x1b[0m%s\n", line)
	}
}
