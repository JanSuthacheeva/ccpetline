package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/jan/claude-pet/internal/pet"
	"github.com/jan/claude-pet/internal/protocol"
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

	var evt protocol.Event

	switch h.HookEventName {
	case "PostToolUse":
		evt = protocol.Event{
			Type:     protocol.EventSnack,
			ToolName: h.ToolName,
		}
	case "SessionStart":
		evt = protocol.Event{Type: protocol.EventWake}
	case "SessionEnd":
		evt = protocol.Event{Type: protocol.EventSleep}
	default:
		os.Exit(0)
	}

	// Best-effort: if pet isn't running, just exit.
	_ = protocol.SendToSocket(evt)

	// Print snack flavor for fun (visible in hook logs)
	if evt.Type == protocol.EventSnack {
		os.Stdout.WriteString(pet.SnackFlavor(h.ToolName) + "\n")
	}
}
