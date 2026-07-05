package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jansuthacheeva/ccpetline/internal/pet"
)

func main() {
	// The hook runs async and best-effort: it always exits 0 and reports
	// problems on stderr only.
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}
	in := pet.ParseClaudeInput(data)
	if in == nil {
		return
	}

	statePath := pet.StatePath(in.SessionID)
	state := pet.LoadState(statePath)

	switch in.HookEventName {
	case "PostToolUse":
		state.Feed()
	case "SessionStart":
		if in.Source == "resume" {
			state.Wake()
		} else {
			state = pet.NewState()
			state.Wake()
		}
		pet.CleanStaleStates(14 * 24 * time.Hour)
	case "SessionEnd":
		state.Sleep()
	default:
		return
	}

	if err := pet.SaveState(statePath, state); err != nil {
		fmt.Fprintf(os.Stderr, "ccpetline-hook: saving state: %v\n", err)
	}
}
