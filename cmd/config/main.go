package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jan/claude-pet/internal/pet"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	accentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Italic(true)
	valueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	checkStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	boxStyle    = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)
)

type section int

const (
	sectionMenu        section = iota
	sectionSpecies
	sectionContextMode
	sectionSeparator
	sectionLinesPicker
	sectionLineEdit
	sectionInstall
	sectionDisplayMode
	sectionWrapCommandPicker
	sectionWrapCommandEdit
)

const maxLines = 3

type speciesOption struct {
	species pet.Species
	label   string
	preview string
}

type contextModeOption struct {
	mode  pet.ContextMode
	label string
	desc  string
}

func speciesOptions() []speciesOption {
	var opts []speciesOption
	for _, sp := range pet.AllSpecies {
		emojis := make([]string, 4)
		for i := 0; i < 4; i++ {
			emojis[i] = pet.SizeEmoji(sp, pet.Size(i))
		}
		opts = append(opts, speciesOption{
			species: sp,
			label:   string(sp),
			preview: strings.Join(emojis, "  "),
		})
	}
	return opts
}

func contextModeOptions() []contextModeOption {
	return []contextModeOption{
		{mode: pet.ContextModeCtx, label: "Ctx", desc: "total context window"},
		{mode: pet.ContextModeCtxU, label: "Ctx(u)", desc: "usable context (80% threshold)"},
	}
}

type menuItem struct {
	label   string
	emoji   string
	section section
}

func menuItems() []menuItem {
	return []menuItem{
		{label: "Display Mode", emoji: "\U0001F4FA", section: sectionDisplayMode},
		{},
		{label: "Edit Lines", emoji: "\u270f\ufe0f ", section: sectionLinesPicker},
		{label: "Select Pet", emoji: "\U0001F43E", section: sectionSpecies},
		{label: "Separator", emoji: "\u2702\ufe0f ", section: sectionSeparator},
		{},
		{label: "Context Mode", emoji: "\U0001F4CA", section: sectionContextMode},
		{label: "Install to Claude Code", emoji: "\U0001F527", section: sectionInstall},
	}
}

var tokenEmoji = map[string]string{
	"pet":     "\U0001F43E",
	"mood":    "\U0001F60A",
	"joy":     "\U0001F496",
	"bar":     "\U0001F4CA",
	"model":   "\U0001F916",
	"ctx":     "\U0001F4D0",
	"cost":    "\U0001F4B0",
	"changes": "\U0001F4DD",
	"cwd":     "\U0001F4C2",
	"branch":  "\U0001F33F",
}

type editMode int

const (
	modeList    editMode = iota
	modePicker
	modeCmdEdit
)

type model struct {
	section    section
	options    []speciesOption
	ctxOptions []contextModeOption
	cursor     int
	ctxCursor  int
	menuCursor int

	current        pet.Species
	currentCtxMode pet.ContextMode
	separator      string

	lines       [maxLines][]pet.Segment
	lineFocused int
	segCursor   int
	mode        editMode

	pickerItems   []string
	pickerCursor  int
	pickerInsert  bool
	pickerReplace bool

	editBuf     []rune
	editCursor  int
	editInPlace bool

	displayMode      pet.DisplayMode
	displayCursor    int
	wrapCommand      string
	wrapPickerCursor int

	installStatus string

	quitting bool
}

func buildPickerItems() []string {
	items := make([]string, 0, len(pet.AllTokens)+2)
	for _, t := range pet.AllTokens {
		items = append(items, t)
	}
	items = append(items, "Separator", "Command")
	return items
}

func initialModel() model {
	cfg := pet.LoadConfig()
	opts := speciesOptions()
	ctxOpts := contextModeOptions()

	cursor := 0
	for i, o := range opts {
		if o.species == cfg.Species {
			cursor = i
			break
		}
	}
	ctxCursor := 0
	for i, o := range ctxOpts {
		if o.mode == cfg.ContextMode {
			ctxCursor = i
			break
		}
	}

	var lines [maxLines][]pet.Segment
	for i, tmpl := range cfg.Lines {
		if i >= maxLines {
			break
		}
		lines[i] = pet.TemplateToSegments(tmpl)
	}

	displayCursor := 0
	for i, dm := range pet.AllDisplayModes {
		if dm == cfg.DisplayMode {
			displayCursor = i
			break
		}
	}

	return model{
		section:        sectionMenu,
		options:        opts,
		ctxOptions:     ctxOpts,
		cursor:         cursor,
		ctxCursor:      ctxCursor,
		current:        cfg.Species,
		currentCtxMode: cfg.ContextMode,
		separator:      cfg.Separator,
		lines:          lines,
		pickerItems:    buildPickerItems(),
		displayMode:    cfg.DisplayMode,
		displayCursor:  displayCursor,
		wrapCommand:    cfg.WrapCommand,
	}
}

