package bluetooth

import (
	"errors"
	"os/exec"
	"strings"
	"time"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type BluetoothSkill struct{}

func New() skill.Skill {
	return &BluetoothSkill{}
}

func (s *BluetoothSkill) Id() string { return "bluetooth" }

func (s *BluetoothSkill) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("bluetoothctl") {
		ctx.Notifier.ReportError(errors.New("bluetoothctl not found"))
		return
	}

	ctx.Store.UpsertEntry(entry.Entry{
		Label: "bluetooth on",
		Run: func(ec entry.Context) {
			if err := exec.Command("bluetoothctl", "power", "on").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "bluetooth off",
		Run: func(ec entry.Context) {
			if err := exec.Command("bluetoothctl", "power", "off").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})

	ctx.Status.Register("bluetooth:icon", 10*time.Second, func() string {
		return bluetoothIcon(ctx.Runtime)
	})

	devices, err := pairedDevices(ctx.Runtime)
	if err != nil {
		ctx.Notifier.ReportError(err)
		return
	}
	for _, d := range devices {
		d := d
		ctx.Store.UpsertEntry(entry.Entry{
			Label: "bluetooth " + d.name,
			Run: func(ec entry.Context) {
				if err := exec.Command("bluetoothctl", "connect", d.mac).Run(); err != nil {
					ctx.Notifier.ReportError(err)
					ec.UI.KeepOpen()
				}
			},
		})
	}
}

func bluetoothIcon(rt skill.Runtime) string {
	out, err := rt.Command("bluetoothctl", "show").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Powered: yes") {
			return "󰂯"
		}
	}
	return "󰂲"
}

type device struct {
	mac  string
	name string
}

func pairedDevices(rt skill.Runtime) ([]device, error) {
	out, err := rt.Command("bluetoothctl", "devices", "Paired").Output()
	if err != nil {
		return nil, err
	}

	var devices []device
	for _, line := range strings.Split(string(out), "\n") {
		// Format: "Device <MAC> <Name>"
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 || parts[0] != "Device" {
			continue
		}
		devices = append(devices, device{mac: parts[1], name: parts[2]})
	}
	return devices, nil
}
