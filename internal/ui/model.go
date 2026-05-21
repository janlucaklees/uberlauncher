package ui

import (
	"fmt"
	"slices"
	"strings"
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
	cursorChar        = "┃"
	inputPromptChar   = "› "
	freeTextEntryChar = " …"

	selectedCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	unselectedCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
	selectedEntryStyle    = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("236"))

	infoMessageStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	warningMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	errorMessageStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type model struct {
	entries []entry.Entry
	height  int
	input   textinput.Model
	cursor  int

	messages      []notifier.Event
	messageCursor int

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
		entries: st.GetMatches(""),
		input:   i,
		cursor:  0,

		messageCursor: -1,

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
		m.messageCursor = len(m.messages) - 1
		cmd = waitForNotifierEvent(m.notifier.Events)

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

		case tea.KeyTab:
			m.cycleMessages()

		case tea.KeyEnter:
			selected := m.getSelectedEntry()

			if selected != nil {
				closeAfterRun := true

				selected.Run(entry.Context{
					Input: m.input.Value(),
					UI: entry.UI{
						KeepOpen: func() {
							closeAfterRun = false
						},
					},
				})

				if closeAfterRun {
					return m.quit()
				}
			}

		default:
			m.input, cmd = m.input.Update(msg)

			if !m.isFreeTextModeActive() {
				m.entries = m.store.GetMatches(m.input.Value())
			}

			m.updateCursor(None)
		}
	}

	return m, cmd
}

func (m model) quit() (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m model) getSelectedEntry() *entry.Entry {
	return &m.entries[m.cursor]
}

func (m *model) getEntryListHeight() int {
	// The total height minus one line of input, one line of messages.
	return max(m.height-1-1, 0)
}

func (m *model) isFreeTextModeActive() bool {
	return slices.ContainsFunc(m.entries, func(entry entry.Entry) bool {
		return strings.HasPrefix(m.input.Value(), entry.Label+" ")
	})
}

func (m *model) updateCursor(d Direction) {
	first := 0
	last := min(m.getEntryListHeight(), len(m.entries)) - 1 // -1 to get the index of the last entry

	m.cursor += int(d)
	if m.cursor < first {
		m.cursor = last
	} else if m.cursor > last {
		m.cursor = first
	}
}

func (m *model) cycleMessages() {
	first := 0
	last := len(m.messages) - 1

	m.messageCursor -= 1
	if m.messageCursor < first {
		m.messageCursor = last
	}
}

func (m model) View() tea.View {
	s := ""

	// Render the entry list
	if m.isFreeTextModeActive() {
		// When free-text mode is active, make sure the cursor never shows up in the entry list.
		s += renderEntries(m.entries, -1, m.getEntryListHeight())
	} else {
		s += renderEntries(m.entries, m.cursor, m.getEntryListHeight())
	}

	// Render the input
	if m.isFreeTextModeActive() {
		m.input.SetStyles(selectedInputStyles)
	} else {
		m.input.SetStyles(unselectedInputStyles)
	}
	s += renderEntryLine(m.input.View(), m.isFreeTextModeActive())

	// Render messages
	s += renderMessageLine(m.messages, m.messageCursor)

	return tea.NewView(strings.TrimSuffix(s, "\n"))
}

func renderEntries(entries []entry.Entry, cursor int, availableHeight int) string {
	numEntries := min(availableHeight, len(entries))

	var s string

	// Add empty lines to the list to fill the available space so the input stays at the bottom.
	for i := availableHeight - 1; i >= numEntries; i-- {
		s += "\n"
	}

	// Render the best matching entries in reverse, so that the best match is closest to the input.
	for i := numEntries - 1; i >= 0; i-- {
		s += renderEntry(entries[i], i == cursor)
	}

	return s
}

func renderEntry(entry entry.Entry, isSelected bool) string {
	if entry.IsFreeText {
		return renderEntryLine(entry.Label+freeTextEntryChar, isSelected)
	}

	return renderEntryLine(entry.Label, isSelected)
}

func renderEntryLine(content string, isSelected bool) string {
	cursor := unselectedCursorStyle.Render(cursorChar)

	if isSelected {
		cursor = selectedCursorStyle.Render(cursorChar)
		content = selectedEntryStyle.Render(content)
	}

	return fmt.Sprintf("%s %s\n", cursor, content)
}

func renderMessageLine(messages []notifier.Event, cursor int) string {
	if len(messages) == 0 {
		return "\n"
	}

	message := messages[cursor]

	var style lipgloss.Style
	switch message.Severity {
	case notifier.Warning:
		style = warningMessageStyle
	case notifier.Error:
		style = errorMessageStyle
	default:
		style = infoMessageStyle
	}

	return fmt.Sprintf("(%d/%d) %s\n", cursor+1, len(messages), style.Render(message.Text))
}
