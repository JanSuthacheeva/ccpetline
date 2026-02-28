package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jan/claude-pet/internal/pet"
	"github.com/jan/claude-pet/internal/protocol"
)

func main() {
	tmux := flag.Bool("tmux", false, "auto-create a tmux bottom pane and run there")
	rows := flag.Int("rows", 8, "tmux pane height")
	flag.Parse()

	if *tmux {
		launchInTmux(*rows)
		return
	}

	run()
}

// launchInTmux splits a bottom pane in the current tmux window and runs claude-pet there.
func launchInTmux(rows int) {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot find own executable: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("tmux", "split-window", "-v", "-l", fmt.Sprintf("%d", rows), exe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tmux split failed: %v\n", err)
		os.Exit(1)
	}
}

func run() {
	// Clean up old socket
	os.Remove(protocol.SocketPath)

	ln, err := net.Listen("unix", protocol.SocketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		ln.Close()
		os.Remove(protocol.SocketPath)
	}()

	state := pet.NewState()
	var mu sync.Mutex

	// Accept connections in background
	go acceptLoop(ln, state, &mu)

	// Handle signals for clean shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			return
		case <-ticker.C:
			mu.Lock()
			state.Tick()
			frame := pet.Render(state, termWidth())
			mu.Unlock()
			clearScreen()
			fmt.Print(frame)
		}
	}
}

func acceptLoop(ln net.Listener, state *pet.State, mu *sync.Mutex) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go handleConn(conn, state, mu)
	}
}

func handleConn(conn net.Conn, state *pet.State, mu *sync.Mutex) {
	defer conn.Close()
	protocol.ReadEvents(conn, func(e protocol.Event) {
		mu.Lock()
		defer mu.Unlock()
		switch e.Type {
		case protocol.EventSnack:
			state.Feed(e.ToolName)
		case protocol.EventWake:
			state.Wake()
		case protocol.EventSleep:
			state.Sleep()
		case protocol.EventContextUpdate:
			state.SetContext(e.ContextPct)
		}
	})
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func termWidth() int {
	cmd := exec.Command("tput", "cols")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 60
	}
	s := strings.TrimSpace(string(out))
	var w int
	fmt.Sscanf(s, "%d", &w)
	if w < 20 {
		return 60
	}
	return w
}
