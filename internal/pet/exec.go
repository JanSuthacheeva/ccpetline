package pet

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// commandTimeout bounds every subprocess the statusline spawns per render.
// The statusline runs on every Claude Code refresh, so a slow git (e.g. a
// cold network filesystem) must never block it.
const commandTimeout = 500 * time.Millisecond

// wrapCommandTimeout bounds the user-configured wrap command, which renders
// a whole third-party statusline and gets a more generous budget.
const wrapCommandTimeout = 2 * time.Second

// runCommand executes a command with the standard timeout and returns its
// trimmed stdout. All statusline subprocess calls go through here so the
// timeout policy lives in one place.
func runCommand(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// RunWrapCommand executes the configured wrap command with the given stdin
// data and returns its stdout lines. Returns nil on error or timeout.
func RunWrapCommand(command string, stdinData []byte) []string {
	if command == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), wrapCommandTimeout)
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
