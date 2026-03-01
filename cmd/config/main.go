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

type speciesOption struct {
	species pet.Species
	label   string
	preview string // emoji progression
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

type model struct {
	options  []speciesOption
	cursor   int
	current  pet.Species
	saved    bool
	quitting bool
}

func initialModel() model {
	cfg := pet.LoadConfig()
	opts := speciesOptions()
	cursor := 0
	for i, o := range opts {
		if o.species == cfg.Species {
			cursor = i
			break
		}
	}
	return model{
		options: opts,
		cursor:  cursor,
		current: cfg.Species,
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
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			chosen := m.options[m.cursor].species
			if err := pet.SaveConfig(&pet.Config{Species: chosen}); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				return m, tea.Quit
			}
			m.current = chosen
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
		return savedStyle.Render(fmt.Sprintf("Saved! Your pet is now: %s %s", opt.label, opt.preview)) + "\n"
	}
	if m.quitting {
		return ""
	}

	var b strings.Builder
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
