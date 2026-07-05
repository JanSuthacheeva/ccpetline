package main

import (
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jansuthacheeva/ccpetline/internal/pet"
)

type versionMsg struct {
	latest string
}

type updateDoneMsg struct {
	err error
}

func checkVersionCmd() tea.Msg {
	latest, err := pet.CheckLatestRelease()
	if err != nil || latest == "" {
		return versionMsg{}
	}
	return versionMsg{latest: latest}
}

func (m model) Init() tea.Cmd { return checkVersionCmd }

// --- Update ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case versionMsg:
		if msg.latest != "" {
			m.latestVersion = msg.latest
			m.updateAvailable = true
		}
		return m, nil
	case updateDoneMsg:
		m.updateErr = msg.err != nil
		if msg.err != nil {
			m.updateStatus = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.updateStatus = "Update installed successfully."
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		switch m.section {
		case sectionMenu:
			return m.updateMenu(msg)
		case sectionSpecies:
			return m.updateSpecies(msg)
		case sectionContextMode:
			return m.updateContextMode(msg)
		case sectionSeparator:
			return m.updateSeparator(msg)
		case sectionLinesPicker:
			return m.updateLinesPicker(msg)
		case sectionLineEdit:
			return m.updateLineEdit(msg)
		case sectionInstall:
			return m.updateInstall(msg)
		case sectionDisplayMode:
			return m.updateDisplayMode(msg)
		case sectionWrapCommandPicker:
			return m.updateWrapCommandPicker(msg)
		case sectionWrapCommandEdit:
			return m.updateWrapCommandEdit(msg)
		case sectionColorPicker:
			return m.updateColorPicker(msg)
		case sectionBarStyle:
			return m.updateBarStyle(msg)
		case sectionStyle:
			return m.updateStyle(msg)
		case sectionUpdate:
			return m.updateUpdateResult(msg)
		}
	}
	return m, nil
}

// menuStop returns the total number of selectable positions (items + Exit).
func menuStop(items []menuItem) int {
	n := 0
	for _, item := range items {
		if item.label != "" {
			n++
		}
	}
	return n // Exit is the last position
}

// menuNthSelectable returns the index into items for the nth selectable item.
// Returns -1 if n equals the Exit position.
func menuNthSelectable(items []menuItem, n int) int {
	cur := 0
	for i, item := range items {
		if item.label == "" {
			continue
		}
		if cur == n {
			return i
		}
		cur++
	}
	return -1
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.menuItems()
	stop := menuStop(items)
	switch msg.String() {
	case "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < stop {
			m.menuCursor++
		}
	case "enter":
		if m.menuCursor == stop {
			m.quitting = true
			return m, tea.Quit
		}
		idx := menuNthSelectable(items, m.menuCursor)
		if idx < 0 {
			break
		}
		dest := items[idx].section
		if dest == sectionUpdate {
			m.updateStatus = ""
			m.updateErr = false
			m.section = sectionUpdate
			tag := m.latestVersion
			return m, func() tea.Msg {
				return updateDoneMsg{err: pet.SelfUpdate(tag)}
			}
		}
		if dest == sectionSeparator {
			m.editBuf = []rune(strings.TrimSpace(m.separator))
			m.editCursor = len(m.editBuf)
		}
		if dest == sectionInstall {
			if err := pet.InstallToClaudeCode(); err != nil {
				m.installStatus = fmt.Sprintf("Error: %v", err)
			} else {
				m.installStatus = "Installed! Restart Claude Code to activate."
			}
		}
		m.section = dest
	}
	return m, nil
}

func (m model) updateInstall(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.section = sectionMenu
	return m, nil
}

func (m model) updateUpdateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.updateWaitKey {
		m.updateWaitKey = true
		return m, nil
	}
	if m.updateErr {
		m.section = sectionMenu
		return m, nil
	}
	m.quitting = true
	return m, tea.Quit
}

