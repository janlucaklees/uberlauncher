package custom

import (
	"os/exec"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type CustomSkill struct{}

func New() skill.Skill {
	return &CustomSkill{}
}

func (s *CustomSkill) Id() string { return "custom" }

func (s *CustomSkill) Init(ctx skill.Context) {
	enabled, ok := ctx.Config["enabled"].(bool)
	if ok && !enabled {
		return
	}

	for key, val := range ctx.Config {
		def, ok := val.(map[string]any)
		if !ok {
			continue
		}

		label, ok := def["label"].(string)
		if !ok || label == "" {
			continue
		}

		command, ok := def["command"].(string)
		if !ok || command == "" {
			continue
		}

		ctx.Store.UpsertEntry(entry.Entry{
			Key:   key,
			Label: label,
			Run: func(ec entry.Context) {
				err := exec.Command("sh", "-c", command).Start()
				if err != nil {
					ctx.Notifier.ReportError(err)
					ec.UI.KeepOpen()
				}
			},
		})
	}
}
