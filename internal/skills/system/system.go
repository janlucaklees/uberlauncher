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
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "system shutdown",
		Run: func(ec entry.Context) {
			var err error
			if ctx.Runtime.HasCommand("systemctl") {
				err = exec.Command("systemctl", "poweroff").Start()
			} else {
				err = exec.Command("shutdown", "-h", "now").Start()
			}
			if err != nil {
				ctx.Notifier.ReportError(err)
			}
		},
	})
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "system reboot",
		Run: func(ec entry.Context) {
			var err error
			if ctx.Runtime.HasCommand("systemctl") {
				err = exec.Command("systemctl", "reboot").Start()
			} else {
				err = exec.Command("shutdown", "-r", "now").Start()
			}
			if err != nil {
				ctx.Notifier.ReportError(err)
			}
		},
	})
}
