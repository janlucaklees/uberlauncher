package ui

import (
	"fmt"
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

type executeFn func(cmd types.RunCommandDTO) error

type Model struct {
	input        textinput.Model
	store        *store.Store
	usage        *ranking.UsageStore
	events       <-chan runtime.Event
	execute      executeFn
	skillMap     map[string]skill.Manifest
	ranked       []ranking.RankedEntry
	selectedIdx  int
	userSelected bool
	freeText     bool
	freeTextName string
	notify       string
	errMsg       string
}

func New(store *store.Store, usage *ranking.UsageStore, events <-chan runtime.Event, skills []skill.Manifest, exec executeFn) Model {
	input := textinput.New()
	input.Placeholder = ""
	input.Prompt = "> "
	input.Focus()

	skillMap := make(map[string]skill.Manifest)
	for _, manifest := range skills {
		skillMap[manifest.Name] = manifest
	}

	m := Model{
		input:        input,
		store:        store,
		usage:        usage,
		events:       events,
		execute:      exec,
		skillMap:     skillMap,
		selectedIdx:  0,
		userSelected: false,
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
			m.refreshEntriesPreserveSelection()
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
		case "up", "k":
			if len(m.ranked) > 0 {
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
				m.userSelected = true
				m.freeText = false
			}
			return m, nil
		case "down", "j":
			if len(m.ranked) > 0 {
				if m.selectedIdx < len(m.ranked)-1 {
					m.selectedIdx++
				}
				m.userSelected = true
				m.freeText = false
			}
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
	m.userSelected = false
	m.refreshEntries()
	m.detectFreeText()
	if !m.userSelected {
		m.selectedIdx = 0
	}
}

func (m *Model) refreshEntries() {
	entries := m.store.ListAll()
	m.ranked = ranking.RankEntries(m.input.Value(), entries, m.usage)
}

func (m *Model) refreshEntriesPreserveSelection() {
	prevKey := m.selectedKey()
	m.refreshEntries()
	if m.userSelected && prevKey.SkillName != "" {
		for idx, ranked := range m.ranked {
			if ranked.Entry.SkillName == prevKey.SkillName && ranked.Entry.EntryID == prevKey.EntryID {
				m.selectedIdx = idx
				return
			}
		}
		m.selectedIdx = 0
	}
}

func (m *Model) detectFreeText() {
	m.freeText = false
	m.freeTextName = ""
	query := m.input.Value()
	if strings.TrimLeft(query, " ") != query {
		return
	}
	if !strings.Contains(query, " ") {
		return
	}
	first := strings.SplitN(query, " ", 2)[0]
	if manifest, ok := m.skillMap[first]; ok && manifest.SupportsFreeText {
		m.freeText = true
		m.freeTextName = first
	}
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	if m.freeText {
		cmd := types.RunCommandDTO{
			SkillName:   m.freeTextName,
			EntryID:     m.freeTextName,
			RawInput:    m.input.Value(),
			TriggerType: types.TriggerRawInput,
		}
		if err := m.execute(cmd); err != nil {
			m.errMsg = err.Error()
			return *m, nil
		}
		return *m, tea.Quit
	}

	if len(m.ranked) == 0 {
		return *m, nil
	}

	selected := m.ranked[m.selectedIdx].Entry
	cmd := types.RunCommandDTO{
		SkillName:   selected.SkillName,
		EntryID:     selected.EntryID,
		RawInput:    selected.EntryID,
		TriggerType: types.TriggerEntry,
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
	for idx, ranked := range m.ranked {
		entry := ranked.Entry
		text := entry.DisplayText
		if idx == m.selectedIdx && !m.freeText {
			lines = append(lines, styles.selected.Render(text))
		} else {
			lines = append(lines, styles.entry.Render(text))
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderInput(styles uiStyles) string {
	input := m.input
	if first := firstToken(input.Value()); first != "" {
		if _, ok := m.skillMap[first]; ok {
			input.PromptStyle = styles.promptActive
		}
	}
	if m.freeText {
		return fmt.Sprintf("%s %s", styles.freeText.Render("[free-text]"), input.View())
	}
	return input.View()
}

func (m Model) selectedKey() store.EntryKey {
	if len(m.ranked) == 0 || m.selectedIdx < 0 || m.selectedIdx >= len(m.ranked) {
		return store.EntryKey{}
	}
	entry := m.ranked[m.selectedIdx].Entry
	return store.EntryKey{SkillName: entry.SkillName, EntryID: entry.EntryID}
}

func firstToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.SplitN(value, " ", 2)[0]
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
