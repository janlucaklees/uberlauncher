package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"uberlauncher/internal/cache"
	"uberlauncher/internal/config"
	"uberlauncher/internal/engines"
	"uberlauncher/internal/meta"
	"uberlauncher/internal/notifier"
	"uberlauncher/internal/runtime"
	"uberlauncher/internal/skill"
	"uberlauncher/internal/skills/apps"
	"uberlauncher/internal/skills/audio"
	"uberlauncher/internal/skills/battery"
	"uberlauncher/internal/skills/bluetooth"
	"uberlauncher/internal/skills/brightness"
	"uberlauncher/internal/skills/clock"
	"uberlauncher/internal/skills/custom"
	"uberlauncher/internal/skills/debug"
	"uberlauncher/internal/skills/display"
	"uberlauncher/internal/skills/emoji"
	"uberlauncher/internal/skills/keyboard"
	"uberlauncher/internal/skills/notifications"
	"uberlauncher/internal/skills/power"
	"uberlauncher/internal/skills/search"
	"uberlauncher/internal/skills/system"
	"uberlauncher/internal/skills/todoist"
	"uberlauncher/internal/skills/vpn"
	"uberlauncher/internal/skills/wifi"
	"uberlauncher/internal/statusbar"
	"uberlauncher/internal/store"
	"uberlauncher/internal/ui"
)

// Register your skills here.
var skillList = []skill.Skill{
	apps.New(),
	audio.New(),
	battery.New(),
	bluetooth.New(),
	brightness.New(),
	clock.New(),
	custom.New(),
	display.New(),
	emoji.New(),
	keyboard.New(),
	notifications.New(),
	power.New(),
	search.New(),
	system.New(),
	todoist.New(),
	vpn.New(),
	wifi.New(),
}

func main() {
	verbose := flag.Bool("v", false, "enable debug output")
	flag.BoolVar(verbose, "verbose", false, "enable debug output")
	configPath := flag.String("config", "", "path to config file (overrides default; file is not auto-created)")
	flag.Parse()

	meta.Verbose = *verbose
	meta.ConfigPath = *configPath

	if meta.Verbose {
		skillList = append(skillList, debug.New())
	}

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

	ctx, cancel := context.WithCancel(context.Background())
	sbStore := statusbar.NewStore()
	sbUpdater := statusbar.NewUpdater(ctx, sbStore)
	sbCfg := cfg.GetStatusBarConfig()

	for _, s := range skillList {
		s := s
		skillCfg := cfg.GetForSkill(s)
		if !skillCfg.IsEnabled() {
			continue
		}
		rt.Go(func() {
			s.Init(skill.Context{
				Runtime:  rt,
				Notifier: n,
				Status:   sbUpdater,
				Store:    st.GetForSkill(s),
				Cache:    c.GetForSkill(s),
				Config:   skillCfg,
			})
		})
	}

	model := ui.New(st, n, ui.StatusBarOptions{
		Enabled: sbCfg.Enabled,
		Store:   sbStore,
		Left:    sbCfg.Left,
		Center:  sbCfg.Center,
		Right:   sbCfg.Right,
	})

	program := tea.NewProgram(model)
	_, err = program.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	cancel()
	rt.Shutdown()
	rt.Wait()
}
