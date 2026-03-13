package wifi

import (
	"errors"
	"os/exec"
	"strings"

	"uberlauncher/internal/skills"
	"uberlauncher/internal/types"
)

type Skill struct{}

func New() skills.Skill {
	return &Skill{}
}

func (s *Skill) Name() string {
	return "wifi"
}

func (s *Skill) Init(runtime skills.Runtime) error {
	entries := []types.Entry{
		types.NewEntry(s.Name(), "on"),
		types.NewEntry(s.Name(), "off"),
	}

	names, _ := knownConnections()
	for _, name := range names {
		entries = append(entries, types.NewEntry(s.Name(), name))
	}

	runtime.UpsertEntries(entries)
	return nil
}

func (s *Skill) Execute(cmd types.Command) error {
	if !hasCommand("nmcli") {
		return errors.New("nmcli not found")
	}

	action := strings.TrimSpace(cmd.Entry.EntryID)
	action = strings.TrimPrefix(action, "wifi ")

	switch action {
	case "on":
		return exec.Command("nmcli", "radio", "wifi", "on").Run()
	case "off":
		return exec.Command("nmcli", "radio", "wifi", "off").Run()
	default:
		if action == "" {
			return errors.New("missing wifi action")
		}
		return exec.Command("nmcli", "connection", "up", "id", action).Run()
	}
}

func knownConnections() ([]string, error) {
	if !hasCommand("nmcli") {
		return nil, errors.New("nmcli not found")
	}
	cmd := exec.Command("nmcli", "-t", "-f", "NAME,TYPE", "connection", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	names := make([]string, 0)
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		typ := strings.ToLower(parts[1])
		if strings.Contains(typ, "wifi") || strings.Contains(typ, "wireless") {
			names = append(names, name)
		}
	}
	return names, nil
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
