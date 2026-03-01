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

	"github.com/jansuthacheeva/ccpetline/internal/pet"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	accentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Italic(true)
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
	sectionColorPicker
	sectionBarStyle
	sectionUpdate
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

func (m model) menuItems() []menuItem {
	var items []menuItem
	if m.updateAvailable {
		items = append(items, menuItem{
			label:   fmt.Sprintf("Update to %s", m.latestVersion),
			emoji:   "\U0001F680",
			section: sectionUpdate,
		})
		items = append(items, menuItem{}) // separator
	}
	items = append(items,
		menuItem{label: "Display Mode", emoji: "\U0001F4FA", section: sectionDisplayMode},
		menuItem{},
		menuItem{label: "Edit Lines", emoji: "\u270f\ufe0f ", section: sectionLinesPicker},
		menuItem{label: "Select Pet", emoji: "\U0001F43E", section: sectionSpecies},
		menuItem{label: "Separator", emoji: "\u2702\ufe0f ", section: sectionSeparator},
		menuItem{},
		menuItem{label: "Bar Style", emoji: "\U0001F4CA", section: sectionBarStyle},
		menuItem{label: "Context Mode", emoji: "\U0001F4CA", section: sectionContextMode},
		menuItem{label: "Install to Claude Code", emoji: "\U0001F527", section: sectionInstall},
	)
	return items
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
	"dir":     "\U0001F4C1",
	"branch":  "\U0001F33F",
}

// colorPalette is the curated set of ANSI 256 colors available in the picker.
var colorPalette = []uint8{
	0, 196, 208, 220, 226,
	118, 49, 51, 45, 39,
	63, 99, 141, 177,
	212, 255, 245,
}

const colorsPerRow = 6

// colorLabel returns a human-readable name for a palette color.
func colorLabel(c uint8) string {
	switch c {
	case 0:
		return "none"
	case 196:
		return "red"
	case 208:
		return "orange"
	case 220:
		return "gold"
	case 226:
		return "yellow"
	case 118:
		return "green"
	case 49:
		return "emerald"
	case 51:
		return "cyan"
	case 45:
		return "sky"
	case 39:
		return "blue"
	case 63:
		return "indigo"
	case 99:
		return "purple"
	case 141:
		return "lavender"
	case 177:
		return "orchid"
	case 212:
		return "pink"
	case 255:
		return "white"
	case 245:
		return "gray"
	default:
		return fmt.Sprintf("%d", c)
	}
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
	lineColors  [maxLines][]uint8
	lineFocused int
	segCursor   int
	colorCursor int
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

	barStyle       pet.BarStyle
	barShowPet     bool
	barWidth       int
	barStyleCursor int

	installStatus string

	latestVersion   string
	updateAvailable bool
	updateStatus    string

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
	var lineColors [maxLines][]uint8
	for i, tmpl := range cfg.Lines {
		if i >= maxLines {
			break
		}
		lines[i] = pet.TemplateToSegments(tmpl)
	}
	for i, colors := range cfg.LineColors {
		if i >= maxLines {
			break
		}
		lineColors[i] = make([]uint8, len(colors))
		copy(lineColors[i], colors)
	}

	displayCursor := 0
	for i, dm := range pet.AllDisplayModes {
		if dm == cfg.DisplayMode {
			displayCursor = i
			break
		}
	}

	barShowPet := true
	if cfg.BarShowPet != nil {
		barShowPet = *cfg.BarShowPet
	}
	barStyle := cfg.BarStyle
	if barStyle == "" {
		barStyle = pet.BarClassic
	}
	barStyleCursor := 0
	for i, s := range pet.AllBarStyles {
		if s == barStyle {
			barStyleCursor = i
			break
		}
	}
	barWidth := cfg.BarWidth
	if barWidth < 20 || barWidth > 80 {
		barWidth = 50
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
		lineColors:     lineColors,
		pickerItems:    buildPickerItems(),
		displayMode:    cfg.DisplayMode,
		displayCursor:  displayCursor,
		wrapCommand:    cfg.WrapCommand,
		barStyle:       barStyle,
		barShowPet:     barShowPet,
		barWidth:       barWidth,
		barStyleCursor: barStyleCursor,
	}
}