func (m model) updateDisplayMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	modes := pet.AllDisplayModes
	switch msg.String() {
	case "esc":
		m.section = sectionMenu
	case "up", "k":
		if m.displayCursor > 0 {
			m.displayCursor--
		}
	case "down", "j":
		if m.displayCursor < len(modes)-1 {
			m.displayCursor++
		}
	case "enter":
		m.displayMode = modes[m.displayCursor]
		m = m.save()
		if m.displayMode != pet.ModeStandalone {
			m.wrapPickerCursor = 0
			m.section = sectionWrapCommandPicker
		} else {
			m.section = sectionMenu
		}
	}
	return m, nil
}

func (m model) updateWrapCommandPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	opts := wrapCommandOptions()
	switch msg.String() {
	case "esc":
		m.section = sectionMenu
	case "up", "k":
		if m.wrapPickerCursor > 0 {
			m.wrapPickerCursor--
		}
	case "down", "j":
		if m.wrapPickerCursor < len(opts)-1 {
			m.wrapPickerCursor++
		}
	case "enter":
		opt := opts[m.wrapPickerCursor]
		if opt.custom {
			m.editBuf = []rune(m.wrapCommand)
			m.editCursor = len(m.editBuf)
			m.section = sectionWrapCommandEdit
		} else {
			m.wrapCommand = opt.command
			m = m.save()
			m.section = sectionMenu
		}
	}
	return m, nil
}

// applyEditKey applies a text-editing key (backspace, cursor movement,
// printable rune insertion) to a rune buffer with its cursor position.
// maxLen 0 means unlimited. All TUI line editors share this handler.
func applyEditKey(buf []rune, cursor int, msg tea.KeyMsg, maxLen int) ([]rune, int) {
	switch msg.String() {
	case "backspace":
		if len(buf) > 0 && cursor > 0 {
			buf = append(buf[:cursor-1], buf[cursor:]...)
			cursor--
		}
	case "left":
		if cursor > 0 {
			cursor--
		}
	case "right":
		if cursor < len(buf) {
			cursor++
		}
	default:
		for _, r := range msg.String() {
			if unicode.IsPrint(r) && (maxLen == 0 || len(buf) < maxLen) {
				newBuf := make([]rune, len(buf)+1)
				copy(newBuf, buf[:cursor])
				newBuf[cursor] = r
				copy(newBuf[cursor+1:], buf[cursor:])
				buf = newBuf
				cursor++
			}
		}
	}
	return buf, cursor
}

func (m model) updateWrapCommandEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.section = sectionMenu
	case "enter":
		m.wrapCommand = string(m.editBuf)
		m = m.save()
		m.section = sectionMenu
	default:
		m.editBuf, m.editCursor = applyEditKey(m.editBuf, m.editCursor, msg, 0)
	}
	return m, nil
}

func (m model) updateSpecies(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.section = sectionMenu
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.options)-1 {
			m.cursor++
		}
	case "enter":
		m.current = m.options[m.cursor].species
		m = m.save()
		m.section = sectionMenu
	}
	return m, nil
}

// Style screen row indices. Rows are laid out top-to-bottom in this fixed
// order; Icons/Powerline exist only when Nerd Font is on, and Separator only
// when Powerline is also on, so a visible cursor index maps directly to its
// constant.
func (m model) styleRowCount() int {
	if !m.nerdFont {
		return 1
	}
	if m.powerline {
		return 4
	}
	return 3
}

func (m model) clampStyleCursor() model {
	if rows := m.styleRowCount(); m.styleCursor >= rows {
		m.styleCursor = rows - 1
	}
	return m
}

// styleAdjust toggles or cycles the focused Style row, then persists.
func (m model) styleAdjust(delta int) model {
	switch m.styleCursor {
	case styleRowNerdFont:
		m.nerdFont = !m.nerdFont
		if m.nerdFont {
			// Default to glyphs when the capability is first enabled.
			m.iconTheme = pet.IconThemeNerd
		} else {
			m.iconTheme = pet.IconThemeText
			m.powerline = false
		}
		m = m.clampStyleCursor()
	case styleRowIcons:
		if m.iconTheme == pet.IconThemeNerd {
			m.iconTheme = pet.IconThemeText
		} else {
			m.iconTheme = pet.IconThemeNerd
		}
	case styleRowPowerline:
		m.powerline = !m.powerline
		m = m.clampStyleCursor()
	case styleRowSeparator:
		m.powerlineSep = cyclePowerlineSep(m.powerlineSep, delta)
	}
	return m.save()
}