func (m model) Init() tea.Cmd { return nil }

// --- Update ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
	items := menuItems()
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
		if dest == sectionSeparator {
			m.editBuf = []rune(strings.TrimSpace(m.separator))
			m.editCursor = len(m.editBuf)
		}
		if dest == sectionInstall {
			m.installStatus = installToClaudeCode()
		}
		m.section = dest
	}
	return m, nil
}

func (m model) updateInstall(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.section = sectionMenu
	return m, nil
}

func installToClaudeCode() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")

	var settings map[string]interface{}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Sprintf("Error reading settings: %v", err)
		}
		settings = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Sprintf("Error parsing settings: %v", err)
		}
	}

	want := map[string]interface{}{
		"type":    "command",
		"command": "claude-pet-statusline",
	}

	if existing, ok := settings["statusLine"]; ok {
		if m, ok := existing.(map[string]interface{}); ok {
			if m["type"] == "command" && m["command"] == "claude-pet-statusline" {
				return "Already installed!"
			}
		}
	}

	settings["statusLine"] = want

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error encoding settings: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err)
	}

	if err := os.WriteFile(settingsPath, append(out, '\n'), 0o644); err != nil {
		return fmt.Sprintf("Error writing settings: %v", err)
	}

	return "Installed! Restart Claude Code to activate."
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
		m.save()
		if m.displayMode != pet.ModeStandalone {
			m.wrapPickerCursor = 0
			m.section = sectionWrapCommandPicker
		} else {
			m.section = sectionMenu
		}
	}
	return m, nil
}

const ccstatuslineCmd = "npx -y ccstatusline@latest"

type wrapOption struct {
	label   string
	command string
	custom  bool
}

func wrapCommandOptions() []wrapOption {
	return []wrapOption{
		{label: "ccstatusline", command: ccstatuslineCmd},
		{label: "Custom command", custom: true},
	}
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
			m.save()
			m.section = sectionMenu
		}
	}
	return m, nil
}

func (m model) updateWrapCommandEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.section = sectionMenu
	case "enter":
		m.wrapCommand = string(m.editBuf)
		m.save()
		m.section = sectionMenu
	case "backspace":
		if len(m.editBuf) > 0 && m.editCursor > 0 {
			m.editBuf = append(m.editBuf[:m.editCursor-1], m.editBuf[m.editCursor:]...)
			m.editCursor--
		}
	case "left":
		if m.editCursor > 0 {
			m.editCursor--
		}
	case "right":
		if m.editCursor < len(m.editBuf) {
			m.editCursor++
		}
	default:
		for _, r := range msg.String() {
			if unicode.IsPrint(r) {
				newBuf := make([]rune, len(m.editBuf)+1)
				copy(newBuf, m.editBuf[:m.editCursor])
				newBuf[m.editCursor] = r
				copy(newBuf[m.editCursor+1:], m.editBuf[m.editCursor:])
				m.editBuf = newBuf
				m.editCursor++
			}
		}
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
		m.save()
		m.section = sectionMenu
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
		m.save()
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
		m.save()
		m.section = sectionMenu
	case "backspace":
		if len(m.editBuf) > 0 && m.editCursor > 0 {
			m.editBuf = append(m.editBuf[:m.editCursor-1], m.editBuf[m.editCursor:]...)
			m.editCursor--
		}
	case "left":
		if m.editCursor > 0 {
			m.editCursor--
		}
	case "right":
		if m.editCursor < len(m.editBuf) {
			m.editCursor++
		}
	default:
		for _, r := range msg.String() {
			if unicode.IsPrint(r) && len(m.editBuf) < 3 {
				newBuf := make([]rune, len(m.editBuf)+1)
				copy(newBuf, m.editBuf[:m.editCursor])
				newBuf[m.editCursor] = r
				copy(newBuf[m.editCursor+1:], m.editBuf[m.editCursor:])
				m.editBuf = newBuf
				m.editCursor++
			}
		}
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

func (m *model) clampSegCursor() {
	segs := m.lines[m.lineFocused]
	if len(segs) == 0 {
		m.segCursor = 0
	} else if m.segCursor >= len(segs) {
		m.segCursor = len(segs) - 1
	}
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
			m.clampSegCursor()
			m.save()
		}
	case "c":
		m.lines[m.lineFocused] = nil
		m.segCursor = 0
		m.save()
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
			m.applySegment(seg)
			m.mode = modeList
			m.save()
		case "Command":
			m.mode = modeCmdEdit
			m.editBuf = nil
			m.editCursor = 0
			m.editInPlace = false
		default:
			seg := pet.Segment{Kind: pet.KindToken, Value: item}
			m.applySegment(seg)
			m.mode = modeList
			m.save()
		}
	}
	return m, nil
}

