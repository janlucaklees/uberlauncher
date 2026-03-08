package wifi

import (
	"context"
	"errors"
	"os/exec"
	"strings"

	"uberlauncher/internal/skill"
	"uberlauncher/internal/types"
)

type Skill struct{}

func New() skill.Skill {
	return &Skill{}
}

func (s *Skill) Manifest() skill.Manifest {
	return skill.Manifest{Name: "wifi", SupportsFreeText: false}
}

func (s *Skill) Start(ctx context.Context, runtime skill.Runtime) error {
	entries := []types.EntryDTO{
		{SkillName: "wifi", EntryID: "wifi on", DisplayText: "wifi on", IsFreeText: false},
		{SkillName: "wifi", EntryID: "wifi off", DisplayText: "wifi off", IsFreeText: false},
		{SkillName: "wifi", EntryID: "wifi toggle", DisplayText: "wifi toggle", IsFreeText: false},
	}

	names, _ := knownConnections()
	for _, name := range names {
		entries = append(entries, types.EntryDTO{
			SkillName:   "wifi",
			EntryID:     "wifi " + name,
			DisplayText: "wifi " + name,
			IsFreeText:  false,
		})
	}

	runtime.PublishEntries(entries)
	return nil
}

func (s *Skill) Execute(ctx context.Context, cmd types.RunCommandDTO) error {
	if cmd.TriggerType != types.TriggerEntry {
		return errors.New("wifi skill only supports entry triggers")
	}
	if !hasCommand("nmcli") {
		return errors.New("nmcli not found")
	}

	action := strings.TrimSpace(cmd.EntryID)
	action = strings.TrimPrefix(action, "wifi ")

	switch action {
	case "on":
		return exec.CommandContext(ctx, "nmcli", "radio", "wifi", "on").Run()
	case "off":
		return exec.CommandContext(ctx, "nmcli", "radio", "wifi", "off").Run()
	case "toggle":
		return toggleWifi(ctx)
	default:
		if action == "" {
			return errors.New("missing wifi action")
		}
		return exec.CommandContext(ctx, "nmcli", "connection", "up", "id", action).Run()
	}
}

func (s *Skill) Stop(ctx context.Context) error {
	return nil
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

func toggleWifi(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "nmcli", "radio", "wifi")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	status := strings.TrimSpace(strings.ToLower(string(output)))
	if status == "enabled" {
		return exec.CommandContext(ctx, "nmcli", "radio", "wifi", "off").Run()
	}
	return exec.CommandContext(ctx, "nmcli", "radio", "wifi", "on").Run()
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
