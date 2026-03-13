package search

import (
	"errors"
	"net/url"
	"os/exec"
	"strings"

	"uberlauncher/internal/skills"
	"uberlauncher/internal/types"
)

type Skill struct {
	runtime skills.Runtime
}

func New() skills.Skill {
	return &Skill{}
}

func (s *Skill) Name() string {
	return "search"
}

func (s *Skill) Init(runtime skills.Runtime) {
	s.runtime = runtime

	entry := types.NewEntry(s.Name(), s.Name())
	entry.DisplayText = s.Name()
	entry.SupportsFreeText = true

	runtime.UpsertEntries([]types.Entry{entry})
}

func (s *Skill) Execute(cmd types.Command) {
	text := strings.TrimSpace(cmd.RawInput)
	text = strings.TrimSpace(strings.TrimPrefix(text, "search"))
	if text == "" {
		s.runtime.ReportError(errors.New("search query is empty"))
		return
	}

	searchURL := "https://www.google.com/search?q=" + url.QueryEscape(text)
	if err := exec.Command("xdg-open", searchURL).Start(); err != nil {
		s.runtime.ReportError(err)
	}
}
