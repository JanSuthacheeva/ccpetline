# claude-pet — Plan

## Context

Build a terminal "pet" that hooks into Claude Code sessions via the hooks system. The pet eats tool calls as "snacks" and visually grows fatter as usage increases. It runs in a tmux bottom pane on the same screen as Claude Code.

## Decisions

- **Language:** Go
- **Display:** Tmux bottom pane (~6-8 rows), auto-created
- **Communication:** Unix domain socket (`/tmp/claude-pet.sock`)
- **Persistence:** None — fresh pet each session
- **Hook events:** `PostToolUse` (snack), `SessionStart` (wake), `SessionEnd` (sleep)
- **Context %:** Use Claude Code's statusline fields (`contextUsed`/`contextLimit`) — the hook binary can also forward these if available, or we configure the statusline to pipe context data to the pet

## Architecture

```
Claude Code
  |
  |-- PostToolUse hook (async) --> cmd/hook/       -- sends "snack" event
  |                                   |
  |-- Statusline command ----------> cmd/statusline/ -- sends context % then
  |                                   |                pipes to real statusline
  |                                   v
  |                            unix socket (/tmp/claude-pet.sock)
  |                                   |
  |                                   v
  |                            cmd/pet/  -- TUI in tmux bottom pane
  |                                   |
  |                                   uses
  |                                   v
  |                            internal/pet/    -- state, render
  |                            internal/socket/ -- protocol
```

Three binaries:
- `claude-pet` — the TUI pet (socket server + render loop)
- `claude-pet-hook` — hook handler (reads hook JSON, sends snack/wake/sleep events)
- `claude-pet-statusline` — statusline wrapper (extracts context %, forwards to pet, pipes to real statusline)

## Pet mechanics

| Signal | Source | Effect |
|--------|--------|--------|
| Snack | PostToolUse hook | +1 snack, mood -> eating, brief animation |
| Context update | Statusline polling | Updates fatness (primary size driver) |
| Wake up | SessionStart hook | mood -> happy |
| Sleep | SessionEnd hook | mood -> sleeping |
| Idle | no events for 10s | mood -> idle -> bored, pet wanders |

### Size stages (based on context window %)

The pet's physical size is driven by **context usage %**, not snack count. Snacks are the feeding animation / mood trigger; context % is what makes it fat.

1. **Tiny** (0-20%) — small, energetic, lots of room
2. **Normal** (21-45%) — default size
3. **Chonky** (46-70%) — wider body, slower
4. **Mega chonk** (71-90%) — very wide, sweating
5. **Absolute unit** (91-100%) — fills the pane, about to explode

### Context % source — Statusline wrapper

Claude Code's statusline command receives JSON on stdin with these fields:
- `context_window.used_percentage` — pre-calculated 0-100
- `context_window.context_window_size` — max tokens
- `context_window.current_usage.input_tokens` — current input tokens

Updates fire after each assistant message (debounced 300ms).

**Approach:** A wrapper script that:
1. Reads JSON from stdin (tee)
2. Extracts `context_window.used_percentage`
3. Sends it to the pet via the unix socket
4. Pipes the original JSON to the user's real statusline command (e.g. `ccstatusline`)

This way we **don't replace** the user's statusline — we intercept and forward.

Statusline config becomes:
```json
{
  "statusLine": {
    "type": "command",
    "command": "claude-pet-statusline npx -y ccstatusline@latest",
    "padding": 0
  }
}
```

Where `claude-pet-statusline` is a tiny Go binary (or shell script) that wraps the real command.

### Snack flavors (just for fun display)

- Bash -> "spicy taco"
- Read -> "mild salad"
- Edit/Write -> "crunchy cookie"
- Grep/Glob -> "popcorn"
- Agent -> "whole pizza"
- Other -> "mystery snack"

## Files to create

```
claude-pet/
  go.mod
  Makefile
  README.md
  cmd/pet/main.go             -- TUI: socket server, tick loop, render, tmux pane mgmt
  cmd/hook/main.go            -- hook handler: read stdin JSON, send snack/wake/sleep to socket
  cmd/statusline/main.go      -- statusline wrapper: extract context %, send to socket, pipe to real cmd
  internal/pet/state.go       -- State struct, Feed(), SetContext(), Tick(), size/mood logic
  internal/pet/render.go      -- ASCII art per size x mood, frame animation
  internal/protocol/protocol.go -- Event types + JSON-lines codec for unix socket
  hooks.example.json          -- example Claude Code hook + statusline config
```

## Implementation order

1. ~~**internal/protocol/**~~ DONE
2. ~~**internal/pet/state.go**~~ DONE
3. ~~**internal/pet/render.go**~~ DONE
4. ~~**cmd/hook/main.go**~~ DONE
5. ~~**cmd/statusline/main.go**~~ DONE
6. ~~**cmd/pet/main.go**~~ DONE
7. ~~**Makefile**~~ DONE
8. ~~**README.md**~~ DONE
9. ~~**hooks.example.json**~~ DONE

## Status

**Initial build complete.** All binaries, protocol, state machine, renderer, and config are in place.

## What's next

- [ ] Manual testing (verify build, hook events, sleeping/eating/wandering)
- [ ] Live integration test with actual Claude Code hooks
- [ ] Polish: color support, sound effects, more animations
- [ ] Persistence: save snack count across sessions
- [ ] Multi-session support (multiple pets / socket namespacing)
- [ ] Package / release (goreleaser, homebrew tap)
