package ui

import (
	"fmt"
	"uberlauncher/internal/entry"
	"uberlauncher/internal/notifier"
	"uberlauncher/internal/store"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type Direction int

const (
	Up   Direction = 1
	None Direction = 0
	Down Direction = -1
)

type model struct {
	entries  []entry.Entry
	input    textinput.Model
	cursor   int
	message  string
	store    *store.Store
	notifier *notifier.Notifier
}

func New(st *store.Store, n *notifier.Notifier) model {
	i := textinput.New()
	i.Prompt = "> "
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
		switch msg.Type {
		case notifier.EventError:
			m.message = msg.Err.Error()
		case notifier.EventMessage:
			m.message = msg.Message
		}
		return m, waitForNotifierEvent(m.notifier.Events)

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
			if m.cursor > len(m.entries) {
				m.cursor = len(m.entries)
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
	if m.cursor == 0 {
		return nil
	}
	return &m.entries[m.cursor-1]
}

func (m *model) updateCursor(d Direction) {
	first := 0
	last := min(10, len(m.entries))

	m.cursor += int(d)
	if m.cursor < first {
		m.cursor = last
	} else if m.cursor > last {
		m.cursor = first
	}
}

func (m model) View() tea.View {
	s := ""
	s += renderEntries(m.entries, m.cursor)
	s += renderEntry(m.input.View(), m.cursor == 0)
	if m.message != "" {
		s += fmt.Sprintf("  %s\n", m.message)
	}
	return tea.NewView(s)
}

func renderEntries(entries []entry.Entry, cursor int) string {
	const maxEntries = 10

	if len(entries) > maxEntries {
		entries = entries[:maxEntries]
	}

	var s string

	for i := 0; i < maxEntries-len(entries); i++ {
		s += "\n"
	}

	for i := len(entries) - 1; i >= 0; i-- {
		s += renderEntry(entries[i].Label, i+1 == cursor)
	}

	return s
}

func renderEntry(entry string, isSelected bool) string {
	cursor := "  "
	if isSelected {
		cursor = "| "
	}
	return fmt.Sprintf("%s %s\n", cursor, entry)
}
