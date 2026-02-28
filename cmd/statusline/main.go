package main

import (
	"bytes"
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

	// Load pet state and compute mood
	state := pet.LoadState(pet.DefaultStatePath)
	state.ComputeMood()

	// Line 1: pet status
	petLine := pet.FormatPetLine(state)
	petLine = strings.ReplaceAll(petLine, " ", "\u00A0")
	fmt.Fprintf(os.Stdout, "\x1b[0m%s\n", petLine)
	fmt.Fprintf(os.Stdout, "\x1b[0m%s\n", strings.Repeat("\u2500", 40))

	// Remaining lines: delegate to ccstatusline
	cmd := exec.Command("npx", "-y", "ccstatusline@latest")
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
