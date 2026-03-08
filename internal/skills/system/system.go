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

func (s *powerSkill) Manifest() skill.Manifest {
	return skill.Manifest{Name: s.name, SupportsFreeText: false}
}

func (s *powerSkill) Start(ctx context.Context, runtime skill.Runtime) error {
	entry := types.EntryDTO{
		SkillName:   s.name,
		EntryID:     s.name,
		DisplayText: s.name,
		IsFreeText:  false,
	}
	runtime.PublishEntries([]types.EntryDTO{entry})
	return nil
}

func (s *powerSkill) Execute(ctx context.Context, cmd types.RunCommandDTO) error {
	if cmd.TriggerType != types.TriggerEntry {
		return errors.New("system skills only support entry triggers")
	}

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

func (s *powerSkill) Stop(ctx context.Context) error {
	return nil
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
