package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jansuthacheeva/ccpetline/internal/pet"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "55", Dark: "99"})
	accentStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "162", Dark: "212"})
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "162", Dark: "212"}).Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "239", Dark: "252"})
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "255"}).Italic(true)
	valueStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "114"})
	checkStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "22", Dark: "78"})
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "124", Dark: "196"}).Bold(true)
	boxStyle    = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "55", Dark: "63"}).
			Padding(0, 1)
)

type section int

const (
	sectionMenu section = iota
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
	sectionStyle
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
		menuItem{label: "Style", emoji: "\U0001F3A8", section: sectionStyle},
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
	"ctx_bar": "\U0001F4CA",
	"model":   "\U0001F916",
	"ctx":     "\U0001F4D0",
	"cost":    "\U0001F4B0",
	"changes": "\U0001F4DD",
	"cwd":     "\U0001F4C2",
	"dir":     "\U0001F4C1",
	"branch":  "\U0001F33F",
	"5h":      "⏳",
	"7d":      "\U0001F4C5",
	"5h_bar":  "\U0001F4CA",
	"7d_bar":  "\U0001F4CA",
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
	modeList editMode = iota
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
	iconTheme      pet.IconTheme
	nerdFont       bool
	styleCursor    int
	firstRun       bool
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
	powerline      bool
	powerlineSep   pet.PowerlineSepStyle
	barStyleCursor int

	installStatus string
	saveErr       string

	latestVersion   string
	updateAvailable bool
	updateStatus    string
	updateWaitKey   bool

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

	// LoadConfig normalizes BarShowPet, BarStyle, and BarWidth; trust it.
	barShowPet := cfg.BarShowPet == nil || *cfg.BarShowPet
	barStyleCursor := 0
	for i, s := range pet.AllBarStyles {
		if s == cfg.BarStyle {
			barStyleCursor = i
			break
		}
	}

	// First run when no config file exists yet: open the Style wizard so the
	// user declares terminal capabilities before seeing the full menu.
	section := sectionMenu
	firstRun := false
	if path := pet.ConfigPath(); path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			section = sectionStyle
			firstRun = true
		}
	}

	return model{
		section:        section,
		firstRun:       firstRun,
		options:        opts,
		ctxOptions:     ctxOpts,
		cursor:         cursor,
		ctxCursor:      ctxCursor,
		current:        cfg.Species,
		currentCtxMode: cfg.ContextMode,
		iconTheme:      cfg.IconTheme,
		nerdFont:       cfg.NerdFont,
		separator:      cfg.Separator,
		lines:          lines,
		lineColors:     lineColors,
		pickerItems:    buildPickerItems(),
		displayMode:    cfg.DisplayMode,
		displayCursor:  displayCursor,
		wrapCommand:    cfg.WrapCommand,
		barStyle:       cfg.BarStyle,
		barShowPet:     barShowPet,
		barWidth:       cfg.BarWidth,
		powerline:      cfg.Powerline,
		powerlineSep:   cfg.PowerlineSep,
		barStyleCursor: barStyleCursor,
	}
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

const (
	styleRowNerdFont = iota
	styleRowIcons
	styleRowPowerline
	styleRowSeparator
)

// styleRowCount returns how many rows are currently visible on the Style
// screen, which shrinks as capabilities are turned off.

const barRowPetToggle = 4 // len(AllBarStyles)
const barRowWidth = 5

// cyclePowerlineSep returns the separator style delta steps away from cur in
// AllPowerlineSepStyles, wrapping around.
