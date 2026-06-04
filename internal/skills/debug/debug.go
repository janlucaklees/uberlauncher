package debug

import (
	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type DebugSkill struct{}

func New() skill.Skill {
	return &DebugSkill{}
}

func (s *DebugSkill) Id() string { return "debug" }

func (s *DebugSkill) Init(ctx skill.Context) {
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "debug report message",
		Run: func(ec entry.Context) {
			ctx.Notifier.ReportMessage("test")
			ec.UI.KeepOpen()
		},
	})
}
