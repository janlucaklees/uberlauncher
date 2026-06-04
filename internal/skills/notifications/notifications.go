package notifications

import (
	"errors"
	"os/exec"
	"strings"
	"time"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type NotificationsSkill struct{}

func New() skill.Skill {
	return &NotificationsSkill{}
}

func (s *NotificationsSkill) Id() string { return "notifications" }

func (s *NotificationsSkill) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("makoctl") {
		ctx.Notifier.ReportError(errors.New("makoctl not found"))
		return
	}

	ctx.Status.Register("notification:icon", 5*time.Second, func() string {
		return notificationIcon(ctx.Runtime)
	})

	ctx.Store.UpsertEntry(entry.Entry{
		Label: "notifications on",
		Run: func(ec entry.Context) {
			if err := exec.Command("makoctl", "mode", "-r", "hide").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "notifications off",
		Run: func(ec entry.Context) {
			if err := exec.Command("makoctl", "mode", "-a", "hide").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
}

func notificationIcon(rt skill.Runtime) string {
	out, err := rt.Command("makoctl", "mode").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "hide" {
			return "󰂛"
		}
	}
	return "󰂚"
}
