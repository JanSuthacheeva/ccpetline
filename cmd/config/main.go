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
	cursor         int
	ctxCursor      int
	current        pet.Species
	currentCtxMode pet.ContextMode
	chosenSpecies  pet.Species
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
	return model{
		section:        sectionSpecies,
		options:        opts,
		ctxOptions:     ctxOpts,
		cursor:         cursor,
		ctxCursor:      ctxCursor,
		current:        cfg.Species,
		currentCtxMode: cfg.ContextMode,
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
			if m.section == sectionSpecies {
				if m.cursor > 0 {
					m.cursor--
				}
			} else {
				if m.ctxCursor > 0 {
					m.ctxCursor--
				}
			}
		case "down", "j":
			if m.section == sectionSpecies {
				if m.cursor < len(m.options)-1 {
					m.cursor++
				}
			} else {
				if m.ctxCursor < len(m.ctxOptions)-1 {
					m.ctxCursor++
				}
			}
		case "enter":
			if m.section == sectionSpecies {
				m.chosenSpecies = m.options[m.cursor].species
				m.section = sectionContextMode
				return m, nil
			}
			chosenMode := m.ctxOptions[m.ctxCursor].mode
			cfg := &pet.Config{Species: m.chosenSpecies, ContextMode: chosenMode}
			if err := pet.SaveConfig(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				return m, tea.Quit
			}
			m.current = m.chosenSpecies
			m.currentCtxMode = chosenMode
			m.saved = true
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.saved {
		opt := m.options[m.cursor]
		ctxOpt := m.ctxOptions[m.ctxCursor]
		return savedStyle.Render(fmt.Sprintf("Saved! Pet: %s %s | Context: %s", opt.label, opt.preview, ctxOpt.label)) + "\n"
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
	} else {
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