func (m model) updateStyle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.styleRowCount()
	switch msg.String() {
	case "esc":
		// During the first-run wizard there is no menu to return to; enter
		// completes setup instead.
		if !m.firstRun {
			m.section = sectionMenu
		}
	case "enter":
		m = m.save()
		m.firstRun = false
		m.section = sectionMenu
	case "up", "k":
		if m.styleCursor > 0 {
			m.styleCursor--
		}
	case "down", "j":
		if m.styleCursor < rows-1 {
			m.styleCursor++
		}
	case "left", "h":
		m = m.styleAdjust(-1)
	case "right", "l", " ":
		m = m.styleAdjust(1)
	}
	return m, nil
}

func (m model) updateContextMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.section = sectionMenu
	case "up", "k":
		if m.ctxCursor > 0 {
			m.ctxCursor--
		}
	case "down", "j":
		if m.ctxCursor < len(m.ctxOptions)-1 {
			m.ctxCursor++
		}
	case "enter":
		m.currentCtxMode = m.ctxOptions[m.ctxCursor].mode
		m = m.save()
		m.section = sectionMenu
	}
	return m, nil
}

func (m model) updateSeparator(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.section = sectionMenu
	case "enter":
		sep := strings.TrimSpace(string(m.editBuf))
		if sep == "" {
			sep = "|"
		}
		m.separator = " " + sep + " "
		m = m.save()
		m.section = sectionMenu
	default:
		m.editBuf, m.editCursor = applyEditKey(m.editBuf, m.editCursor, msg, 3)
	}
	return m, nil
}

func (m model) updateLinesPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.section = sectionMenu
	case "up", "k":
		if m.lineFocused > 0 {
			m.lineFocused--
		}
	case "down", "j":
		if m.lineFocused < maxLines-1 {
			m.lineFocused++
		}
	case "enter":
		m.segCursor = 0
		m.mode = modeList
		m.section = sectionLineEdit
	}
	return m, nil
}

func (m model) updateLineEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeList:
		return m.updateList(msg)
	case modePicker:
		return m.updatePicker(msg)
	case modeCmdEdit:
		return m.updateTextEdit(msg)
	}
	return m, nil
}

func (m model) currentSegments() []pet.Segment {
	return m.lines[m.lineFocused]
}

func (m model) clampSegCursor() model {
	segs := m.lines[m.lineFocused]
	if len(segs) == 0 {
		m.segCursor = 0
	} else if m.segCursor >= len(segs) {
		m.segCursor = len(segs) - 1
	}
	return m
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	segs := m.currentSegments()
	switch msg.String() {
	case "esc":
		m.section = sectionLinesPicker
	case "up", "k":
		if m.segCursor > 0 {
			m.segCursor--
		}
	case "down", "j":
		if m.segCursor < len(segs)-1 {
			m.segCursor++
		}
	case "f":
		if len(segs) > 0 {
			// Find current color index in palette.
			m.colorCursor = 0
			colors := m.lineColors[m.lineFocused]
			if m.segCursor < len(colors) {
				cur := colors[m.segCursor]
				for i, c := range colorPalette {
					if c == cur {
						m.colorCursor = i
						break
					}
				}
			}
			m.section = sectionColorPicker
		}
	case "a":
		m.mode = modePicker
		m.pickerCursor = 0
		m.pickerInsert = false
		m.pickerReplace = false
	case "i":
		m.mode = modePicker
		m.pickerCursor = 0
		m.pickerInsert = true
		m.pickerReplace = false
	case "d":
		if len(segs) > 0 {
			newSegs := make([]pet.Segment, 0, len(segs)-1)
			newSegs = append(newSegs, segs[:m.segCursor]...)
			newSegs = append(newSegs, segs[m.segCursor+1:]...)
			m.lines[m.lineFocused] = newSegs
			// Remove color at index too.
			colors := m.lineColors[m.lineFocused]
			if m.segCursor < len(colors) {
				newColors := make([]uint8, 0, len(colors)-1)
				newColors = append(newColors, colors[:m.segCursor]...)
				newColors = append(newColors, colors[m.segCursor+1:]...)
				m.lineColors[m.lineFocused] = newColors
			}
			m = m.clampSegCursor()
			m = m.save()
		}
	case "c":
		m.lines[m.lineFocused] = nil
		m.lineColors[m.lineFocused] = nil
		m.segCursor = 0
		m = m.save()
	case "enter", "left", "right":
		if len(segs) > 0 {
			m.mode = modePicker
			m.pickerCursor = 0
			m.pickerInsert = false
			m.pickerReplace = true
		}
	}
	return m, nil
}

