package apps

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rkoesters/xdg/desktop"
)

func listDesktopFiles() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	roots := []string{"/usr/share/applications", filepath.Join(homeDir, ".local/share/applications")}

	// User root is last, so same-basename files in ~/.local override /usr/share.
	byID := make(map[string]string)
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(d.Name(), ".desktop") {
				byID[d.Name()] = path
			}
			return nil
		})
	}

	paths := make([]string, 0, len(byID))
	for _, path := range byID {
		paths = append(paths, path)
	}
	return paths, nil
}

func hashDesktopFiles(paths []string) (string, error) {
	// Make sure the paths are sorted for deterministic hashing.
	sort.Strings(paths)

	h := sha256.New()
	for _, path := range paths {
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

func parseDesktopFiles(paths []string) []appEntry {
	var entries []appEntry
	for _, path := range paths {
		if e, ok := parseDesktopFile(path); ok {
			entries = append(entries, e)
		}
	}
	return entries
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
	var b strings.Builder
	for i := 0; i < len(cmd); i++ {
		if cmd[i] != '%' || i+1 >= len(cmd) {
			b.WriteByte(cmd[i])
			continue
		}
		next := cmd[i+1]
		if next == '%' {
			b.WriteByte('%')
		} else if next >= 'a' && next <= 'z' || next >= 'A' && next <= 'Z' {
			// field code: expand to nothing
		} else {
			b.WriteByte(cmd[i])
			continue
		}
		i++
	}
	return strings.TrimSpace(b.String())
}