func (m *model) applySegment(seg pet.Segment) {
	segs := m.lines[m.lineFocused]
	if m.pickerReplace {
		if len(segs) > 0 {
			segs[m.segCursor] = seg
			m.lines[m.lineFocused] = segs
		}
	} else if m.pickerInsert {
		newSegs := make([]pet.Segment, 0, len(segs)+1)
		newSegs = append(newSegs, segs[:m.segCursor]...)
		newSegs = append(newSegs, seg)
		newSegs = append(newSegs, segs[m.segCursor:]...)
		m.lines[m.lineFocused] = newSegs
	} else {
		m.lines[m.lineFocused] = append(segs, seg)
		m.segCursor = len(m.lines[m.lineFocused]) - 1
	}
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
			m.applySegment(seg)
		}
		m.mode = modeList
		m.save()
	case "backspace":
		if len(m.editBuf) > 0 && m.editCursor > 0 {
			m.editBuf = append(m.editBuf[:m.editCursor-1], m.editBuf[m.editCursor:]...)
			m.editCursor--
		}
	case "left":
		if m.editCursor > 0 {
			m.editCursor--
		}
	case "right":
		if m.editCursor < len(m.editBuf) {
			m.editCursor++
		}
	default:
		for _, r := range msg.String() {
			if unicode.IsPrint(r) {
				newBuf := make([]rune, len(m.editBuf)+1)
				copy(newBuf, m.editBuf[:m.editCursor])
				newBuf[m.editCursor] = r
				copy(newBuf[m.editCursor+1:], m.editBuf[m.editCursor:])
				m.editBuf = newBuf
				m.editCursor++
			}
		}
	}
	return m, nil
}

func (m *model) save() {
	var lines []string
	for i := 0; i < maxLines; i++ {
		if len(m.lines[i]) > 0 {
			lines = append(lines, pet.SegmentsToTemplate(m.lines[i], m.separator))
		}
	}
	if len(lines) == 0 {
		lines = pet.DefaultLines
	}
	cfg := &pet.Config{
		Species:     m.current,
		ContextMode: m.currentCtxMode,
		Separator:   m.separator,
		Lines:       lines,
		DisplayMode: m.displayMode,
		WrapCommand: m.wrapCommand,
	}
	if err := pet.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
	}
}

// --- Views ---

func (m model) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	switch m.section {
	case sectionMenu:
		m.viewMenu(&b)
	case sectionSpecies:
		m.viewSpecies(&b)
	case sectionContextMode:
		m.viewContextMode(&b)
	case sectionSeparator:
		m.viewSeparator(&b)
	case sectionLinesPicker:
		m.viewLinesPicker(&b)
	case sectionLineEdit:
		m.viewLineEdit(&b)
	case sectionInstall:
		m.viewInstall(&b)
	case sectionDisplayMode:
		m.viewDisplayMode(&b)
	case sectionWrapCommandPicker:
		m.viewWrapCommandPicker(&b)
	case sectionWrapCommandEdit:
		m.viewWrapCommandEdit(&b)
	}
	b.WriteString("\n")
	return b.String()
}

const emojiCol = 3 // fixed visual width for emoji column

// padEmoji pads an emoji string to emojiCol visual width using spaces.
func padEmoji(emoji string) string {
	w := lipgloss.Width(emoji)
	if w >= emojiCol {
		return emoji
	}
	return emoji + strings.Repeat(" ", emojiCol-w)
}

func row(b *strings.Builder, selected bool, emoji, text string) {
	icon := padEmoji(emoji)
	if selected {
		b.WriteString(fmt.Sprintf("  %s %s%s\n", cursorStyle.Render("\u25b8"), icon, accentStyle.Render(text)))
	} else {
		b.WriteString(fmt.Sprintf("    %s%s\n", icon, text))
	}
}

func header(b *strings.Builder, emoji, title string) {
	b.WriteString(fmt.Sprintf("\n  %s%s\n", padEmoji(emoji), titleStyle.Render(title)))
}

func nav(b *strings.Builder, hint string) {
	b.WriteString(hintStyle.Render("      " + hint))
	b.WriteString("\n")
}

