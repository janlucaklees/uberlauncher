package wifi

import (
	"errors"
	"os/exec"
	"strings"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type WifiSkill struct{}

func New() skill.Skill {
	return &WifiSkill{}
}

func (s *WifiSkill) Id() string { return "wifi" }

func (s *WifiSkill) Init(ctx skill.Context) {
	enabled, ok := ctx.Config["enabled"].(bool)
	if ok && !enabled {
		return
	}

	if !ctx.Runtime.HasCommand("nmcli") {
		ctx.Notifier.ReportError(errors.New("nmcli not found"))
		return
	}

	ctx.Store.UpsertEntry(entry.Entry{
		Label: "wifi on",
		Run: func(ec entry.Context) {
			if err := exec.Command("nmcli", "radio", "wifi", "on").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "wifi off",
		Run: func(ec entry.Context) {
			if err := exec.Command("nmcli", "radio", "wifi", "off").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})

	names, err := knownConnections()
	if err != nil {
		ctx.Notifier.ReportError(err)
		return
	}
	for _, name := range names {
		name := name
		ctx.Store.UpsertEntry(entry.Entry{
			Label: "wifi " + name,
			Run: func(ec entry.Context) {
				if err := exec.Command("nmcli", "connection", "up", "id", name).Run(); err != nil {
					ctx.Notifier.ReportError(err)
					ec.UI.KeepOpen()
				}
			},
		})
	}
}

func knownConnections() ([]string, error) {
	out, err := exec.Command("nmcli", "-t", "-f", "NAME,TYPE", "connection", "show").Output()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.Contains(strings.ToLower(parts[1]), "wifi") || strings.Contains(strings.ToLower(parts[1]), "wireless") {
			names = append(names, parts[0])
		}
	}
	return names, nil
}
