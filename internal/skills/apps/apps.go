package apps

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"uberlauncher/internal/entry"
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
	enabled, ok := ctx.Config["enabled"].(bool)
	if ok && !enabled {
		return
	}

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
				var err error
				if s.ctx.Runtime.HasCommand("hyprctl") {
					err = exec.Command("hyprctl", "dispatch", "exec", execCmd).Start()
				} else {
					parts := strings.Fields(execCmd)
					err = exec.Command(parts[0], parts[1:]...).Start()
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
