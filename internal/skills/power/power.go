package power

import (
	"errors"
	"os/exec"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type PowerSkill struct{}

func New() skill.Skill {
	return &PowerSkill{}
}

func (s *PowerSkill) Id() string { return "power" }

func (s *PowerSkill) Init(ctx skill.Context) {
	enabled, ok := ctx.Config["enabled"].(bool)
	if ok && !enabled {
		return
	}

	if !ctx.Runtime.HasCommand("powerprofilesctl") {
		ctx.Notifier.ReportError(errors.New("powerprofilesctl not found"))
		return
	}

	ctx.Store.UpsertEntry(entry.Entry{
		Label: "power save",
		Run: func(ec entry.Context) {
			if err := exec.Command("powerprofilesctl", "set", "power-saver").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "power balanced",
		Run: func(ec entry.Context) {
			if err := exec.Command("powerprofilesctl", "set", "balanced").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "power performance",
		Run: func(ec entry.Context) {
			if err := exec.Command("powerprofilesctl", "set", "performance").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
}
