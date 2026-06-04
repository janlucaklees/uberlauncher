package apps

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/meta"
	"uberlauncher/internal/skill"
)

const cacheVersion = 1

type appEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Exec string `json:"exec"`
}

type appCache struct {
	Version int        `json:"version"`
	Hash    string     `json:"hash"`
	Entries []appEntry `json:"entries"`
}

type AppsSkill struct {
	ctx skill.Context
}

func New() skill.Skill {
	return &AppsSkill{}
}

func (s *AppsSkill) Id() string { return "apps" }

func (s *AppsSkill) Init(ctx skill.Context) {
	s.ctx = ctx

	c, ok := loadCache(ctx.Cache)
	if ok {
		s.upsertApps(c.Entries)
	}

	ctx.Runtime.Go(func() {
		refreshedApps, err := refreshCache(ctx.Cache, c.Hash)
		if err != nil {
			ctx.Notifier.ReportError(err)
			return
		}
		if refreshedApps == nil {
			ctx.Notifier.SendNotification("Apps cache up to date")
			return
		}
		s.upsertApps(refreshedApps)
		ctx.Notifier.SendNotification(fmt.Sprintf("Apps cache refreshed (%d apps)", len(refreshedApps)))
	})
}

func (s *AppsSkill) upsertApps(apps []appEntry) {
	for _, app := range apps {
		execCmd := app.Exec
		s.ctx.Store.UpsertEntry(entry.Entry{
			Key:   app.ID,
			Label: app.Name,
			Run: func(ec entry.Context) {
				if meta.Verbose {
					s.ctx.Notifier.Debug(fmt.Sprintf("Launching \"%s\"", execCmd))
					ec.UI.KeepOpen()
				}

				var err error
				if s.ctx.Runtime.HasCommand("hyprctl") {
					s.ctx.Notifier.Debug("Hyprland detected!")
					// Run (not Start) so we wait for hyprctl to finish sending its
					// socket message to Hyprland before the launcher closes.
					err = exec.Command("hyprctl", "dispatch", "exec", execCmd).Run()
				} else {
					s.ctx.Notifier.Debug("No DE detected, falling back to just executing the command!")
					parts := strings.Fields(execCmd)
					cmd := exec.Command(parts[0], parts[1:]...)
					// Detach from launcher's process group so the app survives after
					// the launcher exits.
					cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
					err = cmd.Start()
				}
				if err != nil {
					s.ctx.Notifier.ReportError(err)
					ec.UI.KeepOpen()
				}
			},
		})
	}
}

func loadCache(c skill.Cache) (appCache, bool) {
	data, err := c.ReadFile("apps.json")
	if err != nil {
		return appCache{}, false
	}
	var cache appCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return appCache{}, false
	}
	if cache.Version != cacheVersion {
		return appCache{}, false
	}
	return cache, true
}

func saveCache(c skill.Cache, entries []appEntry, hash string) error {
	data, err := json.Marshal(appCache{
		Version: cacheVersion,
		Hash:    hash,
		Entries: entries,
	})
	if err != nil {
		return err
	}
	return c.WriteFile("apps.json", data)
}

func refreshCache(c skill.Cache, currentHash string) ([]appEntry, error) {
	paths, err := listDesktopFiles()
	if err != nil {
		return nil, err
	}

	newHash, err := hashDesktopFiles(paths)
	if err != nil {
		return nil, err
	}

	if currentHash == newHash {
		return nil, nil
	}

	entries := parseDesktopFiles(paths)
	if err := saveCache(c, entries, newHash); err != nil {
		return nil, err
	}
	return entries, nil
}
