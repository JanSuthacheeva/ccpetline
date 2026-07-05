package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jansuthacheeva/ccpetline/internal/pet"
)

func main() {
	// The statusline is best-effort display code: it never exits non-zero.
	// Missing or malformed stdin just renders the pet from persisted state.
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		data = nil
	}
	in := pet.ParseClaudeInput(data)

	var sessionID string
	if in != nil {
		sessionID = in.SessionID
	}
	statePath := pet.StatePath(sessionID)
	state := pet.LoadState(statePath)
	if in != nil && in.ContextWindow.UsedPercentage != nil {
		state.SetContext(*in.ContextWindow.UsedPercentage)
	}
	state.ComputeMood()
	if err := pet.SaveState(statePath, state); err != nil {
		fmt.Fprintf(os.Stderr, "ccpetline: saving state: %v\n", err)
	}

	lines := pet.RenderLines(state, in)

	if state.DisplayMode == pet.ModePrepend || state.DisplayMode == pet.ModeAppend {
		wrapped := pet.RunWrapCommand(state.WrapCommand, data)
		if state.DisplayMode == pet.ModePrepend {
			lines = append(lines, wrapped...)
		} else {
			lines = append(wrapped, lines...)
		}
	}

	for _, line := range lines {
		// Claude Code collapses runs of regular spaces in statusline output,
		// so substitute NBSP to preserve the layout.
		line = strings.ReplaceAll(line, " ", "\u00A0")
		fmt.Fprintf(os.Stdout, "\x1b[0m%s\n", line)
	}
}
