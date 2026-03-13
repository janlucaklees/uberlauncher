package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"uberlauncher/internal/runtime"
	"uberlauncher/internal/store"
	"uberlauncher/internal/types"
)

type Model struct {
	input         textinput.Model
	runtime       *runtime.Runtime
	ranked        []types.Entry
	selectedEntry *types.Entry
	message       string
	height        int
}

func New(rt *runtime.Runtime) Model {
	input := textinput.New()
	input.Placeholder = ""
	input.Prompt = "> "
	input.Focus()

	m := Model{
		input:   input,
		runtime: rt,
	}
	m.refreshEntries()
	return m
}

func (m Model) Init() tea.Cmd {
	return waitForEvent(m.runtime.Channel)
}

type eventMsg runtime.Event

func waitForEvent(ch <-chan runtime.Event) tea.Cmd {
	return func() tea.Msg {
		return eventMsg(<-ch)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case eventMsg:
		switch msg.Type {
		case runtime.EventNewEntry:
			m.refreshEntries()
		case runtime.EventMessage:
			m.message = msg.Message
		case runtime.EventError:
			if msg.Err != nil {
				m.message = msg.Err.Error()
			}
		}
		return m, waitForEvent(m.runtime.Channel)
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "esc" {
			return m, tea.Quit
		}

		switch msg.String() {
		case "up":
			// Todo: implement
			return m, nil
		case "down":
			// Todo: implement
			return m, nil
		case "enter":
			return m.handleEnter()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.onInputChanged()
	return m, cmd
}

func (m *Model) onInputChanged() {
	m.refreshEntries()
	m.detectFreeText()
}

func (m *Model) refreshEntries() {
	m.ranked = m.runtime.Store.GetMatches(m.input.Value())
}

func (m *Model) detectFreeText() {
	query := m.input.Value()

	if !strings.Contains(query, " ") {
		return
	}

	first := strings.SplitN(query, " ", 2)[0]
	id := store.BuildGlobalEntryId(types.Entry{SkillName: first, EntryID: first})
	if entry, ok := m.runtime.Store.GetEntry(id); ok && entry.SupportsFreeText {
		m.selectedEntry = &entry
	}
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	if m.selectedEntry == nil {
		return *m, nil
	}

	skill, ok := m.runtime.SkillMap[m.selectedEntry.SkillName]
	if !ok {
		m.message = "skill not found: " + m.selectedEntry.SkillName
		return *m, nil
	}

	cmd := types.Command{
		Entry:    *m.selectedEntry,
		RawInput: m.input.Value(),
	}

	skill.Execute(cmd)

	return *m, tea.Quit
}

func (m Model) View() string {
	styles := defaultStyles

	if m.height <= 3 {
		return styles.empty.Render("(no space)")
	}

	lines := make([]string, 0)

	lines = m.appendEntries(lines, m.height-2, styles)
	lines = m.appendInput(lines, 1, styles)
	lines = m.appendMessage(lines, 1, styles)

	return strings.Join(lines, "\n")
}

func (m Model) appendEntries(lines []string, availableLines int, styles uiStyles) []string {
	entries := m.ranked

	// Shorten the list of entries to the available height
	if len(entries) > availableLines {
		entries = entries[:availableLines]
	}

	entryLineList := make([]string, len(entries))
	for i := range entries {
		entry := &entries[i]
		style := styles.entry

		if m.selectedEntry != nil &&
			entry.SkillName == m.selectedEntry.SkillName &&
			entry.EntryID == m.selectedEntry.EntryID {
			style = styles.selected
		}

		entryLineList[i] = style.Render(entry.DisplayText)
	}

	// If no entries render message
	if len(entryLineList) == 0 {
		entryLineList = []string{styles.empty.Render("(no matches)")}
	}

	numberOfEmptyLines := availableLines - len(entryLineList)
	lines = append(lines, make([]string, numberOfEmptyLines)...)
	lines = append(lines, entryLineList...)

	return lines
}

func (m Model) appendInput(lines []string, availableLines int, styles uiStyles) []string {
	input := m.input

	// TODO make this work and look like selection when free text active.

	return append(lines, input.View())
}

func (m Model) appendMessage(lines []string, availableLines int, styles uiStyles) []string {
	if m.message == "" {
		return append(lines, "")
	}

	return append(lines, styles.notice.Render(m.message))
}

type uiStyles struct {
	selected     lipgloss.Style
	entry        lipgloss.Style
	empty        lipgloss.Style
	promptActive lipgloss.Style
	freeText     lipgloss.Style
	notice       lipgloss.Style
	error        lipgloss.Style
}

var defaultStyles = uiStyles{
	selected:     lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")),
	entry:        lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
	empty:        lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	promptActive: lipgloss.NewStyle().Foreground(lipgloss.Color("141")),
	freeText:     lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
	notice:       lipgloss.NewStyle().Foreground(lipgloss.Color("72")),
	error:        lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
}
