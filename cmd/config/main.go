package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jan/claude-pet/internal/pet"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
)

type section int

const (
	sectionMenu        section = iota // main menu
	sectionSpecies                    // pet picker
	sectionContextMode                // context mode picker
	sectionSeparator                  // separator editor
	sectionLinesPicker                // pick which line (1/2/3)
	sectionLineEdit                   // segment editor for one line
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

// Menu items
type menuItem struct {
	label   string
	section section
}

func menuItems() []menuItem {
	return []menuItem{
		{label: "Edit Lines", section: sectionLinesPicker},
		{label: "Select Pet", section: sectionSpecies},
		{label: "Context Mode", section: sectionContextMode},
		{label: "Separator", section: sectionSeparator},
	}
}

// editMode tracks sub-modes within the segment list editor.
type editMode int

const (
	modeList    editMode = iota // browsing segment list
	modePicker                  // choosing segment type
	modeSepEdit                 // editing separator text
	modeCmdEdit                 // editing command text
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

	// Segment editor state
	lines       [maxLines][]pet.Segment
	lineFocused int
	segCursor   int
	mode        editMode

	// Picker sub-mode
	pickerItems   []string
	pickerCursor  int
	pickerInsert  bool // true = insert before cursor, false = append or replace
	pickerReplace bool // true = replacing existing segment

	// Inline text edit
	editBuf    []rune
	editCursor int
	editInPlace bool // true = editing existing segment, false = creating new one

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
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
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
		}
	}

	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := menuItems()
	switch msg.String() {
	case "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < len(items) {
			m.menuCursor++
		}
	case "enter":
		if m.menuCursor == len(items) {
			// Exit
			m.quitting = true
			return m, tea.Quit
		}
		dest := items[m.menuCursor].section
		if dest == sectionSeparator {
			m.editBuf = []rune(strings.TrimSpace(m.separator))
			m.editCursor = len(m.editBuf)
		}
		m.section = dest
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
	case modeSepEdit:
		return m.updateTextEdit(msg, false)
	case modeCmdEdit:
		return m.updateTextEdit(msg, true)
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
	case "a": // append
		m.mode = modePicker
		m.pickerCursor = 0
		m.pickerInsert = false
		m.pickerReplace = false
	case "i": // insert before cursor
		m.mode = modePicker
		m.pickerCursor = 0
		m.pickerInsert = true
		m.pickerReplace = false
	case "d": // delete
		if len(segs) > 0 {
			newSegs := make([]pet.Segment, 0, len(segs)-1)
			newSegs = append(newSegs, segs[:m.segCursor]...)
			newSegs = append(newSegs, segs[m.segCursor+1:]...)
			m.lines[m.lineFocused] = newSegs
			m.clampSegCursor()
			m.save()
		}
	case "c": // clear line
		m.lines[m.lineFocused] = nil
		m.segCursor = 0
		m.save()
	case " ": // space: edit separator inline
		if len(segs) > 0 && segs[m.segCursor].Kind == pet.KindSeparator {
			m.mode = modeSepEdit
			m.editBuf = []rune(segs[m.segCursor].Value)
			m.editCursor = len(m.editBuf)
			m.editInPlace = true
		}
	case "enter", "left", "right": // change segment type
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
			m.mode = modeSepEdit
			m.editBuf = []rune(" | ")
			m.editCursor = len(m.editBuf)
			m.editInPlace = false
		case "Command":
			m.mode = modeCmdEdit
			m.editBuf = nil
			m.editCursor = 0
			m.editInPlace = false
		default:
			// Token
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

func (m model) updateTextEdit(msg tea.KeyMsg, isCmd bool) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeList
	case "enter":
		text := string(m.editBuf)
		if text == "" && !isCmd {
			text = " "
		}
		kind := pet.KindSeparator
		if isCmd {
			kind = pet.KindCommand
		}
		seg := pet.Segment{Kind: kind, Value: text}
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
		// Insert printable characters
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
	}
	if err := pet.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
	}
}

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
	}

	b.WriteString("\n")
	return b.String()
}

func (m model) viewMenu(b *strings.Builder) {
	b.WriteString(titleStyle.Render("Claude Pet Config"))
	b.WriteString("\n\n")

	items := menuItems()
	for i, item := range items {
		cursor := "  "
		if i == m.menuCursor {
			cursor = cursorStyle.Render("> ")
		}

		label := item.label
		switch item.section {
		case sectionSpecies:
			label += fmt.Sprintf(" (%s)", m.current)
		case sectionContextMode:
			for _, o := range m.ctxOptions {
				if o.mode == m.currentCtxMode {
					label += fmt.Sprintf(" (%s)", o.label)
					break
				}
			}
		case sectionSeparator:
			label += fmt.Sprintf(" (%q)", m.separator)
		}

		if i == m.menuCursor {
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, selectedStyle.Render(label)))
		} else {
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, dimStyle.Render(label)))
		}
	}

	// Exit item
	cursor := "  "
	if m.menuCursor == len(items) {
		cursor = cursorStyle.Render("> ")
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, selectedStyle.Render("Exit")))
	} else {
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, dimStyle.Render("Exit")))
	}
}

func (m model) viewSpecies(b *strings.Builder) {
	b.WriteString(titleStyle.Render("Select Pet"))
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("esc: back"))
	b.WriteString("\n\n")

	for i, opt := range m.options {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		name := opt.label
		if opt.species == m.current {
			name += " (current)"
		}

		if i == m.cursor {
			b.WriteString(fmt.Sprintf("%s%s  %s\n", cursor, selectedStyle.Render(name), opt.preview))
		} else {
			b.WriteString(fmt.Sprintf("%s%s  %s\n", cursor, dimStyle.Render(name), dimStyle.Render(opt.preview)))
		}
	}
}

