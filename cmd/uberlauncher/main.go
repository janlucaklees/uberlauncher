package main

import (
	"context"
	"fmt"
	"os"

	"uberlauncher/internal/core"
	"uberlauncher/internal/skill"
	"uberlauncher/internal/skills/apps"
	"uberlauncher/internal/skills/system"
	"uberlauncher/internal/skills/todo"
	"uberlauncher/internal/skills/wifi"
)

var skills = []skill.Skill{
	apps.New(),
	todo.New(),
	system.NewShutdown(),
	system.NewRestart(),
	wifi.New(),
}

func main() {
	if err := core.Run(context.Background(), skills); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
