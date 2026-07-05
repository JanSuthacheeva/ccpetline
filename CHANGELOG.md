# Changelog

## 0.0.9

- New: custom hex colors for any segment. The color picker in **Edit Lines** gained a **custom hex** entry (press `#` or select the row below the swatches) accepting any `#rgb` / `#rrggbb` code with a live preview; hex colors render as 24-bit truecolor in both the plain and powerline looks, with the powerline foreground still auto-contrasting
- `line_colors` in `config.json` now accepts hex strings alongside the existing ANSI-256 numbers; old configs load and save unchanged, and invalid color entries are normalized to "no color" instead of failing the load

## 0.0.8

- Fixed: `{changes}` no longer renders a fake `+0/-0` outside git repositories; the segment is simply omitted
- Fixed: state files now use the OS temp directory instead of a hardcoded `/tmp`, so hooks and the statusline work on Windows
- Fixed: self-update validates the downloaded archive before replacing the installed binaries; a corrupt download can no longer leave you with no binaries
- Fixed: concurrent state writes (async hook + statusline) use unique temp files, so a torn state file can no longer be renamed into place
- Fixed: pre-release tags like `0.0.8-rc1` compare correctly in the update check
- Changed: a malformed `config.json` is preserved as `config.json.bad` and defaults are used, instead of the file being silently overwritten on the next save
- Changed: invalid config values (unknown species, bar style, separator style, out-of-range bar width) are normalized to defaults on load, for state files as well
- Changed: the statusline never exits non-zero (it renders from persisted state on bad input) and the async hook always exits 0, reporting problems on stderr only
- Changed: every subprocess the statusline spawns is bounded by a timeout; a slow git can no longer block rendering
- Changed: config save failures are shown inside the TUI instead of being written to stderr where bubbletea makes them invisible
- Removed: the unused Kitty graphics experiment, including the embedded sprite PNGs, shrinking all three binaries
- Internal: the config TUI split into focused files, shared helpers for atomic writes / bar clamping / text editing, a typed Claude payload struct, `make test` / `make lint` targets, and broad test coverage for config loading, migration, state lifecycle, and the settings.json installer

## 0.0.7

- New **Style** screen in `ccpetline-config` that consolidates the look-and-feel choices: a single **Nerd Font** capability toggle gates the icon style (glyphs vs spelled-out text labels), the **Powerline** look, and the powerline separator. Options that require a Nerd Font are hidden entirely when it is off, so you can't pick something your terminal can't render
- First-run wizard: launching `ccpetline-config` with no config yet opens the Style screen with a live preview so you set up your look before seeing the full menu; re-run it anytime from the **Style** menu item
- Moved the **Icons** setting and the **Powerline** / **Separator** toggles (previously under Bar Style) onto the new Style screen; Bar Style now covers just the bar glyphs, pet-in-bar, and width
- New `nerd_font` config field records the capability; configs already using glyphs or the powerline look are recognized as Nerd Font capable so the wizard is skipped, and an explicit `nerd_font: false` coerces the icon theme to text and disables powerline

## 0.0.6

- New **Icons** setting in `ccpetline-config`: choose between the `text` theme (spelled-out labels, the default) and the `nerd` theme (monochrome Nerd Font glyphs)
- Nerd theme prefixes tokens with glyphs (folder, git branch, microchip, gauge, dollar, diff, clock, heart); the pet stays a colorful emoji in both themes
- `DefaultLineColors` / `DefaultTokenColors`: a default color scheme mapping each token to a sensible ANSI color
- New **Powerline** toggle (Bar Style screen): renders each segment as a filled block joined by powerline arrows, using the per-segment colors as backgrounds with auto-contrasting text (needs a Nerd Font)
- Configurable powerline separator via the **Separator** row below the Powerline toggle: Arrow (default), Round, Slant, Backslant, Flame, Pixels, or None for flush blocks with a straight edge and no glyph
- Old configs without the `powerline_sep` field keep the arrow separator
- Shortened the rate-limit reset hint from `(reset in 2h 14m)` to `(2h 14m)`
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
