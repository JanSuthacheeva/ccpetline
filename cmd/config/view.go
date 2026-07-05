package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"

	"github.com/jansuthacheeva/ccpetline/internal/pet"
)

func (m model) viewBarStyle(b *strings.Builder) {
	header(b, "\U0001F4CA", "Bar Style")
	nav(b, "esc back \u00b7 enter select \u00b7 \u2190\u2192 adjust")
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

// powerlineSepPreview renders two small colored blocks joined by the given
// separator style so the glyph is shown in context.
func powerlineSepPreview(s pet.PowerlineSepStyle) string {
	g := pet.PowerlineSepGlyph(s)
	return fmt.Sprintf("\x1b[48;5;31m  \x1b[38;5;31;48;5;238m%s\x1b[0m\x1b[48;5;238m  \x1b[0m\x1b[38;5;238m%s\x1b[0m", g, g)
}

func (m model) barPreview(style pet.BarStyle, showPet bool) string {
	s := &pet.State{
		Species:    m.current,
		Size:       pet.SizeNormal,
		ContextPct: 53.1,
		BarStyle:   style,
		BarShowPet: showPet,
		BarWidth:   m.barWidth,
		IconTheme:  m.iconTheme,
	}
	return pet.FormatSeparator(s)
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
	case sectionStyle:
		m.viewStyle(&b)
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
		case sectionStyle:
			detail = m.styleSummary()
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

// styleSummary is the one-line description shown next to the Style menu item.
func (m model) styleSummary() string {
	if !m.nerdFont {
		return "Text"
	}
	icons := "glyphs"
	if m.iconTheme == pet.IconThemeText {
		icons = "text"
	}
	s := "Nerd Font · " + icons
	if m.powerline {
		s += " · Powerline " + pet.PowerlineSepLabel(m.powerlineSep)
	}
	return s
}

// stylePreviewTmpl is the sample line rendered in the Style screen preview box.
const stylePreviewTmpl = "{cwd} | {branch} | {model} | {pet} {mood}"

func (m model) viewStyle(b *strings.Builder) {
	// Preview box reflecting the current icon theme and powerline look.
	sample := pet.SampleSegmentData(m.current, pet.SizeNormal, m.barStyle, m.barShowPet, m.barWidth, m.iconTheme)
	var preview string
	if m.powerline {
		segs := pet.TemplateToSegments(stylePreviewTmpl)
		colors := pet.DefaultLineColors([]string{stylePreviewTmpl})[0]
		preview = pet.RenderPowerlineLine(segs, colors, sample, m.powerlineSep)
	} else {
		preview = pet.RenderTemplate(stylePreviewTmpl, sample)
	}
	b.WriteString("\n")
	b.WriteString(boxStyle.Render("\U0001F441  Preview\n  " + preview))
	b.WriteString("\n")

	if m.firstRun {
		header(b, "\U0001F3A8", "Welcome to ccpetline")
		nav(b, "let's set up how your status line looks")
		nav(b, "↑↓ move · ←→ change · enter done")
	} else {
		header(b, "\U0001F3A8", "Style")
		nav(b, "esc back · ↑↓ move · ←→ change")
	}
	b.WriteString("\n")

	onOff := func(v bool) string {
		if v {
			return "on"
		}
		return "off"
	}

	// Nerd Font capability row (always shown).
	row(b, m.styleCursor == styleRowNerdFont, "✨",
		fmt.Sprintf("Nerd Font %s", valueStyle.Render(onOff(m.nerdFont))))
	if !m.nerdFont {
		b.WriteString(hintStyle.Render("      enables glyph icons and the Powerline look"))
		b.WriteString("\n")
		b.WriteString(hintStyle.Render("      no glyphs? install one at nerdfonts.com"))
		b.WriteString("\n")
		return
	}

	// Icon style row.
	iconVal := "Nerd glyphs"
	if m.iconTheme == pet.IconThemeText {
		iconVal = "Text labels"
	}
	row(b, m.styleCursor == styleRowIcons, "\U0001F524",
		fmt.Sprintf("Icons %s", valueStyle.Render(iconVal)))

	// Powerline toggle row.
	row(b, m.styleCursor == styleRowPowerline, "▓",
		fmt.Sprintf("Powerline (segment backgrounds) %s", valueStyle.Render(onOff(m.powerline))))

	// Separator glyph row (only while powerline is on).
	if m.powerline {
		sepText := fmt.Sprintf("Separator %s  %s",
			valueStyle.Render(pet.PowerlineSepLabel(m.powerlineSep)),
			dimStyle.Render(powerlineSepPreview(m.powerlineSep)))
		row(b, m.styleCursor == styleRowSeparator, "➤", sepText)
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
	if strings.HasPrefix(m.updateStatus, "Error:") {
		b.WriteString(fmt.Sprintf("      Changelog: %s\n", dimStyle.Render(changelogURL)))
		b.WriteString("\n")
		nav(b, "press any key to return")
	} else {
		b.WriteString(fmt.Sprintf("      You can read more about the changes here:\n"))
		b.WriteString(fmt.Sprintf("      %s\n", dimStyle.Render(changelogURL)))
		b.WriteString("\n")
		nav(b, "press any key to close")
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
	sample := pet.SampleSegmentData(m.current, pet.SizeNormal, m.barStyle, m.barShowPet, m.barWidth, m.iconTheme)
	for i := 0; i < maxLines; i++ {
		preview := dimStyle.Render("(empty)")
		if len(m.lines[i]) > 0 {
			if m.powerline {
				preview = pet.RenderPowerlineLine(m.lines[i], m.lineColors[i], sample, m.powerlineSep)
			} else {
				tmpl := pet.SegmentsToTemplate(m.lines[i], m.separator)
				preview = valueStyle.Render(pet.RenderTemplate(tmpl, sample))
			}
		}
		row(b, i == m.lineFocused, lineEmojis[i], preview)
	}
}

func (m model) viewLineEdit(b *strings.Builder) {
	// Preview box
	sample := pet.SampleSegmentData(m.current, pet.SizeNormal, m.barStyle, m.barShowPet, m.barWidth, m.iconTheme)
	var previewLines []string
	for i := 0; i < maxLines; i++ {
		if len(m.lines[i]) == 0 {
			continue
		}
		colors := m.lineColors[i]
		var rendered string
		switch {
		case m.powerline:
			rendered = pet.RenderPowerlineLine(m.lines[i], colors, sample, m.powerlineSep)
		case len(colors) > 0:
			rendered = m.renderColoredPreview(m.lines[i], colors, sample)
		default:
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
		emoji, label := m.segmentParts(seg)
		// Show color swatch if segment has a color.
		if i < len(colors) && colors[i] != 0 {
			swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", colors[i]))).Render("\u2588")
			label = swatch + " " + label
		}
		row(b, i == m.segCursor, emoji, label)
	}
}

// tokenIcon returns the display icon for a token, honoring the active icon
// theme: the Nerd Font glyph in nerd mode (when one exists), otherwise the
// emoji fallback. Keeps the config's per-token icons consistent with the
// rendered status line.
func (m model) tokenIcon(token string) string {
	if g := pet.TokenIcon(m.iconTheme, token); g != "" {
		return g
	}
	if e := tokenEmoji[token]; e != "" {
		return e
	}
	return " "
}

func (m model) segmentParts(seg pet.Segment) (emoji, label string) {
	switch seg.Kind {
	case pet.KindToken:
		return m.tokenIcon(seg.Value), capitalize(seg.Value)
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
		emoji := m.tokenIcon(item)
		switch item {
		case "Separator":
			emoji = "\u2502"
		case "Command":
			emoji = "\u26a1"
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
		sample := pet.SampleSegmentData(m.current, pet.SizeNormal, m.barStyle, m.barShowPet, m.barWidth, m.iconTheme)
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
