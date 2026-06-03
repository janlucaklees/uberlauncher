package clock

import (
	gtime "time"
	"uberlauncher/internal/skill"
)

type ClockSkill struct{}

func New() skill.Skill {
	return &ClockSkill{}
}

func (s *ClockSkill) Id() string { return "clock" }

func (s *ClockSkill) Init(ctx skill.Context) {
	enabled, ok := ctx.Config["enabled"].(bool)
	if ok && !enabled {
		return
	}

	format := "15:04"
	if f, ok := ctx.Config["format"].(string); ok && f != "" {
		format = f
	}

	ctx.Status.Register("time", 10*gtime.Second, func() string {
		return gtime.Now().Format(format)
	})
}
