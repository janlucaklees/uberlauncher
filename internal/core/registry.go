package core

import (
	"uberlauncher/internal/skill"
	"uberlauncher/internal/skills/apps"
	"uberlauncher/internal/skills/system"
	"uberlauncher/internal/skills/todo"
	"uberlauncher/internal/skills/wifi"
)

func RegisterSkills() []skill.Skill {
	return []skill.Skill{
		apps.New(),
		todo.New(),
		system.NewShutdown(),
		system.NewRestart(),
		wifi.New(),
	}
}
