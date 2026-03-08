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

func Run(ctx context.Context) error {
	entryStore := store.New()
	manager := runtime.NewManager(entryStore)

	cacheBase, err := os.UserCacheDir()
	if err != nil {
		cacheBase = filepath.Join(os.TempDir(), "uberlauncher")
	}
	usageStore := ranking.NewUsageStore(filepath.Join(cacheBase, "uberlauncher"))
	_ = usageStore.Load()

	skills := RegisterSkills()
	manifests := make([]skill.Manifest, 0, len(skills))
	skillMap := make(map[string]skill.Skill)

	for _, skill := range skills {
		manifest := skill.Manifest()
		manifests = append(manifests, manifest)
		skillMap[manifest.Name] = skill
		runtime := manager.ForSkill(manifest.Name)
		if err := skill.Start(ctx, runtime); err != nil {
			runtime.ReportError(err)
		}
	}

	execFn := func(cmd types.RunCommandDTO) error {
		skill, ok := skillMap[cmd.SkillName]
		if !ok {
			return nil
		}
		if err := skill.Execute(ctx, cmd); err != nil {
			return err
		}
		entryID := cmd.EntryID
		if cmd.TriggerType == types.TriggerRawInput {
			entryID = cmd.SkillName
		}
		usageKey := ranking.UsageKey(cmd.SkillName, entryID)
		usageStore.Bump(usageKey, now())
		_ = usageStore.Save()
		return nil
	}

	model := ui.New(entryStore, usageStore, manager.Events, manifests, execFn)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func now() time.Time {
	return time.Now()
}
