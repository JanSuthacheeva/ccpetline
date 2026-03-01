package main

import (
	"fmt"
	"os"
	"strings"

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
)

type section int

const (
	sectionSpecies section = iota
	sectionContextMode
	sectionShowSnacks
	sectionLayout
	sectionPosition
)

type speciesOption struct {
	species pet.Species
	label   string
	preview string // emoji progression
}

type contextModeOption struct {
	mode  pet.ContextMode
	label string
	desc  string
}

type boolOption struct {
	value bool
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

type model struct {
	section        section
	options        []speciesOption
	ctxOptions     []contextModeOption
	snackOptions    []boolOption
	layoutOptions   []boolOption
	positionOptions []boolOption
	cursor          int
	ctxCursor       int
	snackCursor     int
	layoutCursor    int
	positionCursor  int
	current         pet.Species
	currentCtxMode  pet.ContextMode
	currentSnacks   bool
	currentLayout   bool // true = single line
	currentPosition bool // true = pet on top
	chosenSpecies   pet.Species
	chosenCtxMode   pet.ContextMode
	chosenSnacks    bool
	chosenLayout    bool
	saved          bool
	quitting       bool
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
	showSnacks := cfg.ShowSnacks == nil || *cfg.ShowSnacks
	snackOpts := []boolOption{
		{value: true, label: "Yes", desc: "show snack counter"},
		{value: false, label: "No", desc: "hide snack counter"},
	}
	snackCursor := 0
	if !showSnacks {
		snackCursor = 1
	}
	layoutOpts := []boolOption{
		{value: false, label: "Two lines", desc: "pet status + context bar"},
		{value: true, label: "Single line", desc: "compact context bar only"},
	}
	layoutCursor := 0
	if cfg.SingleLine {
		layoutCursor = 1
	}
	petOnTop := cfg.PetOnTop == nil || *cfg.PetOnTop
	positionOpts := []boolOption{
		{value: true, label: "Above", desc: "pet lines above status line"},
		{value: false, label: "Below", desc: "pet lines below status line"},
	}
	positionCursor := 0
	if !petOnTop {
		positionCursor = 1
	}
	return model{
		section:         sectionSpecies,
		options:         opts,
		ctxOptions:      ctxOpts,
		snackOptions:    snackOpts,
		layoutOptions:   layoutOpts,
		positionOptions: positionOpts,
		cursor:          cursor,
		ctxCursor:       ctxCursor,
		snackCursor:     snackCursor,
		layoutCursor:    layoutCursor,
		positionCursor:  positionCursor,
		current:         cfg.Species,
		currentCtxMode:  cfg.ContextMode,
		currentSnacks:   showSnacks,
		currentLayout:   cfg.SingleLine,
		currentPosition: petOnTop,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			switch m.section {
			case sectionSpecies:
				if m.cursor > 0 {
					m.cursor--
				}
			case sectionContextMode:
				if m.ctxCursor > 0 {
					m.ctxCursor--
				}
			case sectionShowSnacks:
				if m.snackCursor > 0 {
					m.snackCursor--
				}
			case sectionLayout:
				if m.layoutCursor > 0 {
					m.layoutCursor--
				}
			case sectionPosition:
				if m.positionCursor > 0 {
					m.positionCursor--
				}
			}
		case "down", "j":
			switch m.section {
			case sectionSpecies:
				if m.cursor < len(m.options)-1 {
					m.cursor++
				}
			case sectionContextMode:
				if m.ctxCursor < len(m.ctxOptions)-1 {
					m.ctxCursor++
				}
			case sectionShowSnacks:
				if m.snackCursor < len(m.snackOptions)-1 {
					m.snackCursor++
				}
			case sectionLayout:
				if m.layoutCursor < len(m.layoutOptions)-1 {
					m.layoutCursor++
				}
			case sectionPosition:
				if m.positionCursor < len(m.positionOptions)-1 {
					m.positionCursor++
				}
			}
		case "enter":
			switch m.section {
			case sectionSpecies:
				m.chosenSpecies = m.options[m.cursor].species
				m.section = sectionContextMode
				return m, nil
			case sectionContextMode:
				m.chosenCtxMode = m.ctxOptions[m.ctxCursor].mode
				m.section = sectionShowSnacks
				return m, nil
			case sectionShowSnacks:
				m.chosenSnacks = m.snackOptions[m.snackCursor].value
				m.section = sectionLayout
				return m, nil
			case sectionLayout:
				m.chosenLayout = m.layoutOptions[m.layoutCursor].value
				m.section = sectionPosition
				return m, nil
			case sectionPosition:
				chosenPosition := m.positionOptions[m.positionCursor].value
				showSnacks := m.chosenSnacks
				cfg := &pet.Config{
					Species:     m.chosenSpecies,
					ContextMode: m.chosenCtxMode,
					ShowSnacks:  &showSnacks,
					SingleLine:  m.chosenLayout,
					PetOnTop:    &chosenPosition,
				}
				if err := pet.SaveConfig(cfg); err != nil {
					fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
					return m, tea.Quit
				}
				m.current = m.chosenSpecies
				m.currentCtxMode = m.chosenCtxMode
				m.currentSnacks = showSnacks
				m.currentLayout = m.chosenLayout
				m.currentPosition = chosenPosition
				m.saved = true
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.saved {
		opt := m.options[m.cursor]
		ctxOpt := m.ctxOptions[m.ctxCursor]
		snackOpt := m.snackOptions[m.snackCursor]
		layoutOpt := m.layoutOptions[m.layoutCursor]
		posOpt := m.positionOptions[m.positionCursor]
		return savedStyle.Render(fmt.Sprintf("Saved! Pet: %s %s | Context: %s | Snacks: %s | Layout: %s | Position: %s",
			opt.label, opt.preview, ctxOpt.label, snackOpt.label, layoutOpt.label, posOpt.label)) + "\n"
	}
	if m.quitting {
		return ""
	}

	var b strings.Builder

	if m.section == sectionSpecies {
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
	} else if m.section == sectionContextMode {
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
			detail := fmt.Sprintf("%s — %s", name, opt.desc)

			if i == m.ctxCursor {
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, selectedStyle.Render(detail)))
			} else {
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, dimStyle.Render(detail)))
			}
		}
	} else if m.section == sectionShowSnacks {
		b.WriteString(titleStyle.Render("Show snack counter?"))
		b.WriteString("\n\n")

		for i, opt := range m.snackOptions {
			cursor := "  "
			if i == m.snackCursor {
				cursor = cursorStyle.Render("> ")
			}

			name := opt.label
			if opt.value == m.currentSnacks {
				name += " (current)"
			}
			detail := fmt.Sprintf("%s — %s", name, opt.desc)

			if i == m.snackCursor {
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, selectedStyle.Render(detail)))
			} else {
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, dimStyle.Render(detail)))
			}
		}
	} else if m.section == sectionLayout {
		b.WriteString(titleStyle.Render("Layout"))
		b.WriteString("\n\n")

		for i, opt := range m.layoutOptions {
			cursor := "  "
			if i == m.layoutCursor {
				cursor = cursorStyle.Render("> ")
			}

			name := opt.label
			if opt.value == m.currentLayout {
				name += " (current)"
			}
			detail := fmt.Sprintf("%s — %s", name, opt.desc)

			if i == m.layoutCursor {
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, selectedStyle.Render(detail)))
			} else {
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, dimStyle.Render(detail)))
			}
		}
	} else {
		b.WriteString(titleStyle.Render("Pet position"))
		b.WriteString("\n\n")

		for i, opt := range m.positionOptions {
			cursor := "  "
			if i == m.positionCursor {
				cursor = cursorStyle.Render("> ")
			}

			name := opt.label
			if opt.value == m.currentPosition {
				name += " (current)"
			}
			detail := fmt.Sprintf("%s — %s", name, opt.desc)

			if i == m.positionCursor {
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, selectedStyle.Render(detail)))
			} else {
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, dimStyle.Render(detail)))
			}
		}
	}

	b.WriteString(dimStyle.Render("\n↑/↓ navigate • enter select • q quit"))
	b.WriteString("\n")
	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
