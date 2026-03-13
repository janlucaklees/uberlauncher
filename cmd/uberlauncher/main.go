package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"uberlauncher/internal/cache"
	"uberlauncher/internal/engines"
	"uberlauncher/internal/runtime"
	"uberlauncher/internal/skills"
	"uberlauncher/internal/skills/apps"
	"uberlauncher/internal/skills/search"
	"uberlauncher/internal/skills/system"
	"uberlauncher/internal/skills/todo"
	"uberlauncher/internal/skills/wifi"
	"uberlauncher/internal/store"
	"uberlauncher/internal/ui"
)

// Register your skills here.
var skillList = []skills.Skill{
	apps.New(),
	search.New(),
	todo.New(),
	system.NewShutdown(),
	system.NewRestart(),
	wifi.New(),
}

func main() {
	cache, err := cache.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize cache: %v\n", err)
		os.Exit(1)
	}

	engine := engines.New()

	store := store.New(engine)

	runtime := runtime.New(cache, store)
	runtime.Init(skillList)

	model := ui.New(runtime)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	runtime.Wait()
}