func (m model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeList
	case "up", "k":
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}
	case "down", "j":
		if m.pickerCursor < len(m.pickerItems)-1 {
			m.pickerCursor++
		}
	case "enter":
		item := m.pickerItems[m.pickerCursor]
		switch item {
		case "Separator":
			seg := pet.Segment{Kind: pet.KindSeparator}
			m = m.applySegment(seg)
			m.mode = modeList
			m = m.save()
		case "Command":
			m.mode = modeCmdEdit
			m.editBuf = nil
			m.editCursor = 0
			m.editInPlace = false
		default:
			seg := pet.Segment{Kind: pet.KindToken, Value: item}
			m = m.applySegment(seg)
			m.mode = modeList
			m = m.save()
		}
	}
	return m, nil
}

func (m model) applySegment(seg pet.Segment) model {
	segs := m.lines[m.lineFocused]
	colors := m.lineColors[m.lineFocused]
	if m.pickerReplace {
		if len(segs) > 0 {
			segs[m.segCursor] = seg
			m.lines[m.lineFocused] = segs
			// Keep existing color on replace.
		}
	} else if m.pickerInsert {
		newSegs := make([]pet.Segment, 0, len(segs)+1)
		newSegs = append(newSegs, segs[:m.segCursor]...)
		newSegs = append(newSegs, seg)
		newSegs = append(newSegs, segs[m.segCursor:]...)
		m.lines[m.lineFocused] = newSegs
		// Insert 0 color at index.
		newColors := make([]uint8, 0, len(colors)+1)
		newColors = append(newColors, colors[:min(m.segCursor, len(colors))]...)
		newColors = append(newColors, 0)
		if m.segCursor < len(colors) {
			newColors = append(newColors, colors[m.segCursor:]...)
		}
		m.lineColors[m.lineFocused] = newColors
	} else {
		m.lines[m.lineFocused] = append(segs, seg)
		m.segCursor = len(m.lines[m.lineFocused]) - 1
		// Append 0 color.
		m.lineColors[m.lineFocused] = append(colors, 0)
	}
	return m
}

func (m model) updateTextEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeList
	case "enter":
		text := string(m.editBuf)
		seg := pet.Segment{Kind: pet.KindCommand, Value: text}
		if m.editInPlace {
			segs := m.lines[m.lineFocused]
			if len(segs) > 0 && m.segCursor < len(segs) {
				segs[m.segCursor] = seg
			}
		} else {
			m = m.applySegment(seg)
		}
		m.mode = modeList
		m = m.save()
	default:
		m.editBuf, m.editCursor = applyEditKey(m.editBuf, m.editCursor, msg, 0)
	}
	return m, nil
}

func (m model) updateColorPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.section = sectionLineEdit
		m.mode = modeList
	case "up", "k":
		if m.colorCursor >= colorsPerRow {
			m.colorCursor -= colorsPerRow
		}
	case "down", "j":
		if m.colorCursor+colorsPerRow < len(colorPalette) {
			m.colorCursor += colorsPerRow
		}
	case "left", "h":
		if m.colorCursor > 0 {
			m.colorCursor--
		}
	case "right", "l":
		if m.colorCursor < len(colorPalette)-1 {
			m.colorCursor++
		}
	case "enter":
		color := colorPalette[m.colorCursor]
		// Ensure lineColors slice is large enough.
		colors := m.lineColors[m.lineFocused]
		for len(colors) <= m.segCursor {
			colors = append(colors, 0)
		}
		colors[m.segCursor] = color
		m.lineColors[m.lineFocused] = colors
		m = m.save()
		m.section = sectionLineEdit
		m.mode = modeList
	}
	return m, nil
}

