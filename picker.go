package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neuralinkcorp/tsui/ui"
)

// A picker is a modal third-column list that opens on top of a submenu row.
// It lets the user choose a value from a list, then calls onSelect with the
// chosen label when the user hits Enter.
type pickerState struct {
	title    string
	options  []string
	cursor   int
	// Currently-applied value, marked with a colored bullet. Empty for none.
	current  string
	onSelect func(label string) tea.Msg
}

const pickerColumnWidth = 30

// newPicker builds a picker. If current matches one of options, the cursor
// starts on that row; otherwise it starts at 0.
func newPicker(title string, options []string, current string, onSelect func(string) tea.Msg) *pickerState {
	p := &pickerState{
		title:    title,
		options:  options,
		current:  current,
		onSelect: onSelect,
	}
	for i, opt := range options {
		if opt == current {
			p.cursor = i
			break
		}
	}
	return p
}

func (p *pickerState) cursorUp() {
	if p.cursor > 0 {
		p.cursor--
	}
}

func (p *pickerState) cursorDown() {
	if p.cursor < len(p.options)-1 {
		p.cursor++
	}
}

// commit returns the tea.Msg for the currently-highlighted option.
func (p *pickerState) commit() tea.Msg {
	if len(p.options) == 0 {
		return nil
	}
	return p.onSelect(p.options[p.cursor])
}

func (p *pickerState) render(height int) string {
	var s strings.Builder

	title := lipgloss.NewStyle().
		Faint(true).
		PaddingLeft(2).
		Render(p.title)
	s.WriteString(title)
	s.WriteByte('\n')
	s.WriteByte('\n')

	for i, opt := range p.options {
		innerStyle := lipgloss.NewStyle()
		if i == p.cursor {
			innerStyle = innerStyle.
				Background(ui.CurrentTheme.Secondary).
				Foreground(ui.CurrentTheme.FgOnSecondary)
		}

		outerStyle := innerStyle.
			PaddingLeft(2).
			PaddingRight(1).
			Width(pickerColumnWidth)

		var inner string
		if opt == p.current {
			inner = innerStyle.Foreground(ui.CurrentTheme.Success).Render("●") +
				innerStyle.Render(" "+opt)
		} else {
			inner = innerStyle.Render("  " + opt)
		}

		s.WriteString(outerStyle.Render(inner))
		if i < len(p.options)-1 {
			s.WriteByte('\n')
		}
	}

	// Fill to requested height so the column aligns with siblings.
	return lipgloss.NewStyle().Height(height).Render(s.String())
}

// openPickerMsg asks the main update loop to install a new picker.
type openPickerMsg struct {
	picker *pickerState
}
