package main

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/jansuthacheeva/ccpetline/internal/pet"
)

// hookInput is the JSON structure Claude Code sends on stdin to hooks.
type hookInput struct {
	SessionID     string `json:"session_id"`
	HookEventName string `json:"hook_event_name"`
	Source        string `json:"source"`
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

	statePath := pet.StatePath(h.SessionID)
	state := pet.LoadState(statePath)

	switch h.HookEventName {
	case "PostToolUse":
		state.Feed(h.ToolName)
	case "SessionStart":
		if h.Source == "resume" {
			state.Wake()
		} else {
			state = pet.NewState()
			state.Wake()
		}
		pet.CleanStaleStates(14 * 24 * time.Hour)
	case "SessionEnd":
		state.Sleep()
	default:
		os.Exit(0)
	}

	if err := pet.SaveState(statePath, state); err != nil {
		os.Exit(1)
	}
}
