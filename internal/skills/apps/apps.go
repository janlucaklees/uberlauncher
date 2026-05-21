package apps

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rkoesters/xdg/desktop"

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
		if len(parts) < 2 {
			continue
		}
		var id, name, execCmd string
		if len(parts) == 2 {
			name = strings.TrimSpace(parts[0])
			execCmd = strings.TrimSpace(parts[1])
			id = execCmd
		} else {
			id = strings.TrimSpace(parts[0])
			name = strings.TrimSpace(parts[1])
			execCmd = strings.TrimSpace(parts[2])
		}
		if name == "" || execCmd == "" {
			continue
		}
		if id == "" {
			id = execCmd
		}
		entries = append(entries, appEntry{ID: id, Name: name, Exec: execCmd})
	}
	return entries, scanner.Err()
}

func refreshCache(c skill.Cache) ([]appEntry, error) {
	newHash, err := desktopTreeHash()
	if err != nil {
		return nil, err
	}

	oldHash, _ := c.ReadFile("apps.hash")
	if strings.TrimSpace(string(oldHash)) == newHash {
		return nil, nil
	}

	entries, err := listApps()
	if err != nil {
		return nil, err
	}

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
	seen := make(map[string]struct{})
	for _, e := range entries {
		key := e.ID + "\t" + e.Name + "\t" + e.Exec
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", e.ID, e.Name, e.Exec); err != nil {
			return nil, err
		}
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func listApps() ([]appEntry, error) {
	files, err := listDesktopFiles()
	if err != nil {
		return nil, err
	}
	var entries []appEntry
	for _, file := range files {
		e, ok := parseDesktopFile(file)
		if !ok {
			continue
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

func listDesktopFiles() ([]string, error) {
	roots := []string{"/usr/share/applications", filepath.Join(os.Getenv("HOME"), ".local/share/applications")}
	var files []string
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(d.Name(), ".desktop") {
				files = append(files, path)
			}
			return nil
		})
	}
	return files, nil
}

func parseDesktopFile(path string) (appEntry, bool) {
	file, err := os.Open(path)
	if err != nil {
		return appEntry{}, false
	}
	defer func() { _ = file.Close() }()

	e, err := desktop.New(file)
	if err != nil {
		return appEntry{}, false
	}

	if e.Type != desktop.Application || e.Hidden || e.NoDisplay || e.Terminal || e.Exec == "" || e.Name == "" {
		return appEntry{}, false
	}

	execCmd := sanitizeExec(e.Exec)
	if execCmd == "" {
		return appEntry{}, false
	}
	return appEntry{ID: path, Name: e.Name, Exec: execCmd}, true
}

func sanitizeExec(cmd string) string {
	replacer := strings.NewReplacer(
		"%%", "\x01",
		"%f", "", "%F", "",
		"%u", "", "%U", "",
		"%d", "", "%D", "",
		"%n", "", "%N", "",
		"%i", "", "%c", "",
		"%k", "", "%v", "",
		"%m", "",
	)
	cleaned := replacer.Replace(cmd)
	cleaned = strings.ReplaceAll(cleaned, "\x01", "%")
	return strings.TrimSpace(cleaned)
}

func desktopTreeHash() (string, error) {
	files, err := listDesktopFiles()
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	h := sha256.New()
	for _, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if _, err := fmt.Fprintf(h, "%s\t%d\t%d\n", path, info.ModTime().Unix(), info.Size()); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