func (m model) viewMenu(b *strings.Builder) {
	header(b, "\U0001F9F8", "Claude Pet Config")
	b.WriteString("\n")

	items := menuItems()
	stop := menuStop(items)
	selIdx := 0
	for _, item := range items {
		if item.label == "" {
			b.WriteString("\n")
			continue
		}
		detail := ""
		switch item.section {
		case sectionSpecies:
			detail = string(m.current)
		case sectionContextMode:
			for _, o := range m.ctxOptions {
				if o.mode == m.currentCtxMode {
					detail = o.label
					break
				}
			}
		case sectionSeparator:
			detail = fmt.Sprintf("%q", strings.TrimSpace(m.separator))
		case sectionDisplayMode:
			detail = pet.DisplayModeLabel(m.displayMode)
		}
		text := item.label
		if detail != "" {
			if selIdx == m.menuCursor {
				text += " " + valueStyle.Render(detail)
			} else {
				text += " " + dimStyle.Render(detail)
			}
		}
		row(b, selIdx == m.menuCursor, item.emoji, text)
		selIdx++
	}

	b.WriteString("\n")
	row(b, m.menuCursor == stop, "\U0001F44B", "Exit")
}

func (m model) viewSpecies(b *strings.Builder) {
	header(b, "\U0001F43E", "Select Pet")
	nav(b, "esc back \u00b7 enter select")
	b.WriteString("\n")

	for i, opt := range m.options {
		check := " "
		if opt.species == m.current {
			check = checkStyle.Render("\u2713")
		}
		// Use first emoji of the species as the row icon
		icon := pet.SizeEmoji(opt.species, pet.SizeNormal)
		text := fmt.Sprintf("%s %s  %s", check, opt.label, opt.preview)
		if i != m.cursor {
			text = fmt.Sprintf("%s %s  %s", check, dimStyle.Render(opt.label), dimStyle.Render(opt.preview))
		}
		row(b, i == m.cursor, icon, text)
	}
}

func (m model) viewContextMode(b *strings.Builder) {
	header(b, "\U0001F4CA", "Context Mode")
	nav(b, "esc back \u00b7 enter select")
	b.WriteString("\n")

	for i, opt := range m.ctxOptions {
		check := " "
		if opt.mode == m.currentCtxMode {
			check = checkStyle.Render("\u2713")
		}
		text := fmt.Sprintf("%s %s \u2014 %s", check, opt.label, opt.desc)
		row(b, i == m.ctxCursor, "\U0001F4CA", text)
	}
}

func (m model) viewSeparator(b *strings.Builder) {
	header(b, "\u2702\ufe0f ", "Separator")
	nav(b, "esc back \u00b7 enter save \u00b7 max 3 chars")
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("      %s%s\n",
		accentStyle.Render(string(m.editBuf)),
		cursorStyle.Render("\u2588")))
	preview := " " + strings.TrimSpace(string(m.editBuf)) + " "
	b.WriteString(fmt.Sprintf("      %s %s\n",
		dimStyle.Render("preview:"),
		valueStyle.Render(fmt.Sprintf("{a}%s{b}", preview))))
}

func (m model) viewInstall(b *strings.Builder) {
	header(b, "\U0001F527", "Install to Claude Code")
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("      %s\n", m.installStatus))
	b.WriteString("\n")
	nav(b, "press any key to return")
}

func (m model) viewDisplayMode(b *strings.Builder) {
	header(b, "\U0001F4FA", "Display Mode")
	nav(b, "esc back \u00b7 enter select")
	b.WriteString("\n")

	descs := map[pet.DisplayMode]string{
		pet.ModeStandalone: "pet renders its own status line",
		pet.ModePrepend:    "pet lines above wrapped command",
		pet.ModeAppend:     "pet lines below wrapped command",
	}

	for i, dm := range pet.AllDisplayModes {
		check := " "
		if dm == m.displayMode {
			check = checkStyle.Render("\u2713")
		}
		text := fmt.Sprintf("%s %s \u2014 %s", check, pet.DisplayModeLabel(dm), descs[dm])
		row(b, i == m.displayCursor, "\U0001F4FA", text)
	}
}

func (m model) viewWrapCommandPicker(b *strings.Builder) {
	header(b, "\U0001F527", "Wrap Command")
	nav(b, "esc back \u00b7 enter select")
	b.WriteString("\n")

	opts := wrapCommandOptions()
	for i, opt := range opts {
		check := " "
		if !opt.custom && m.wrapCommand == opt.command {
			check = checkStyle.Render("\u2713")
		}
		text := fmt.Sprintf("%s %s", check, opt.label)
		if !opt.custom {
			detail := dimStyle.Render(opt.command)
			text += " " + detail
		}
		row(b, i == m.wrapPickerCursor, "\U0001F527", text)
	}
}

