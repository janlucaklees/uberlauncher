package ui

import (
	"fmt"
	"uberlauncher/internal/entry"
	"uberlauncher/internal/notifier"
	"uberlauncher/internal/store"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Direction int

const (
	Up   Direction = 1
	None Direction = 0
	Down Direction = -1
)

var (
	selectedInputStyles   = textinput.DefaultDarkStyles()
	unselectedInputStyles = textinput.DefaultDarkStyles()
)

var (
	cursorChar            = "┃"
	inputPromptChar       = "› "
	selectedCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	unselectedCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
	selectedEntryStyle    = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("236"))
)

type model struct {
	entries  []entry.Entry
	height   int
	input    textinput.Model
	cursor   int
	messages []notifier.Event
	store    *store.Store
	notifier *notifier.Notifier
}

func New(st *store.Store, n *notifier.Notifier) model {
	selectedInputStyles.Focused.Text = selectedEntryStyle

	i := textinput.New()
	i.Prompt = inputPromptChar
	i.Placeholder = ""
	i.Focus()

	return model{
		entries:  st.GetMatches(""),
		input:    i,
		cursor:   1,
		store:    st,
		notifier: n,
	}
}

func waitForNotifierEvent(ch <-chan notifier.Event) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (m model) Init() tea.Cmd {
	return waitForNotifierEvent(m.notifier.Events)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case notifier.Event:
		m.messages = append(m.messages, msg)

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.updateCursor(None)

	case tea.KeyMsg:
		switch key := msg.Key(); key.Code {

		case tea.KeyEscape:
			return m.quit()

		case tea.KeyUp:
			m.updateCursor(Up)

		case tea.KeyDown:
			m.updateCursor(Down)

		case tea.KeyEnter:
			selected := m.getSelectedEntry()
			if selected != nil {
				selected.Run(entry.Context{Input: m.input.Value()})
			}

		default:
			m.input, cmd = m.input.Update(msg)
			m.entries = m.store.GetMatches(m.input.Value())
			m.updateCursor(None)
		}
	}

	return m, cmd
}

func (m model) quit() (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m model) getSelectedEntry() *entry.Entry {
	if m.cursor == 0 {
		return nil
	}
	return &m.entries[m.cursor-1]
}

func (m *model) getEntryListHeight() int {
	// The total height minus one line of input
	return max(m.height-1, 0)
}

func (m *model) isInputSelected() bool {
	return m.cursor == 0
}

func (m *model) updateCursor(d Direction) {
	first := 0
	last := min(m.getEntryListHeight(), len(m.entries)) - 1

	m.cursor += int(d)
	if m.cursor < first {
		m.cursor = last
	} else if m.cursor > last {
		m.cursor = first
	}
}

func (m model) View() tea.View {
	s := ""

	// Render the entries
	s += renderEntries(m.entries, m.cursor, m.getEntryListHeight())

	// Render the input
	if m.isInputSelected() {
		m.input.SetStyles(selectedInputStyles)
	} else {
		m.input.SetStyles(unselectedInputStyles)
	}

	s += renderEntry(m.input.View(), m.isInputSelected())
	return tea.NewView(s)
}

func renderEntries(entries []entry.Entry, cursor int, availableHeight int) string {
	numEntries := min(availableHeight, len(entries))

	var s string

	// Add empty lines to the list to fill the available space so the input stays at the bottom.
	for i := availableHeight; i > numEntries; i-- {
		s += "\n"
	}

	// Render the best matching entries in reverse, so that the best match is closest to the input.
	for i := numEntries; i > 0; i-- {
		s += renderEntry(entries[i-1].Label, i == cursor)
	}

	return s
}

func renderEntry(entry string, isSelected bool) string {
	cursor := unselectedCursorStyle.Render(cursorChar)

	if isSelected {
		cursor = selectedCursorStyle.Render(cursorChar)
		entry = selectedEntryStyle.Render(entry)
	}

	return fmt.Sprintf("%s %s\n", cursor, entry)
}