type versionMsg struct {
	latest string
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
			m.section = sectionUpdate
			if err := pet.SelfUpdate(m.latestVersion); err != nil {
				m.updateStatus = fmt.Sprintf("Error: %v", err)
			} else {
				m.updateStatus = "Updated successfully!"
			}
			return m, nil
		}
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

func (m model) updateUpdateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if strings.HasPrefix(m.updateStatus, "Error:") {
		m.section = sectionMenu
		return m, nil
	}
	m.quitting = true
	return m, tea.Quit
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

	// Status line
	settings["statusLine"] = map[string]interface{}{
		"type":    "command",
		"command": "ccpetline",
	}

	// Hooks — append to existing entries, skip if ccpetline-hook already present
	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	petHookEntry := map[string]interface{}{"type": "command", "command": "ccpetline-hook", "async": true}

	// PostToolUse needs a matcher
	appendHookEntry(hooks, "PostToolUse", map[string]interface{}{
		"matcher": "*",
		"hooks":   []interface{}{petHookEntry},
	})
	// SessionStart / SessionEnd have no matcher
	simpleEntry := map[string]interface{}{
		"hooks": []interface{}{petHookEntry},
	}
	appendHookEntry(hooks, "SessionStart", simpleEntry)
	appendHookEntry(hooks, "SessionEnd", simpleEntry)

	settings["hooks"] = hooks

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

// hookEntryHasPetline checks if a hook entry already references ccpetline-hook.
func hookEntryHasPetline(entry interface{}) bool {
	m, ok := entry.(map[string]interface{})
	if !ok {
		return false
	}
	cmds, ok := m["hooks"].([]interface{})
	if !ok {
		return false
	}
	for _, cmd := range cmds {
		if h, ok := cmd.(map[string]interface{}); ok {
			if h["command"] == "ccpetline-hook" {
				return true
			}
		}
	}
	return false
}

// appendHookEntry appends a hook entry to the given event key, unless
// ccpetline-hook is already present in any existing entry.
func appendHookEntry(hooks map[string]interface{}, event string, entry map[string]interface{}) {
	existing, _ := hooks[event].([]interface{})
	for _, e := range existing {
		if hookEntryHasPetline(e) {
			return
		}
	}
	hooks[event] = append(existing, entry)
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
			m.clampSegCursor()
			m.save()
		}
	case "c":
		m.lines[m.lineFocused] = nil
		m.lineColors[m.lineFocused] = nil
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
		m.save()
		m.section = sectionLineEdit
		m.mode = modeList
	}
	return m, nil
}

// barStyleRows: 0..len(AllBarStyles)-1 = styles, next = pet toggle, next = width
const barRowPetToggle = 4 // len(AllBarStyles)
const barRowWidth = 5

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
			m.save()
		} else if m.barStyleCursor == barRowPetToggle {
			m.barShowPet = !m.barShowPet
			m.save()
		}
	case "left", "h":
		if m.barStyleCursor == barRowWidth && m.barWidth > 20 {
			m.barWidth--
			m.save()
		}
	case "right", "l":
		if m.barStyleCursor == barRowWidth && m.barWidth < 80 {
			m.barWidth++
			m.save()
		}
	case "-":
		if m.barStyleCursor == barRowWidth && m.barWidth > 20 {
			m.barWidth--
			m.save()
		}
	case "+", "=":
		if m.barStyleCursor == barRowWidth && m.barWidth < 80 {
			m.barWidth++
			m.save()
		}
	}
	return m, nil
}

func (m model) viewBarStyle(b *strings.Builder) {
	header(b, "\U0001F4CA", "Bar Style")
	nav(b, "esc back \u00b7 enter select \u00b7 \u2190\u2192 adjust width")
	b.WriteString("\n")

	for i, style := range pet.AllBarStyles {
		check := " "
		if style == m.barStyle {
			check = checkStyle.Render("\u2713")
		}
		// Build a short preview bar for this style.
		preview := m.barPreview(style, m.barShowPet)
		text := fmt.Sprintf("%s %s  %s", check, pet.BarStyleLabel(style), dimStyle.Render(preview))
		row(b, i == m.barStyleCursor, "\U0001F4CA", text)
	}

	b.WriteString("\n")

	// Pet toggle row
	petLabel := "Show pet in bar"
	petVal := "off"
	if m.barShowPet {
		petVal = "on"
	}
	petText := fmt.Sprintf("%s %s", petLabel, valueStyle.Render(petVal))
	row(b, m.barStyleCursor == barRowPetToggle, "\U0001F43E", petText)

	// Width row
	widthText := fmt.Sprintf("Bar width %s", valueStyle.Render(fmt.Sprintf("%d", m.barWidth)))
	row(b, m.barStyleCursor == barRowWidth, "\u2194\ufe0f", widthText)
}

