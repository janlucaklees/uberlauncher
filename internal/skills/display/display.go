package display

import (
	"errors"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type DisplaySkill struct{}

func New() skill.Skill {
	return &DisplaySkill{}
}

func (s *DisplaySkill) Id() string { return "display" }

func (s *DisplaySkill) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("hyprctl") {
		ctx.Notifier.ReportError(errors.New("hyprctl not found"))
		return
	}

	ctx.Store.UpsertEntry(entry.Entry{
		Label: "display off",
		Run: func(ec entry.Context) {
			if err := ctx.Runtime.Command("hyprctl", "dispatch", `hl.dsp.dpms({ action = "disable" })`).Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "display on",
		Run: func(ec entry.Context) {
			if err := ctx.Runtime.Command("hyprctl", "dispatch", `hl.dsp.dpms({ action = "enable" })`).Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
}