// barStyleRows: 0..len(AllBarStyles)-1 = styles, then pet toggle and width.
// The Nerd Font / Powerline choices live on the Style screen.

func cyclePowerlineSep(cur pet.PowerlineSepStyle, delta int) pet.PowerlineSepStyle {
	styles := pet.AllPowerlineSepStyles
	idx := 0
	for i, s := range styles {
		if s == cur {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(styles)) % len(styles)
	return styles[idx]
}

func (m model) updateBarStyle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxRow := barRowWidth
	switch msg.String() {
	case "esc":
		m.section = sectionMenu
	case "up", "k":
		if m.barStyleCursor > 0 {
			m.barStyleCursor--
		}
	case "down", "j":
		if m.barStyleCursor < maxRow {
			m.barStyleCursor++
		}
	case "enter":
		if m.barStyleCursor < len(pet.AllBarStyles) {
			m.barStyle = pet.AllBarStyles[m.barStyleCursor]
			m = m.save()
		} else if m.barStyleCursor == barRowPetToggle {
			m.barShowPet = !m.barShowPet
			m = m.save()
		}
	case "left", "h", "-":
		if m.barStyleCursor == barRowWidth && m.barWidth > pet.MinBarWidth {
			m.barWidth--
			m = m.save()
		}
	case "right", "l", "+", "=":
		if m.barStyleCursor == barRowWidth && m.barWidth < pet.MaxBarWidth {
			m.barWidth++
			m = m.save()
		}
	}
	return m, nil
}

func (m model) save() model {
	var lines []string
	for i := 0; i < maxLines; i++ {
		if len(m.lines[i]) > 0 {
			lines = append(lines, pet.SegmentsToTemplate(m.lines[i], m.separator))
		}
	}
	if len(lines) == 0 {
		lines = pet.DefaultLines
	}
	// Collect line colors, omitting trailing all-zero slices.
	var lc [][]uint8
	for i := 0; i < maxLines; i++ {
		if len(m.lines[i]) > 0 {
			lc = append(lc, m.lineColors[i])
		}
	}
	// Trim trailing empty color slices.
	for len(lc) > 0 {
		last := lc[len(lc)-1]
		allZero := true
		for _, c := range last {
			if c != 0 {
				allZero = false
				break
			}
		}
		if allZero || len(last) == 0 {
			lc = lc[:len(lc)-1]
		} else {
			break
		}
	}
	// Without a Nerd Font, glyph icons and the powerline look cannot render;
	// keep the persisted config consistent with the declared capability.
	iconTheme := m.iconTheme
	powerline := m.powerline
	if !m.nerdFont {
		iconTheme = pet.IconThemeText
		powerline = false
	}
	barShowPet := &m.barShowPet
	cfg := &pet.Config{
		Species:      m.current,
		ContextMode:  m.currentCtxMode,
		NerdFont:     m.nerdFont,
		IconTheme:    iconTheme,
		Separator:    m.separator,
		Lines:        lines,
		LineColors:   lc,
		DisplayMode:  m.displayMode,
		WrapCommand:  m.wrapCommand,
		BarStyle:     m.barStyle,
		BarShowPet:   barShowPet,
		BarWidth:     m.barWidth,
		Powerline:    powerline,
		PowerlineSep: m.powerlineSep,
	}
	// bubbletea owns the terminal, so stderr would be invisible or garbled;
	// surface persistence failures in the view instead.
	if err := pet.SaveConfig(cfg); err != nil {
		m.saveErr = fmt.Sprintf("Error saving config: %v", err)
	} else {
		m.saveErr = ""
	}
	return m
}

// --- Views ---
