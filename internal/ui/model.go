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
	message       string
	height        int
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
			m.message = msg.Message
		case runtime.EventError:
			if msg.Err != nil {
				m.message = msg.Err.Error()
			}
		}
		return m, waitForEvent(m.events)
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil
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
		m.message = err.Error()
		return *m, nil
	}

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
	rankedEntryList := m.ranked

	// Shorten the list of entries to the available height
	if len(rankedEntryList) > availableLines {
		rankedEntryList = rankedEntryList[:availableLines]
	}

	entryLineList := make([]string, len(rankedEntryList))
	for i := range rankedEntryList {
		entry := &rankedEntryList[i].Entry
		style := styles.entry

		if m.selectedEntry != nil && entry == m.selectedEntry {
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

	// TODO make this work and lok like selection when free text active.

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
