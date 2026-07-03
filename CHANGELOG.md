# Changelog

## 0.0.6

- New **Icons** setting in `ccpetline-config`: choose between the `text` theme (spelled-out labels, the default) and the `nerd` theme (monochrome Nerd Font glyphs)
- Nerd theme prefixes tokens with glyphs (folder, git branch, microchip, gauge, dollar, diff, clock, heart); the pet stays a colorful emoji in both themes
- `DefaultLineColors` / `DefaultTokenColors`: a default color scheme mapping each token to a sensible ANSI color
- Fixed the branch marker in the default text theme: `⌥` (macOS Option key) replaced with `⎇`
- Fixed bar width to measure display cells instead of bytes, so the progress bar no longer overruns its configured width when the pet or suffix contains wide/multi-byte characters
- Nerd theme requires a Nerd Font installed; falls back to the text theme otherwise

## 0.0.5

- New tokens `{5h}` and `{7d}` showing Claude subscription rate limit usage with reset countdown, e.g. `5h: 13% (reset in 2h 14m)`
- New tokens `{5h_bar}` and `{7d_bar}` rendering the limits as progress bars using the configured bar style and width
- Tokens render empty until Claude Code provides `rate_limits` (first API response of the session)
- Renamed `{bar}` to `{ctx_bar}`; `{bar}` still works in existing configs

## 0.0.4

- Use AdaptiveColor for light/dark terminal support in TUI config
- Show updating overlay while download runs in background
- Show update success screen instead of auto-closing TUI
- Recommend Noto Color Emoji as fallback font in README

## 0.0.3

- Show version number in TUI config heading
- Fix keycap emoji rendering on macOS (replaced with L1/L2/L3)
- Improve legend/hint text visibility (white instead of dark gray)

## 0.0.2

- Self-update from TUI: new "Update to vX.Y.Z" menu item when an update is available
- Downloads and replaces binaries in-place, prompts to restart
- Changelog link shown in menu and update result screen
- Supports tar.gz (Linux/macOS) and zip (Windows) release archives

## 0.0.1

Initial release.

- Terminal pet that lives in the Claude Code status line
- Eats tool calls as snacks, grows fatter with context window usage
- Four size stages: tiny, normal, chonky, mega chonk
- Mood system: happy, bored, sleeping
- TUI configurator (`ccpetline-config`)
- Configurable bar style, pet visibility, and width
- Cross-platform install script
