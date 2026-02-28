package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

const SocketPath = "/tmp/claude-pet.sock"

type EventType string

const (
	EventSnack         EventType = "snack"
	EventWake          EventType = "wake"
	EventSleep         EventType = "sleep"
	EventContextUpdate EventType = "context_update"
)

type Event struct {
	Type       EventType `json:"type"`
	ToolName   string    `json:"tool_name,omitempty"`
	ContextPct float64   `json:"context_pct,omitempty"`
}

// Send writes a single JSON-line event to a connection.
func Send(conn net.Conn, e Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(conn, "%s\n", data)
	return err
}

// SendToSocket dials the unix socket, sends one event, and closes.
func SendToSocket(e Event) error {
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		return err
	}
	defer conn.Close()
	return Send(conn, e)
}

// ReadEvents reads JSON-line events from a reader, calling fn for each.
func ReadEvents(r io.Reader, fn func(Event)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}
		fn(e)
	}
}
