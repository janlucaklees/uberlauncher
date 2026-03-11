package apps

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"uberlauncher/internal/skill"
	"uberlauncher/internal/types"
)

type appEntry struct {
	Name string
	Exec string
}

type Skill struct{}

func New() skill.Skill {
	return &Skill{}
}

func (s *Skill) Name() string {
	return "apps"
}

func (s *Skill) Init(ctx context.Context, runtime skill.Runtime) error {
	entries, err := loadEntries(s.Name(), runtime.CacheDir())
	if err != nil {
		runtime.ReportError(err)
		return nil
	}
	runtime.PublishEntries(entries)

	if err := spawnRefreshJob(); err != nil {
		runtime.ReportError(err)
	}
	return nil
}

func (s *Skill) Execute(ctx context.Context, cmd types.Command) error {
	if cmd.RawInput == "" && cmd.Entry.EntryID == "" {
		return errors.New("missing app entry")
	}

	execCmd := cmd.RawInput
	if execCmd == "" {
		execCmd = cmd.Entry.EntryID
	}

	if !hasCommand("hyprctl") {
		return exec.CommandContext(ctx, execCmd).Start()
	}

	return exec.CommandContext(ctx, "hyprctl", "dispatch", "exec", execCmd).Start()
}

func (s *Skill) CacheDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = filepath.Join(os.TempDir(), "uberlauncher")
	}
	return filepath.Join(base, "uberlauncher", s.Name())
}

func cachePaths(base string) (string, string) {
	return filepath.Join(base, "apps.tsv"), filepath.Join(base, "apps.hash")
}

func loadEntries(skillName, cacheBase string) ([]types.Entry, error) {
	tsvPath, hashPath := cachePaths(cacheBase)

	if err := os.MkdirAll(cacheBase, 0o755); err != nil {
		return nil, err
	}

	if err := refreshCache(tsvPath, hashPath); err != nil {
		return nil, err
	}

	file, err := os.Open(tsvPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	entries := make([]types.Entry, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		execCmd := strings.TrimSpace(parts[1])
		if name == "" || execCmd == "" {
			continue
		}
		entry := types.NewEntry(skillName, execCmd)
		entry.DisplayText = name

		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func refreshCache(tsvPath, hashPath string) error {
	newHash, err := desktopTreeHash()
	if err != nil {
		return err
	}

	oldHash, _ := os.ReadFile(hashPath)
	if string(oldHash) != newHash || !fileExists(tsvPath) {
		entries, err := listApps()
		if err != nil {
			return err
		}
		if err := writeCache(tsvPath, entries); err != nil {
			return err
		}
		if err := os.WriteFile(hashPath, []byte(newHash), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func writeCache(path string, entries []appEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	w := bufio.NewWriter(file)
	seen := make(map[string]struct{})
	for _, entry := range entries {
		key := entry.Name + "\t" + entry.Exec
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if _, err := fmt.Fprintf(w, "%s\t%s\n", entry.Name, entry.Exec); err != nil {
			return err
		}
	}
	return w.Flush()
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

	return appEntry{Name: name, Exec: execCmd}, true
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

func spawnRefreshJob() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, "__internal", "refresh-app-cache")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
