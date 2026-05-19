package search

import (
	"errors"
	"net/url"
	"os/exec"
	"strings"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type SearchSkill struct{}

func New() skill.Skill {
	return &SearchSkill{}
}

func (s *SearchSkill) Id() string { return "search" }

func (s *SearchSkill) Init(ctx skill.Context) {
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "search",
		Run: func(ec entry.Context) {
			text := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ec.Input), "search"))
			if text == "" {
				ctx.Notifier.ReportError(errors.New("search query is empty"))
				return
			}
			searchURL := "https://www.google.com/search?q=" + url.QueryEscape(text)
			if err := exec.Command("xdg-open", searchURL).Start(); err != nil {
				ctx.Notifier.ReportError(err)
			}
		},
	})
}
