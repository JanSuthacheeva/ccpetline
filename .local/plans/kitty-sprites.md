# Kitty Graphics Protocol Support for Pet Sprites

## Context

The pet currently renders as emoji characters, which are locked to terminal cell size. Ghostty (and kitty, WezTerm) support the Kitty graphics protocol, which can display inline PNG images at arbitrary pixel sizes. This lets the pet have real pixel art sprites that vary in size.

## Approach

Embed pre-made PNG sprites in the binary via `go:embed`. Detect Kitty protocol support via environment variables (no terminal query -- stdin is piped JSON, not a tty). Fall back to current emoji rendering on unsupported terminals. Zero new dependencies.

## Sprite Matrix

5 sizes x 3 moods = **15 PNGs**. Snack types stay as emoji appended after the sprite. Target size: ~32x32px (fits in 1-2 terminal rows).

```
internal/pet/sprites/
  tiny_eating.png      tiny_bored.png      tiny_sleeping.png
  normal_eating.png    normal_bored.png    normal_sleeping.png
  chonky_eating.png    chonky_bored.png    chonky_sleeping.png
  megachonk_eating.png megachonk_bored.png megachonk_sleeping.png
  unit_eating.png      unit_bored.png      unit_sleeping.png
```

## New Files (all in `internal/pet/`)

| File | Purpose |
|------|---------|
| `sprites/*.png` | 15 embedded PNG sprite assets |
| `sprites.go` | `go:embed` directive, `SpritePNG(size, mood) []byte` lookup |
| `kitty.go` | `KittyInlinePNG(png) string` -- base64 encode, chunk at 4096 bytes, wrap in `ESC_G...ESC\` escape sequences. Uses `q=2` to suppress terminal response. |
| `kitty_detect.go` | `KittySupported() bool` -- env-based detection (see below) |
| `render_kitty.go` | `FormatPetLineKitty(state)`, `FormatSeparatorKitty(state, width)` -- kitty-aware variants that fall back to emoji if sprite not found |

## Detection (`kitty_detect.go`)

No terminal query (stdin is piped). Priority:
1. `CLAUDE_PET_KITTY=1|0` -- explicit user override
2. `KITTY_PID` or `KITTY_WINDOW_ID` env vars (set by kitty)
3. `TERM_PROGRAM` in `{kitty, ghostty, WezTerm}`
4. Default: false (emoji fallback)

## Changes to Existing Files

**`cmd/statusline/main.go`** -- ~6 lines changed: branch on `KittySupported()` to call kitty or emoji format functions.

No other existing files modified. `render.go` stays as-is (fallback path).

## Implementation Order

1. Add PNGs to `internal/pet/sprites/` (prerequisite -- need art first)
2. `sprites.go` -- embed + lookup
3. `kitty.go` -- protocol implementation + tests
4. `kitty_detect.go` -- env detection + tests
5. `render_kitty.go` -- kitty rendering functions
6. `cmd/statusline/main.go` -- add detection branch
7. Visual testing in ghostty/kitty and a non-supporting terminal

## Risks

- **Separator line sprite alignment**: Kitty images consume variable cell widths. V1: just drop the sprite in, tune later.
- **Binary size**: 15 PNGs at 32x32 ~= 15-30 KB. Negligible.

## Verification

1. `make build` succeeds
2. `CLAUDE_PET_KITTY=0 echo '{}' | bin/claude-pet-statusline` shows emoji (fallback)
3. `CLAUDE_PET_KITTY=1 echo '{}' | bin/claude-pet-statusline` shows escape sequences (pipe to `cat -v` to verify)
4. Visual test in ghostty: sprite renders inline at correct size
5. `make test` passes
