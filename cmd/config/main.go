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
	savedStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("78"))
	previewStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("249"))
	boxStyle      = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
)

type section int

const (
	sectionSpecies section = iota
	sectionContextMode
	sectionLines
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

	current        pet.Species
	currentCtxMode pet.ContextMode

	chosenSpecies pet.Species
	chosenCtxMode pet.ContextMode

	// Segment editor state
	lines       [maxLines][]pet.Segment
	lineFocused int
	segCursor   int
	mode        editMode

	// Picker sub-mode
	pickerItems  []string
	pickerCursor int
	pickerInsert bool // true = insert before cursor, false = append or replace
	pickerReplace bool // true = replacing existing segment

	// Inline text edit
	editBuf     []rune
	editCursor  int
	editInPlace bool // true = editing existing segment, false = creating new one

	saved    bool
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
		section:        sectionSpecies,
		options:        opts,
		ctxOptions:     ctxOpts,
		cursor:         cursor,
		ctxCursor:      ctxCursor,
		current:        cfg.Species,
		currentCtxMode: cfg.ContextMode,
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
		case sectionSpecies:
			return m.updateSpecies(msg)
		case sectionContextMode:
			return m.updateContextMode(msg)
		case sectionLines:
			return m.updateLines(msg)
		}
	}

	return m, nil
}

func (m model) updateSpecies(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.options)-1 {
			m.cursor++
		}
	case "enter":
		m.chosenSpecies = m.options[m.cursor].species
		m.section = sectionContextMode
	}
	return m, nil
}

func (m model) updateContextMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.ctxCursor > 0 {
			m.ctxCursor--
		}
	case "down", "j":
		if m.ctxCursor < len(m.ctxOptions)-1 {
			m.ctxCursor++
		}
	case "enter":
		m.chosenCtxMode = m.ctxOptions[m.ctxCursor].mode
		m.section = sectionLines
	}
	return m, nil
}

func (m model) updateLines(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		m.quitting = true
		return m, tea.Quit
	case "ctrl+s":
		return m.save()
	case "tab":
		m.lineFocused = (m.lineFocused + 1) % maxLines
		m.clampSegCursor()
	case "shift+tab":
		m.lineFocused = (m.lineFocused - 1 + maxLines) % maxLines
		m.clampSegCursor()
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
		}
	case "c": // clear line
		m.lines[m.lineFocused] = nil
		m.segCursor = 0
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

func (m model) save() (tea.Model, tea.Cmd) {
	var lines []string
	for i := 0; i < maxLines; i++ {
		if len(m.lines[i]) > 0 {
			lines = append(lines, pet.SegmentsToTemplate(m.lines[i]))
		}
	}
	if len(lines) == 0 {
		lines = pet.DefaultLines
	}
	cfg := &pet.Config{
		Species:     m.chosenSpecies,
		ContextMode: m.chosenCtxMode,
		Lines:       lines,
	}
	if err := pet.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		return m, tea.Quit
	}
	m.saved = true
	m.quitting = true
	return m, tea.Quit
}

func (m model) View() string {
	if m.saved {
		opt := m.options[m.cursor]
		ctxOpt := m.ctxOptions[m.ctxCursor]
		lineCount := 0
		for i := 0; i < maxLines; i++ {
			if len(m.lines[i]) > 0 {
				lineCount++
			}
		}
		return savedStyle.Render(fmt.Sprintf("Saved! Pet: %s %s | Context: %s | Lines: %d",
			opt.label, opt.preview, ctxOpt.label, lineCount)) + "\n"
	}
	if m.quitting {
		return ""
	}

	var b strings.Builder

	switch m.section {
	case sectionSpecies:
		b.WriteString(titleStyle.Render("Choose your pet species"))
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

	case sectionContextMode:
		b.WriteString(titleStyle.Render("Context bar mode"))
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

	case sectionLines:
		m.viewLines(&b)
	}

	b.WriteString(dimStyle.Render("\nesc quit"))
	b.WriteString("\n")
	return b.String()
}

func (m model) viewLines(b *strings.Builder) {
	// Live preview box
	species := m.chosenSpecies
	sample := pet.SampleSegmentData(species, pet.SizeNormal)
	var previewLines []string
	for i := 0; i < maxLines; i++ {
		if len(m.lines[i]) == 0 {
			continue
		}
		tmpl := pet.SegmentsToTemplate(m.lines[i])
		rendered := pet.RenderTemplate(tmpl, sample)
		if rendered != "" {
			previewLines = append(previewLines, "  "+rendered)
		}
	}
	previewContent := strings.Join(previewLines, "\n")
	if previewContent == "" {
		previewContent = dimStyle.Render("  (empty)")
	}
	header := "> Preview  (ctrl+s to save)"
	b.WriteString(boxStyle.Render(header + "\n" + previewContent))
	b.WriteString("\n\n")

	// Line tabs
	b.WriteString(titleStyle.Render(fmt.Sprintf("Edit Line %d", m.lineFocused+1)))
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("tab: switch line"))
	b.WriteString("\n")

	// Key hints
	switch m.mode {
	case modeList:
		b.WriteString(dimStyle.Render("up/down select  left/right change type  space edit sep"))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("(a)dd  (i)nsert  (d)elete  (c)lear  ctrl+s save"))
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
