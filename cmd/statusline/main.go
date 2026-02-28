package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/jan/claude-pet/internal/pet"
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	// Load pet state, update context from Claude's JSON, compute mood
	state := pet.LoadState(pet.DefaultStatePath)
	var claudeJSON map[string]any
	if json.Unmarshal(data, &claudeJSON) == nil {
		if cw, ok := claudeJSON["context_window"].(map[string]any); ok {
			if pct, ok := cw["used_percentage"].(float64); ok && pct > 0 {
				state.SetContext(pct)
			}
		}
	}
	state.ComputeMood()

	// Line 1: pet status
	petLine := pet.FormatPetLine(state)
	petLine = strings.ReplaceAll(petLine, " ", "\u00A0")
	fmt.Fprintf(os.Stdout, "\x1b[0m%s\n", petLine)
	sep := pet.FormatSeparator(state, 40)
	fmt.Fprintf(os.Stdout, "\x1b[0m%s\n", sep)

	// Remaining lines: delegate to ccstatusline, fall back to built-in
	cmd := exec.Command("npx", "-y", "ccstatusline@latest")
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if line := pet.FormatFallbackStatus(claudeJSON); line != "" {
			line = strings.ReplaceAll(line, " ", "\u00A0")
			fmt.Fprintf(os.Stdout, "\x1b[0m%s\n", line)
		}
	}
}
