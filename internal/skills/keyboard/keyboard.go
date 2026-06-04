package keyboard

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type KeyboardSkill struct{}

func New() skill.Skill {
	return &KeyboardSkill{}
}

func (s *KeyboardSkill) Id() string { return "keyboard" }

func (s *KeyboardSkill) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("hyprctl") {
		ctx.Notifier.ReportError(errors.New("hyprctl not found"))
		return
	}

	layouts, err := configuredLayouts(ctx.Runtime)
	if err != nil {
		ctx.Notifier.ReportError(err)
		return
	}

	for i, layout := range layouts {
		i, layout := i, layout
		ctx.Store.UpsertEntry(entry.Entry{
			Label: "keyboard " + layout,
			Run: func(ec entry.Context) {
				if err := exec.Command("hyprctl", "switchxkblayout", "all", fmt.Sprintf("%d", i)).Run(); err != nil {
					ctx.Notifier.ReportError(err)
					ec.UI.KeepOpen()
				}
			},
		})
	}
}

func configuredLayouts(rt skill.Runtime) ([]string, error) {
	out, err := rt.Command("hyprctl", "getoption", "input:kb_layout").Output()
	if err != nil {
		return nil, err
	}

	// Output format: "str: us,de,fr\nint: 0\n..."
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "str: ") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, "str: "))
		if value == "" {
			return nil, errors.New("no keyboard layouts configured in Hyprland")
		}
		return strings.Split(value, ","), nil
	}
	return nil, errors.New("could not parse hyprctl getoption output")
}
