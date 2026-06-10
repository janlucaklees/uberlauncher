package power

import (
	"errors"
	"os/exec"
	"strings"
	"time"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type PowerSkill struct{}

func New() skill.Skill {
	return &PowerSkill{}
}

func (s *PowerSkill) Id() string { return "power" }

func (s *PowerSkill) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("powerprofilesctl") {
		ctx.Notifier.ReportError(errors.New("powerprofilesctl not found"))
		return
	}

	ctx.Status.Register("power:icon", 10*time.Second, func() string {
		out, err := ctx.Runtime.Command("powerprofilesctl", "get").Output()
		if err != nil {
			return ""
		}
		switch strings.TrimSpace(string(out)) {
		case "power-saver":
			return "󰌪"
		case "balanced":
			return "󰈐"
		case "performance":
			return "󱐋"
		default:
			return ""
		}
	})

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
