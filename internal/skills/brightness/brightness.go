package brightness

import (
	"os/exec"
	"strings"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type Brightness struct{}

func New() skill.Skill {
	return &Brightness{}
}

func (b *Brightness) Id() string { return "brightness" }

func (b *Brightness) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("brightnessctl") {
		return
	}
	upsert(ctx)
}

func upsert(ctx skill.Context) {
	ctx.Store.UpsertEntry(entry.Entry{
		Key:   "brightness",
		Label: label(ctx.Runtime),
		Run: func(ec entry.Context) {
			ec.UI.KeepOpen()
			ec.UI.SetKeyHandler(func(key entry.Key) {
				var err error
				switch key {
				case entry.KeyLeft:
					err = exec.Command("brightnessctl", "set", "5%-").Run()
				case entry.KeyRight:
					err = exec.Command("brightnessctl", "set", "5%+").Run()
				}
				if err != nil {
					ctx.Notifier.ReportError(err)
					return
				}
				upsert(ctx)
			})
		},
	})
}

func label(rt skill.Runtime) string {
	out, err := rt.Command("brightnessctl", "-m").Output()
	if err != nil {
		return "brightness"
	}
	parts := strings.SplitN(strings.TrimSpace(string(out)), ",", 5)
	if len(parts) >= 4 {
		return "brightness " + parts[3]
	}
	return "brightness"
}
