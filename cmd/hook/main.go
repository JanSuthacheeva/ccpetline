package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/jan/claude-pet/internal/pet"
)

// hookInput is the JSON structure Claude Code sends on stdin to hooks.
type hookInput struct {
	HookEventName string `json:"hook_event_name"`
	ToolName      string `json:"tool_name"`
}

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(0) // fail silently — async hook
	}

	var h hookInput
	if err := json.Unmarshal(data, &h); err != nil {
		os.Exit(0)
	}

	state := pet.LoadState(pet.DefaultStatePath)

	switch h.HookEventName {
	case "PostToolUse":
		state.Feed(h.ToolName)
	case "SessionStart":
		state.Wake()
	case "SessionEnd":
		state.Sleep()
	default:
		os.Exit(0)
	}

	if err := pet.SaveState(pet.DefaultStatePath, state); err != nil {
		os.Exit(1)
	}

	// Print snack flavor for fun (visible in hook logs)
	if h.HookEventName == "PostToolUse" {
		os.Stdout.WriteString(pet.SnackFlavor(h.ToolName) + "\n")
	}
}
