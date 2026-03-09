package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"uberlauncher/internal/ranking"
	"uberlauncher/internal/runtime"
	"uberlauncher/internal/skill"
	"uberlauncher/internal/store"
	"uberlauncher/internal/types"
)

type executeFn func(cmd types.Command) error

type Model struct {
	input         textinput.Model
	store         *store.Store
	usage         *ranking.UsageStore
	events        <-chan runtime.Event
	execute       executeFn
	skillMap      map[string]skill.Skill
	ranked        []ranking.RankedEntry
	selectedEntry *types.Entry
	userSelected  bool
	notify        string
	errMsg        string
}

func New(skillMap map[string]skill.Skill, store *store.Store, usage *ranking.UsageStore, events <-chan runtime.Event, exec executeFn) Model {
	input := textinput.New()
	input.Placeholder = ""
	input.Prompt = "> "
	input.Focus()

	m := Model{
		input:         input,
		store:         store,
		usage:         usage,
		events:        events,
		execute:       exec,
		skillMap:      skillMap,
		selectedEntry: nil,
		userSelected:  false,
	}
	m.refreshEntries()
	return m
}

func (m Model) Init() tea.Cmd {
	return waitForEvent(m.events)
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
		case runtime.EventEntries:
			m.refreshEntries()
		case runtime.EventNotify:
			m.notify = msg.Message
		case runtime.EventError:
			if msg.Err != nil {
				m.errMsg = msg.Err.Error()
			}
		}
		return m, waitForEvent(m.events)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "esc" {
			if m.errMsg != "" {
				m.errMsg = ""
				return m, nil
			}
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
	entries := m.store.ListAll()
	m.ranked = ranking.RankEntries(m.input.Value(), entries, m.usage)
}

func (m *Model) detectFreeText() {
	query := m.input.Value()

	if !strings.Contains(query, " ") {
		return
	}

	first := strings.SplitN(query, " ", 2)[0]
	entry, ok := m.store.Get(store.EntryKey{
		SkillName: first,
		EntryID:   first,
	})

	if ok && entry.SupportsFreeText {
		m.selectedEntry = &entry
	}
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	if m.selectedEntry == nil {
		return *m, nil
	}

	cmd := types.Command{
		Entry:    *m.selectedEntry,
		RawInput: m.input.Value(),
	}

	if err := m.execute(cmd); err != nil {
		m.errMsg = err.Error()
		return *m, nil
	}

	return *m, tea.Quit
}

func (m Model) View() string {
	styles := defaultStyles
	lines := make([]string, 0)

	if m.notify != "" {
		lines = append(lines, styles.notice.Render(m.notify))
	}
	if m.errMsg != "" {
		lines = append(lines, styles.error.Render(m.errMsg))
	}

	entries := m.renderEntries(styles)
	if entries != "" {
		lines = append(lines, entries)
	}

	lines = append(lines, m.renderInput(styles))
	return strings.Join(lines, "\n")
}

func (m Model) renderEntries(styles uiStyles) string {
	if len(m.ranked) == 0 {
		return styles.empty.Render("(no matches)")
	}

	lines := make([]string, 0, len(m.ranked))
	for _, ranked := range m.ranked {
		entry := ranked.Entry
		text := entry.DisplayText
		style := styles.entry.Render(text)

		if m.selectedEntry != nil && m.selectedEntry.SupportsFreeText && entry == *m.selectedEntry {
			style = styles.selected.Render(text)
		}

		lines = append(lines, style)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderInput(styles uiStyles) string {
	input := m.input

	// TODO make this work and lok like selection when free text active.
	return input.View()
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
