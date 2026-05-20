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
	enabled, ok := ctx.Config["enabled"].(bool)
	if ok && !enabled {
		return
	}

	searchTemplate, ok := ctx.Config["url"].(string)
	if !ok || strings.TrimSpace(searchTemplate) == "" {
		ctx.Notifier.ReportWarning("Missing config: set skills.search.url in config.toml")
		return
	}
	if !strings.Contains(searchTemplate, "{{query}}") {
		ctx.Notifier.ReportWarning("Config issue: Search URL does not contain {{query}} placeholder")
	}

	baseCommand, ok := ctx.Config["command"].([]string)
	if !ok {
		baseCommand = []string{"xdg-open", "{{url}}"}
	}

	ctx.Store.UpsertEntry(entry.Entry{
		Label: "search",
		Run: func(ec entry.Context) {
			text := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ec.Input), "search "))
			if text == "" {
				ctx.Notifier.ReportError(errors.New("Please provide a search query"))
				ec.UI.KeepOpen()
				return
			}
			url := strings.ReplaceAll(searchTemplate, "{{query}}", url.QueryEscape(text))

			command := make([]string, 0, len(baseCommand))
			for _, s := range baseCommand {
				command = append(command, strings.ReplaceAll(s, "{{url}}", url))
			}

			if err := exec.Command(command[0], command[1:]...).Start(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
}
