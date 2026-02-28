# claude-pet

A terminal goose that lives alongside your Claude Code sessions. It eats tool calls as "snacks" and grows fatter as your context window fills up.

## Quick start

```bash
make install    # builds and copies binaries to ~/go/bin
claude-pet      # run in a separate terminal (or use --tmux)
```

Then add hooks to `~/.claude/settings.json` (see `hooks.example.json`).

## Architecture

```
Claude Code
  |-- PostToolUse hook -----> claude-pet-hook -----> unix socket --> claude-pet (TUI)
  |-- SessionStart hook ----> claude-pet-hook ----/
  |-- SessionEnd hook ------> claude-pet-hook ---/
  |-- Statusline -----------> claude-pet-statusline -> (extracts ctx%, forwards to pet, pipes to real statusline)
```

Three binaries:

| Binary | Purpose |
|--------|---------|
| `claude-pet` | TUI goose. Listens on `/tmp/claude-pet.sock`, renders in terminal. |
| `claude-pet-hook` | Hook handler. Reads Claude Code hook JSON from stdin, sends events to pet. |
| `claude-pet-statusline` | Statusline wrapper. Extracts context %, forwards to pet, pipes to wrapped command. |

## Goose mechanics

| Signal | Source | Effect |
|--------|--------|--------|
| Snack | PostToolUse hook | +1 snack, chomp animation |
| Context update | Statusline wrapper | Updates fatness (primary size driver) |
| Wake | SessionStart hook | Goose wakes up |
| Sleep | SessionEnd hook | Goose goes to sleep |
| Idle | No events for 10s | Gets bored, wanders |

### Size stages (driven by context window %)

1. **Tiny** (0-20%) -- small, energetic
2. **Normal** (21-45%) -- default
3. **Chonky** (46-70%) -- wider body
4. **Mega chonk** (71-90%) -- very wide, sweating
5. **Absolute unit** (91-100%) -- fills the pane

### Snack flavors

- Bash -> spicy taco
- Read -> mild salad
- Edit/Write -> crunchy cookie
- Grep/Glob -> popcorn
- Agent -> whole pizza
- WebFetch/WebSearch -> sushi roll

## Hook config

Add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [{
      "matcher": "*",
      "hooks": [{ "type": "command", "command": "claude-pet-hook", "async": true }]
    }],
    "SessionStart": [{
      "hooks": [{ "type": "command", "command": "claude-pet-hook", "async": true }]
    }],
    "SessionEnd": [{
      "hooks": [{ "type": "command", "command": "claude-pet-hook", "async": true }]
    }]
  },
  "statusLine": {
    "type": "command",
    "command": "claude-pet-statusline npx -y ccstatusline@latest",
    "padding": 0
  }
}
```

If you don't use a statusline, just omit the wrapped command:

```json
"command": "claude-pet-statusline"
```

## Testing manually

```bash
# Terminal 1: run the goose
make run

# Terminal 2: poke it
echo '{"hook_event_name":"SessionStart"}' | ./bin/claude-pet-hook
echo '{"hook_event_name":"PostToolUse","tool_name":"Bash"}' | ./bin/claude-pet-hook
echo '{"hook_event_name":"PostToolUse","tool_name":"Agent"}' | ./bin/claude-pet-hook
echo '{"hook_event_name":"SessionEnd"}' | ./bin/claude-pet-hook
```
