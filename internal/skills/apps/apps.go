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

	"uberlauncher/internal/cache"
	"uberlauncher/internal/skills"
	"uberlauncher/internal/types"
)

type appEntry struct {
	ID   string
	Name string
	Exec string
}

type Skill struct {
	runtime  skills.Runtime
	commands map[string]string
}

func New() skills.Skill {
	return &Skill{commands: make(map[string]string)}
}

func (s *Skill) Name() string {
	return "apps"
}

func (s *Skill) Init(runtime skills.Runtime) {
	s.runtime = runtime
	sc := runtime.Cache()

	cachedApps, err := loadEntries(sc)
	if err != nil {
		runtime.ReportError(err)
	} else if len(cachedApps) > 0 {
		s.setCommands(cachedApps)
		runtime.UpsertEntries(buildEntries(s.Name(), cachedApps))
	}

	runtime.Go(func() {
		refreshedApps, err := refreshCache(sc)
		if err != nil {
			runtime.ReportError(err)
			return
		}
		if refreshedApps == nil {
			return // hash unchanged, cache still valid
		}
		s.setCommands(refreshedApps)
		runtime.UpsertEntries(buildEntries(s.Name(), refreshedApps))
	})
}

func (s *Skill) Execute(cmd types.Command) {
	if cmd.RawInput == "" && cmd.Entry.EntryID == "" {
		s.runtime.ReportError(errors.New("missing app entry"))
		return
	}

	execCmd := cmd.RawInput
	if execCmd == "" {
		execCmd = s.commands[cmd.Entry.EntryID]
		if execCmd == "" {
			execCmd = cmd.Entry.EntryID
		}
	}

	var err error
	if !hasCommand("hyprctl") {
		err = exec.Command(execCmd).Start()
	} else {
		err = exec.Command("hyprctl", "dispatch", "exec", execCmd).Start()
	}
	if err != nil {
		s.runtime.ReportError(err)
	}
}

func loadEntries(sc *cache.SkillCache) ([]appEntry, error) {
	data, err := sc.ReadFile("apps.tsv")
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	entries := make([]appEntry, 0)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		entryID := ""
		name := ""
		execCmd := ""
		if len(parts) == 2 {
			name = strings.TrimSpace(parts[0])
			execCmd = strings.TrimSpace(parts[1])
			entryID = execCmd
		} else {
			entryID = strings.TrimSpace(parts[0])
			name = strings.TrimSpace(parts[1])
			execCmd = strings.TrimSpace(parts[2])
		}
		if name == "" || execCmd == "" {
			continue
		}
		if entryID == "" {
			entryID = execCmd
		}
		entries = append(entries, appEntry{ID: entryID, Name: name, Exec: execCmd})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func refreshCache(sc *cache.SkillCache) ([]appEntry, error) {
	newHash, err := desktopTreeHash()
	if err != nil {
		return nil, err
	}

	entries, err := listApps()
	if err != nil {
		return nil, err
	}

	tsvData, err := buildTSV(entries)
	if err != nil {
		return nil, err
	}
	if err := sc.WriteFile("apps.tsv", tsvData); err != nil {
		return nil, err
	}
	if err := sc.WriteFile("apps.hash", []byte(newHash)); err != nil {
		return nil, err
	}
	return entries, nil
}

func buildTSV(entries []appEntry) ([]byte, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	seen := make(map[string]struct{})
	for _, entry := range entries {
		key := entry.ID + "\t" + entry.Name + "\t" + entry.Exec
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", entry.ID, entry.Name, entry.Exec); err != nil {
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
	entries := make([]appEntry, 0)
	for _, file := range files {
		entry, ok := parseDesktopFile(file)
		if !ok {
			continue
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

func listDesktopFiles() ([]string, error) {
	roots := []string{"/usr/share/applications", filepath.Join(os.Getenv("HOME"), ".local/share/applications")}
	files := make([]string, 0)
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
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
	defer func() {
		_ = file.Close()
	}()

	var name string
	var execCmd string
	var entryType string
	noDisplay := false
	hidden := false
	terminal := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "Name=") && name == "" {
			name = strings.TrimPrefix(line, "Name=")
		}
		if strings.HasPrefix(line, "Exec=") && execCmd == "" {
			execCmd = strings.TrimPrefix(line, "Exec=")
		}
		if strings.HasPrefix(line, "Type=") && entryType == "" {
			entryType = strings.TrimPrefix(line, "Type=")
		}
		if strings.HasPrefix(strings.ToLower(line), "nodisplay=") {
			noDisplay = strings.TrimPrefix(strings.ToLower(line), "nodisplay=") == "true"
		}
		if strings.HasPrefix(strings.ToLower(line), "hidden=") {
			hidden = strings.TrimPrefix(strings.ToLower(line), "hidden=") == "true"
		}
		if strings.HasPrefix(strings.ToLower(line), "terminal=") {
			terminal = strings.TrimPrefix(strings.ToLower(line), "terminal=") == "true"
		}
	}

	if entryType != "Application" || hidden || noDisplay || terminal || execCmd == "" || name == "" {
		return appEntry{}, false
	}

	execCmd = sanitizeExec(execCmd)
	if execCmd == "" {
		return appEntry{}, false
	}

	return appEntry{ID: path, Name: name, Exec: execCmd}, true
}

func sanitizeExec(cmd string) string {
	replacer := strings.NewReplacer(
		"%%", "\x01",
		"%f", "",
		"%F", "",
		"%u", "",
		"%U", "",
		"%d", "",
		"%D", "",
		"%n", "",
		"%N", "",
		"%i", "",
		"%c", "",
		"%k", "",
		"%v", "",
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

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func buildEntries(skillName string, apps []appEntry) []types.Entry {
	entries := make([]types.Entry, 0, len(apps))
	for _, app := range apps {
		entry := types.NewEntry(skillName, app.ID)
		entry.DisplayText = app.Name
		entries = append(entries, entry)
	}
	return entries
}

func (s *Skill) setCommands(apps []appEntry) {
	commands := make(map[string]string, len(apps))
	for _, app := range apps {
		commands[app.ID] = app.Exec
	}
	s.commands = commands
}
