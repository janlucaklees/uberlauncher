package system

import (
	"context"
	"errors"
	"os/exec"

	"uberlauncher/internal/skill"
	"uberlauncher/internal/types"
)

type powerSkill struct {
	name   string
	action string
}

func NewShutdown() skill.Skill {
	return &powerSkill{name: "shutdown", action: "poweroff"}
}

func NewRestart() skill.Skill {
	return &powerSkill{name: "restart", action: "reboot"}
}

func (s *powerSkill) Name() string {
	return s.name
}

func (s *powerSkill) Init(ctx context.Context, runtime skill.Runtime) error {
	entry := types.Entry{
		SkillName: s.name,
		EntryID:   s.name,
	}
	runtime.PublishEntries([]types.Entry{entry})
	return nil
}

func (s *powerSkill) Execute(ctx context.Context, cmd types.Command) error {
	if hasCommand("systemctl") {
		return exec.CommandContext(ctx, "systemctl", s.action).Start()
	}

	switch s.action {
	case "poweroff":
		return exec.CommandContext(ctx, "shutdown", "-h", "now").Start()
	case "reboot":
		return exec.CommandContext(ctx, "shutdown", "-r", "now").Start()
	default:
		return errors.New("unknown system action")
	}
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
