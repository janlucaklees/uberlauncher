package system

import (
	"errors"
	"os/exec"

	"uberlauncher/internal/skills"
	"uberlauncher/internal/types"
)

type powerSkill struct {
	runtime skills.Runtime
	name    string
	action  string
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

func (s *powerSkill) Init(runtime skills.Runtime) {
	s.runtime = runtime

	entry := types.NewEntry(s.name, s.name)
	entry.DisplayText = s.name

	runtime.UpsertEntries([]types.Entry{entry})
}

func (s *powerSkill) Execute(cmd types.Command) {
	var err error
	if hasCommand("systemctl") {
		err = exec.Command("systemctl", s.action).Start()
	} else {
		switch s.action {
		case "poweroff":
			err = exec.Command("shutdown", "-h", "now").Start()
		case "reboot":
			err = exec.Command("shutdown", "-r", "now").Start()
		default:
			err = errors.New("unknown system action")
		}
	}
	if err != nil {
		s.runtime.ReportError(err)
	}
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
