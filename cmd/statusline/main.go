package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"

	"github.com/jan/claude-pet/internal/protocol"
)

// statuslineInput matches Claude Code's statusline JSON (subset).
type statuslineInput struct {
	ContextWindow struct {
		UsedPercentage float64 `json:"used_percentage"`
	} `json:"context_window"`
}

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	// Extract context % and send to pet
	var sl statuslineInput
	if err := json.Unmarshal(data, &sl); err == nil && sl.ContextWindow.UsedPercentage > 0 {
		_ = protocol.SendToSocket(protocol.Event{
			Type:       protocol.EventContextUpdate,
			ContextPct: sl.ContextWindow.UsedPercentage,
		})
	}

	// If a wrapped command was given as args, pipe stdin to it
	if len(os.Args) > 1 {
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmd.Stdin = bytes.NewReader(data)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}
	}
}