func (m model) viewWrapCommandEdit(b *strings.Builder) {
	header(b, "\U0001F527", "Wrap Command")
	nav(b, "esc back \u00b7 enter save")
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("      %s%s\n",
		accentStyle.Render(string(m.editBuf)),
		cursorStyle.Render("\u2588")))
	b.WriteString(dimStyle.Render("      command whose stdout is combined with pet lines"))
	b.WriteString("\n")
}

func (m model) viewLinesPicker(b *strings.Builder) {
	header(b, "\u270f\ufe0f ", "Edit Lines")
	nav(b, "esc back \u00b7 enter edit")
	b.WriteString("\n")

	lineEmojis := [maxLines]string{"\u0031\ufe0f\u20e3", "\u0032\ufe0f\u20e3", "\u0033\ufe0f\u20e3"}
	sample := pet.SampleSegmentData(m.current, pet.SizeNormal)
	for i := 0; i < maxLines; i++ {
		preview := dimStyle.Render("(empty)")
		if len(m.lines[i]) > 0 {
			tmpl := pet.SegmentsToTemplate(m.lines[i], m.separator)
			preview = valueStyle.Render(pet.RenderTemplate(tmpl, sample))
		}
		row(b, i == m.lineFocused, lineEmojis[i], preview)
	}
}

func (m model) viewLineEdit(b *strings.Builder) {
	// Preview box
	sample := pet.SampleSegmentData(m.current, pet.SizeNormal)
	var previewLines []string
	for i := 0; i < maxLines; i++ {
		if len(m.lines[i]) == 0 {
			continue
		}
		tmpl := pet.SegmentsToTemplate(m.lines[i], m.separator)
		rendered := pet.RenderTemplate(tmpl, sample)
		if rendered != "" {
			previewLines = append(previewLines, "  "+rendered)
		}
	}
	previewContent := strings.Join(previewLines, "\n")
	if previewContent == "" {
		previewContent = dimStyle.Render("  (empty)")
	}
	b.WriteString("\n")
	b.WriteString(boxStyle.Render("\U0001F441  Preview\n" + previewContent))
	b.WriteString("\n")

	header(b, "\u270f\ufe0f ", fmt.Sprintf("Line %d", m.lineFocused+1))

	// Key hints per mode
	switch m.mode {
	case modeList:
		nav(b, "esc back \u00b7 \u2191\u2193 select \u00b7 \u2190\u2192 change type")
		nav(b, "(a)dd \u00b7 (i)nsert \u00b7 (d)elete \u00b7 (c)lear")
	case modePicker:
		nav(b, "esc back \u00b7 \u2191\u2193 select \u00b7 enter choose")
	case modeCmdEdit:
		nav(b, "esc cancel \u00b7 enter confirm \u00b7 type shell command")
	}
	b.WriteString("\n")

	if m.mode == modePicker {
		m.viewPicker(b)
	} else if m.mode == modeCmdEdit {
		m.viewTextEdit(b)
	} else {
		m.viewSegmentList(b)
	}
}

func (m model) viewSegmentList(b *strings.Builder) {
	segs := m.lines[m.lineFocused]
	if len(segs) == 0 {
		b.WriteString(dimStyle.Render("      (empty \u2014 press 'a' to add)"))
		b.WriteString("\n")
		return
	}
	for i, seg := range segs {
		emoji, label := segmentParts(seg)
		row(b, i == m.segCursor, emoji, label)
	}
}

func segmentParts(seg pet.Segment) (emoji, label string) {
	switch seg.Kind {
	case pet.KindToken:
		emoji = tokenEmoji[seg.Value]
		if emoji == "" {
			emoji = " "
		}
		return emoji, capitalize(seg.Value)
	case pet.KindSeparator:
		return "\u2502", "Separator"
	case pet.KindCommand:
		return "\u26a1", fmt.Sprintf("Command %q", seg.Value)
	default:
		return " ", seg.Value
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func (m model) viewPicker(b *strings.Builder) {
	for i, item := range m.pickerItems {
		emoji := tokenEmoji[item]
		switch item {
		case "Separator":
			emoji = "\u2502"
		case "Command":
			emoji = "\u26a1"
		}
		if emoji == "" {
			emoji = " "
		}
		row(b, i == m.pickerCursor, emoji, capitalize(item))
	}
}

func (m model) viewTextEdit(b *strings.Builder) {
	text := string(m.editBuf)
	b.WriteString(fmt.Sprintf("    %s%s: %s%s\n",
		padEmoji("\u26a1"), "Command",
		accentStyle.Render(text),
		cursorStyle.Render("\u2588")))
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
