package apps

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type appEntry struct {
	ID   string
	Name string
	Exec string
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

	cachedApps, err := loadEntries(ctx.Cache)
	if err != nil {
		ctx.Notifier.ReportError(err)
	} else {
		s.upsertApps(cachedApps)
	}

	ctx.Runtime.Go(func() {
		refreshedApps, err := refreshCache(ctx.Cache)
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

func loadEntries(c skill.Cache) ([]appEntry, error) {
	data, err := c.ReadFile("apps.tsv")
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var entries []appEntry
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		id := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		execCmd := strings.TrimSpace(parts[2])
		if id == "" || name == "" || execCmd == "" {
			continue
		}
		entries = append(entries, appEntry{ID: id, Name: name, Exec: execCmd})
	}
	return entries, scanner.Err()
}

func refreshCache(c skill.Cache) ([]appEntry, error) {
	paths, err := listDesktopFiles()
	if err != nil {
		return nil, err
	}

	newHash, err := hashDesktopFiles(paths)
	if err != nil {
		return nil, err
	}

	oldHash, _ := c.ReadFile("apps.hash")
	if strings.TrimSpace(string(oldHash)) == newHash {
		return nil, nil
	}

	entries := parseDesktopFiles(paths)

	tsvData, err := buildTSV(entries)
	if err != nil {
		return nil, err
	}
	if err := c.WriteFile("apps.tsv", tsvData); err != nil {
		return nil, err
	}
	if err := c.WriteFile("apps.hash", []byte(newHash)); err != nil {
		return nil, err
	}

	return entries, nil
}

func buildTSV(entries []appEntry) ([]byte, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	for _, e := range entries {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", e.ID, e.Name, e.Exec); err != nil {
			return nil, err
		}
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
