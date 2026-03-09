package core

import (
	"context"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"uberlauncher/internal/ranking"
	"uberlauncher/internal/runtime"
	"uberlauncher/internal/skill"
	"uberlauncher/internal/store"
	"uberlauncher/internal/types"
	"uberlauncher/internal/ui"
)

func Run(ctx context.Context, skills []skill.Skill) error {

	entryStore := store.New()
	manager := runtime.NewManager(entryStore)

	cacheBase, err := os.UserCacheDir()
	if err != nil {
		cacheBase = filepath.Join(os.TempDir(), "uberlauncher")
	}
	usageStore := ranking.NewUsageStore(filepath.Join(cacheBase, "uberlauncher"))
	_ = usageStore.Load()

	skillMap := make(map[string]skill.Skill)

	for _, skill := range skills {
		name := skill.Name()
		skillMap[name] = skill
		// TODO rework
		runtime := manager.ForSkill(name)
		if err := skill.Init(ctx, runtime); err != nil {
			runtime.ReportError(err)
		}
	}

	execFn := func(cmd types.Command) error {
		skill, ok := skillMap[cmd.Entry.SkillName]
		if !ok {
			return nil
		}
		if err := skill.Execute(ctx, cmd); err != nil {
			return err
		}
		entryID := cmd.Entry.EntryID
		if entryID == "" {
			entryID = cmd.Entry.SkillName
		}
		usageKey := ranking.UsageKey(cmd.Entry.SkillName, entryID)
		usageStore.Bump(usageKey, time.Now())
		_ = usageStore.Save()
		return nil
	}

	// Todo: rework. the model should not hold the core logic, but the logic for presentation.
	model := ui.New(skillMap, entryStore, usageStore, manager.Events, execFn)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}