func (m model) viewContextMode(b *strings.Builder) {
	b.WriteString(titleStyle.Render("Context Mode"))
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("esc: back"))
	b.WriteString("\n\n")

	for i, opt := range m.ctxOptions {
		cursor := "  "
		if i == m.ctxCursor {
			cursor = cursorStyle.Render("> ")
		}

		name := opt.label
		if opt.mode == m.currentCtxMode {
			name += " (current)"
		}
		detail := fmt.Sprintf("%s -- %s", name, opt.desc)

		if i == m.ctxCursor {
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, selectedStyle.Render(detail)))
		} else {
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, dimStyle.Render(detail)))
		}
	}
}

func (m model) viewSeparator(b *strings.Builder) {
	b.WriteString(titleStyle.Render("Separator"))
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("esc: back  enter: save"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("max 3 chars, spaces added automatically"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("   %s", selectedStyle.Render(string(m.editBuf))))
	b.WriteString(cursorStyle.Render("_"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("   preview: {a}%s{b}", " "+strings.TrimSpace(string(m.editBuf))+" ")))
	b.WriteString("\n")
}

func (m model) viewLinesPicker(b *strings.Builder) {
	b.WriteString(titleStyle.Render("Edit Lines"))
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("esc: back"))
	b.WriteString("\n\n")

	sample := pet.SampleSegmentData(m.current, pet.SizeNormal)
	for i := 0; i < maxLines; i++ {
		cursor := "  "
		if i == m.lineFocused {
			cursor = cursorStyle.Render("> ")
		}

		label := "(empty)"
		if len(m.lines[i]) > 0 {
			tmpl := pet.SegmentsToTemplate(m.lines[i], m.separator)
			label = pet.RenderTemplate(tmpl, sample)
		}

		line := fmt.Sprintf("Line %d: %s", i+1, label)
		if i == m.lineFocused {
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, selectedStyle.Render(line)))
		} else {
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, dimStyle.Render(line)))
		}
	}
}

func (m model) viewLineEdit(b *strings.Builder) {
	// Live preview box
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
	header := "> Preview"
	b.WriteString(boxStyle.Render(header + "\n" + previewContent))
	b.WriteString("\n\n")

	// Line header
	b.WriteString(titleStyle.Render(fmt.Sprintf("Edit Line %d", m.lineFocused+1)))
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("esc: back"))
	b.WriteString("\n")

	// Key hints
	switch m.mode {
	case modeList:
		b.WriteString(dimStyle.Render("up/down select  left/right change type  space edit sep"))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("(a)dd  (i)nsert  (d)elete  (c)lear"))
	case modePicker:
		b.WriteString(dimStyle.Render("up/down select  enter choose  esc back"))
	case modeSepEdit:
		b.WriteString(dimStyle.Render("type separator text  enter confirm  esc cancel"))
	case modeCmdEdit:
		b.WriteString(dimStyle.Render("type shell command  enter confirm  esc cancel"))
	}
	b.WriteString("\n\n")

	// Segment list or picker overlay
	if m.mode == modePicker {
		m.viewPicker(b)
	} else if m.mode == modeSepEdit || m.mode == modeCmdEdit {
		m.viewTextEdit(b)
	} else {
		m.viewSegmentList(b)
	}
}

func (m model) viewSegmentList(b *strings.Builder) {
	segs := m.lines[m.lineFocused]
	if len(segs) == 0 {
		b.WriteString(dimStyle.Render("   (empty -- press 'a' to add)"))
		b.WriteString("\n")
		return
	}
	for i, seg := range segs {
		prefix := "   "
		if i == m.segCursor {
			prefix = cursorStyle.Render(" > ")
		}
		label := segmentLabel(seg)
		num := fmt.Sprintf("%d. ", i+1)
		if i == m.segCursor {
			b.WriteString(fmt.Sprintf("%s%s%s\n", prefix, num, selectedStyle.Render(label)))
		} else {
			b.WriteString(fmt.Sprintf("%s%s%s\n", prefix, dimStyle.Render(num), label))
		}
	}
}

func segmentLabel(seg pet.Segment) string {
	switch seg.Kind {
	case pet.KindToken:
		return capitalize(seg.Value)
	case pet.KindSeparator:
		return fmt.Sprintf("Separator %q", seg.Value)
	case pet.KindCommand:
		return fmt.Sprintf("Command %q", seg.Value)
	default:
		return seg.Value
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
		prefix := "   "
		if i == m.pickerCursor {
			prefix = cursorStyle.Render(" > ")
		}
		label := capitalize(item)
		if i == m.pickerCursor {
			b.WriteString(fmt.Sprintf("%s%s\n", prefix, selectedStyle.Render(label)))
		} else {
			b.WriteString(fmt.Sprintf("%s%s\n", prefix, label))
		}
	}
}

func (m model) viewTextEdit(b *strings.Builder) {
	label := "Separator"
	if m.mode == modeCmdEdit {
		label = "Command"
	}
	text := string(m.editBuf)
	b.WriteString(fmt.Sprintf("   %s: %s", label, selectedStyle.Render(text)))
	b.WriteString(cursorStyle.Render("_"))
	b.WriteString("\n")
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
