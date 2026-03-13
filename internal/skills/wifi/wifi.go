package wifi

import (
	"errors"
	"os/exec"
	"strings"

	"uberlauncher/internal/skills"
	"uberlauncher/internal/types"
)

type Skill struct {
	runtime skills.Runtime
}

func New() skills.Skill {
	return &Skill{}
}

func (s *Skill) Name() string {
	return "wifi"
}

func (s *Skill) Init(runtime skills.Runtime) {
	s.runtime = runtime

	entries := []types.Entry{
		types.NewEntry(s.Name(), "on"),
		types.NewEntry(s.Name(), "off"),
	}

	names, _ := knownConnections()
	for _, name := range names {
		entries = append(entries, types.NewEntry(s.Name(), name))
	}

	runtime.UpsertEntries(entries)
}

func (s *Skill) Execute(cmd types.Command) {
	if !hasCommand("nmcli") {
		s.runtime.ReportError(errors.New("nmcli not found"))
		return
	}

	action := strings.TrimSpace(cmd.Entry.EntryID)
	action = strings.TrimPrefix(action, "wifi ")

	var err error
	switch action {
	case "on":
		err = exec.Command("nmcli", "radio", "wifi", "on").Run()
	case "off":
		err = exec.Command("nmcli", "radio", "wifi", "off").Run()
	default:
		if action == "" {
			s.runtime.ReportError(errors.New("missing wifi action"))
			return
		}
		err = exec.Command("nmcli", "connection", "up", "id", action).Run()
	}
	if err != nil {
		s.runtime.ReportError(err)
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
