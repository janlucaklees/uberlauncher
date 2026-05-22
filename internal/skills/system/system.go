package system

import (
	"os/exec"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type SystemSkill struct{}

func New() skill.Skill {
	return &SystemSkill{}
}

func (s *SystemSkill) Id() string { return "system" }

func (s *SystemSkill) Init(ctx skill.Context) {
	enabled, ok := ctx.Config["enabled"].(bool)
	if ok && !enabled {
		return
	}

	ctx.Store.UpsertEntry(entry.Entry{
		Label: "system shutdown",
		Run: func(ec entry.Context) {
			var err error
			if ctx.Runtime.HasCommand("systemctl") {
				err = exec.Command("systemctl", "poweroff").Run()
			} else {
				err = exec.Command("shutdown", "-h", "now").Run()
			}
			if err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "system reboot",
		Run: func(ec entry.Context) {
			var err error
			if ctx.Runtime.HasCommand("systemctl") {
				err = exec.Command("systemctl", "reboot").Run()
			} else {
				err = exec.Command("shutdown", "-r", "now").Run()
			}
			if err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
}
