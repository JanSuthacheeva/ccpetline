# ccpetline

A terminal pet that lives alongside your Claude Code sessions. It eats tool calls as "snacks" and grows fatter as your context window fills up.

## Quick start

```bash
make install    # builds and copies binaries to ~/.local/bin
```

Then add hooks to `~/.claude/settings.json` (see `hooks.example.json`).

## Architecture

```
Claude Code
  |-- PostToolUse hook -----> ccpetline-hook -----> /tmp state file
  |-- SessionStart hook ----> ccpetline-hook ----/
  |-- SessionEnd hook ------> ccpetline-hook ---/
  |-- Statusline -----------> ccpetline -> (reads state, renders status line)
```

Three binaries:

| Binary | Purpose |
|--------|---------|
| `ccpetline` | Statusline command. Reads pet state, renders status line output. |
| `ccpetline-hook` | Hook handler. Reads Claude Code hook JSON from stdin, updates pet state. |
| `ccpetline-config` | TUI configurator. |

## Pet mechanics

| Signal | Source | Effect |
|--------|--------|--------|
| Snack | PostToolUse hook | +1 snack, mood change |
| Context update | Statusline | Updates fatness (primary size driver) |
| Wake | SessionStart hook | Pet wakes up |
| Sleep | SessionEnd hook | Pet goes to sleep |
| Idle | No events for 10s | Gets bored, wanders |

### Size stages (driven by context window %)

1. **Tiny** (0-20%) -- small, energetic
2. **Normal** (21-35%) -- default
3. **Chonky** (36-60%) -- wider body
4. **Mega chonk** (61-100%) -- very wide, sweating

## Hook config

Add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [{
      "matcher": "*",
      "hooks": [{ "type": "command", "command": "ccpetline-hook", "async": true }]
    }],
    "SessionStart": [{
      "hooks": [{ "type": "command", "command": "ccpetline-hook", "async": true }]
    }],
    "SessionEnd": [{
      "hooks": [{ "type": "command", "command": "ccpetline-hook", "async": true }]
    }]
  },
  "statusLine": {
    "type": "command",
    "command": "ccpetline",
    "padding": 0
  }
}
```

## Testing manually

```bash
# Feed it
echo '{"hook_event_name":"SessionStart"}' | ./bin/ccpetline-hook
echo '{"hook_event_name":"PostToolUse","tool_name":"Bash"}' | ./bin/ccpetline-hook
echo '{"hook_event_name":"SessionEnd"}' | ./bin/ccpetline-hook

# Render status line
echo '{}' | ./bin/ccpetline
```

## Configuration

Run `ccpetline-config` to open the TUI configurator. Config is stored in `~/.ccpetline/config.json`.
