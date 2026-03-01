package pet

import "os"

// KittySupported returns true if the terminal likely supports the Kitty
// graphics protocol. Detection is env-based since the statusline binary
// has piped stdin (no tty query possible).
func KittySupported() bool {
	if v := os.Getenv("CLAUDE_PET_KITTY"); v != "" {
		return v == "1"
	}
	if os.Getenv("KITTY_PID") != "" || os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}
	switch os.Getenv("TERM_PROGRAM") {
	case "kitty", "ghostty", "WezTerm":
		return true
	}
	return false
}
