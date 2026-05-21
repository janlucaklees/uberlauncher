package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"uberlauncher/internal/cache"
	"uberlauncher/internal/config"
	"uberlauncher/internal/engines"
	"uberlauncher/internal/notifier"
	"uberlauncher/internal/runtime"
	"uberlauncher/internal/skill"
	"uberlauncher/internal/skills/apps"
	"uberlauncher/internal/skills/search"
	"uberlauncher/internal/skills/system"
	"uberlauncher/internal/skills/todoist"
	"uberlauncher/internal/skills/wifi"
	"uberlauncher/internal/store"
	"uberlauncher/internal/ui"
)

// Register your skills here.
var skillList = []skill.Skill{
	apps.New(),
	search.New(),
	system.New(),
	todoist.New(),
	wifi.New(),
}

func main() {
	rt := runtime.New()
	n := notifier.New()
	engine := engines.New()
	st := store.New(engine)

	c, err := cache.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize cache: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	for _, s := range skillList {
		s.Init(skill.Context{
			Runtime:  rt,
			Notifier: n,

			Store:  st.GetForSkill(s),
			Cache:  c.GetForSkill(s),
			Config: cfg.GetForSkill(s),
		})
	}

	model := ui.New(st, n)

	program := tea.NewProgram(model)
	_, err = program.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	rt.Wait()
}