func (m model) barPreview(style pet.BarStyle, showPet bool) string {
	s := &pet.State{
		Species:    m.current,
		Size:       pet.SizeNormal,
		ContextPct: 53.1,
		BarStyle:   style,
		BarShowPet: showPet,
		BarWidth:   m.barWidth,
	}
	return pet.FormatSeparator(s)
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
	barShowPet := &m.barShowPet
	cfg := &pet.Config{
		Species:     m.current,
		ContextMode: m.currentCtxMode,
		Separator:   m.separator,
		Lines:       lines,
		LineColors:  lc,
		DisplayMode: m.displayMode,
		WrapCommand: m.wrapCommand,
		BarStyle:    m.barStyle,
		BarShowPet:  barShowPet,
		BarWidth:    m.barWidth,
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
	case sectionColorPicker:
		m.viewColorPicker(&b)
	case sectionBarStyle:
		m.viewBarStyle(&b)
	case sectionUpdate:
		m.viewUpdate(&b)
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
	header(b, "\U0001F9F8", fmt.Sprintf("ccpetline config v%s", pet.Version))
	if m.updateAvailable {
		b.WriteString(accentStyle.Render(fmt.Sprintf("      Update available: %s (current: v%s)", m.latestVersion, pet.Version)))
		b.WriteString("\n")
		changelogURL := fmt.Sprintf("https://github.com/jansuthacheeva/ccpetline/releases/tag/%s", m.latestVersion)
		b.WriteString(dimStyle.Render(fmt.Sprintf("      %s", changelogURL)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	items := m.menuItems()
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
		case sectionBarStyle:
			detail = pet.BarStyleLabel(m.barStyle)
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

func (m model) viewUpdate(b *strings.Builder) {
	header(b, "\U0001F680", fmt.Sprintf("Update to %s", m.latestVersion))
	b.WriteString("\n")
	if m.updateStatus == "" {
		b.WriteString("      Updating...\n")
	} else {
		b.WriteString(fmt.Sprintf("      %s\n", m.updateStatus))
	}
	b.WriteString("\n")
	changelogURL := fmt.Sprintf("https://github.com/jansuthacheeva/ccpetline/releases/tag/%s", m.latestVersion)
	b.WriteString(fmt.Sprintf("      Changelog: %s\n", dimStyle.Render(changelogURL)))
	b.WriteString("\n")
	if strings.HasPrefix(m.updateStatus, "Error:") {
		nav(b, "press any key to return")
	} else {
		b.WriteString("      Please restart ccpetline-config to use the new version.\n")
		b.WriteString("\n")
		nav(b, "press any key to quit")
	}
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

	lineEmojis := [maxLines]string{"L1", "L2", "L3"}
	sample := pet.SampleSegmentData(m.current, pet.SizeNormal, m.barStyle, m.barShowPet, m.barWidth)
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
	sample := pet.SampleSegmentData(m.current, pet.SizeNormal, m.barStyle, m.barShowPet, m.barWidth)
	var previewLines []string
	for i := 0; i < maxLines; i++ {
		if len(m.lines[i]) == 0 {
			continue
		}
		colors := m.lineColors[i]
		var rendered string
		if len(colors) > 0 {
			rendered = m.renderColoredPreview(m.lines[i], colors, sample)
		} else {
			tmpl := pet.SegmentsToTemplate(m.lines[i], m.separator)
			rendered = pet.RenderTemplate(tmpl, sample)
		}
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
		nav(b, "(a)dd \u00b7 (i)nsert \u00b7 (d)elete \u00b7 (c)lear \u00b7 (f)oreground")
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
	colors := m.lineColors[m.lineFocused]
	for i, seg := range segs {
		emoji, label := segmentParts(seg)
		// Show color swatch if segment has a color.
		if i < len(colors) && colors[i] != 0 {
			swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", colors[i]))).Render("\u2588")
			label = swatch + " " + label
		}
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

// renderColoredPreview renders a line with lipgloss colors for the TUI preview.
func (m model) renderColoredPreview(segs []pet.Segment, colors []uint8, sample *pet.SegmentData) string {
	type item struct {
		text  string
		kind  pet.SegmentKind
		color uint8
	}
	var items []item
	for i, seg := range segs {
		var c uint8
		if i < len(colors) {
			c = colors[i]
		}
		var text string
		switch seg.Kind {
		case pet.KindToken:
			text = pet.RenderTemplate("{"+seg.Value+"}", sample)
		case pet.KindSeparator:
			text = seg.Value
			if text == "" {
				text = m.separator
			}
		case pet.KindCommand:
			text = "[cmd]"
		}
		items = append(items, item{text: text, kind: seg.Kind, color: c})
	}
	// Filter empty tokens.
	var filtered []item
	for _, r := range items {
		if r.kind != pet.KindSeparator && r.text == "" {
			continue
		}
		filtered = append(filtered, r)
	}
	// Remove dangling separators.
	var cleaned []item
	for i, r := range filtered {
		if r.kind == pet.KindSeparator {
			if len(cleaned) == 0 || i == len(filtered)-1 || cleaned[len(cleaned)-1].kind == pet.KindSeparator {
				continue
			}
		}
		cleaned = append(cleaned, r)
	}
	if len(cleaned) > 0 && cleaned[len(cleaned)-1].kind == pet.KindSeparator {
		cleaned = cleaned[:len(cleaned)-1]
	}
	var b strings.Builder
	for i, r := range cleaned {
		if i > 0 && r.kind != pet.KindSeparator && cleaned[i-1].kind != pet.KindSeparator {
			b.WriteByte(' ')
		}
		text := r.text
		if r.color != 0 {
			text = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", r.color))).Render(text)
		}
		b.WriteString(text)
	}
	return b.String()
}

func (m model) viewColorPicker(b *strings.Builder) {
	header(b, "\U0001F3A8", "Foreground Color")
	nav(b, "esc back \u00b7 arrows navigate \u00b7 enter select")
	b.WriteString("\n")

	// Render grid of color swatches.
	for i, c := range colorPalette {
		if i > 0 && i%colorsPerRow == 0 {
			b.WriteString("\n")
		}
		if i%colorsPerRow == 0 {
			b.WriteString("      ")
		}
		swatch := "\u2588\u2588"
		if c == 0 {
			swatch = "--"
		} else {
			swatch = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", c))).Render(swatch)
		}
		if i == m.colorCursor {
			b.WriteString(cursorStyle.Render("[") + swatch + cursorStyle.Render("]"))
		} else {
			b.WriteString(" " + swatch + " ")
		}
	}
	b.WriteString("\n\n")

	// Show selected color label and preview.
	selected := colorPalette[m.colorCursor]
	label := colorLabel(selected)
	b.WriteString(fmt.Sprintf("      %s %s\n", dimStyle.Render("color:"), valueStyle.Render(label)))

	// Preview: show the focused segment text in the selected color.
	segs := m.lines[m.lineFocused]
	if m.segCursor < len(segs) {
		sample := pet.SampleSegmentData(m.current, pet.SizeNormal, m.barStyle, m.barShowPet, m.barWidth)
		seg := segs[m.segCursor]
		var text string
		switch seg.Kind {
		case pet.KindToken:
			text = pet.RenderTemplate("{"+seg.Value+"}", sample)
		case pet.KindSeparator:
			text = m.separator
		case pet.KindCommand:
			text = "[cmd]"
		}
		if selected == 0 {
			b.WriteString(fmt.Sprintf("      %s %s\n", dimStyle.Render("preview:"), text))
		} else {
			styled := lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", selected))).Render(text)
			b.WriteString(fmt.Sprintf("      %s %s\n", dimStyle.Render("preview:"), styled))
		}
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
