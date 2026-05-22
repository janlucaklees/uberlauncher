package ui

import (
	"fmt"
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
	Next Direction = 1
	None Direction = 0
	Down Direction = -1
	Prev Direction = -1
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

	skillIdStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	debugMessageStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
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
		m.updateMessageCursor(Next)
		cmd = waitForNotifierEvent(m.notifier.Events)

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.updateCursor(None)

	case tea.KeyMsg:
		switch key := msg.Key(); key.Code {

		case tea.KeyEscape:
			return m.quit()

		case tea.KeyUp:
			if !m.isFreeTextModeActive() {
				m.updateCursor(Up)
			}

		case tea.KeyDown:
			if !m.isFreeTextModeActive() {
				m.updateCursor(Down)
			}

		case tea.KeyTab:
			m.updateMessageCursor(Prev)

		case tea.KeyEnter:
			selected, ok := m.getSelectedEntry()

			if ok {
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
			} else {
				m.notifier.ReportError(fmt.Errorf("No matching entry found!"))
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

func (m model) getSelectedEntry() (entry.Entry, bool) {

	// If free text mode is active, return the matching entry.
	e, ok := m.getActiveFreeTextEntry()
	if ok {
		return e, true
	}

	// Otherwise, return the entry under the cursor, in case there is one.
	if len(m.entries) > 0 {
		return m.entries[m.cursor], true
	}

	return entry.Entry{}, false
}

func (m *model) getEntryListHeight() int {
	// The total height minus one line of input, one line of messages.
	return max(m.height-1-1, 0)
}

func (m *model) getActiveFreeTextEntry() (entry.Entry, bool) {
	i := m.input.Value()

	for _, e := range m.entries {
		if strings.HasPrefix(i, e.Label+" ") {
			return e, true
		}
	}

	return entry.Entry{}, false
}

func (m *model) isFreeTextModeActive() bool {
	_, ok := m.getActiveFreeTextEntry()
	return ok
}

func (m *model) updateCursor(d Direction) {
	m.cursor = wrapCursor(
		0,
		max(0, min(m.getEntryListHeight(), len(m.entries))-1),
		m.cursor+int(d),
	)
}

func (m *model) updateMessageCursor(d Direction) {
	m.messageCursor = wrapCursor(
		0,
		max(0, len(m.messages)-1),
		m.messageCursor+int(d),
	)
}

func wrapCursor(min int, max int, cursor int) int {
	n := max - min + 1
	return min + ((cursor-min)%n+n)%n
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
	s += renderEntryLine(m.input.View(), "", m.isFreeTextModeActive())

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

func renderEntry(e entry.Entry, isSelected bool) string {
	label := e.Label
	if e.IsFreeText {
		label += freeTextEntryChar
	}

	return renderEntryLine(label, skillIdStyle.Render(e.SkillId), isSelected)
}

func renderEntryLine(content, annotation string, isSelected bool) string {
	cursor := unselectedCursorStyle.Render(cursorChar)

	if isSelected {
		cursor = selectedCursorStyle.Render(cursorChar)
		content = selectedEntryStyle.Render(content)
	}

	if annotation != "" {
		return fmt.Sprintf("%s %s %s\n", cursor, content, annotation)
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
	case notifier.Debug:
		style = debugMessageStyle
	case notifier.Warning:
		style = warningMessageStyle
	case notifier.Error:
		style = errorMessageStyle
	default:
		style = infoMessageStyle
	}

	return fmt.Sprintf("(%d/%d) %s\n", cursor+1, len(messages), style.Render(message.Text))
}
