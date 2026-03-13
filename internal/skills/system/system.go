package system

import (
	"errors"
	"os/exec"

	"uberlauncher/internal/skills"
	"uberlauncher/internal/types"
)

type powerSkill struct {
	name   string
	action string
}

func NewShutdown() skills.Skill {
	return &powerSkill{name: "shutdown", action: "poweroff"}
}

func NewRestart() skills.Skill {
	return &powerSkill{name: "restart", action: "reboot"}
}

func (s *powerSkill) Name() string {
	return s.name
}

func (s *powerSkill) Init(runtime skills.Runtime) error {
	entry := types.NewEntry(s.name, s.name)
	entry.DisplayText = s.name

	runtime.UpsertEntries([]types.Entry{entry})
	return nil
}

func (s *powerSkill) Execute(cmd types.Command) error {
	if hasCommand("systemctl") {
		return exec.Command("systemctl", s.action).Start()
	}

	switch s.action {
	case "poweroff":
		return exec.Command("shutdown", "-h", "now").Start()
	case "reboot":
		return exec.Command("shutdown", "-r", "now").Start()
	default:
		return errors.New("unknown system action")
	}
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
