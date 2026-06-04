package wifi

import (
	"bufio"
	"errors"
	"math"
	"strings"
	"time"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type WifiSkill struct{}

func New() skill.Skill {
	return &WifiSkill{}
}

func (s *WifiSkill) Id() string { return "wifi" }

func (s *WifiSkill) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("nmcli") {
		ctx.Notifier.ReportError(errors.New("nmcli not found"))
		return
	}

	ctx.Status.Register("wifi", 5*time.Second, func() string {
		return activeSSID(ctx)
	})
	ctx.Status.Register("wifi:icon", 10*time.Second, func() string {
		return wifiIcon(ctx)
	})

	ctx.Store.UpsertEntry(entry.Entry{
		Label: "wifi on",
		Run: func(ec entry.Context) {
			if err := ctx.Runtime.Command("nmcli", "radio", "wifi", "on").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})
	ctx.Store.UpsertEntry(entry.Entry{
		Label: "wifi off",
		Run: func(ec entry.Context) {
			if err := ctx.Runtime.Command("nmcli", "radio", "wifi", "off").Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	})

	names, active, err := wifiConnections(ctx)
	if err != nil {
		ctx.Notifier.ReportError(err)
		return
	}
	for _, name := range names {
		ctx.Store.UpsertEntry(networkEntry(ctx, name, name == active))
	}

	ctx.Runtime.Go(func() { watchConnections(ctx, names) })
}

func watchConnections(ctx skill.Context, names []string) {
	known := make(map[string]bool, len(names))
	for _, n := range names {
		known[n] = true
	}

	wifiDevices := wifiDeviceNames(ctx)
	isWifiDevice := func(line string) bool {
		for _, d := range wifiDevices {
			if strings.HasPrefix(line, d+": ") {
				return true
			}
		}
		return len(wifiDevices) == 0
	}

	cmd := ctx.Runtime.Command("nmcli", "monitor")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ctx.Notifier.ReportError(err)
		return
	}
	if err := cmd.Start(); err != nil {
		ctx.Notifier.ReportError(err)
		return
	}

	active := activeConnectionName(ctx)
	var pending string

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if !isWifiDevice(line) {
			continue
		}

		if idx := strings.Index(line, ": using connection '"); idx != -1 {
			rest := line[idx+len(": using connection '"):]
			if end := strings.Index(rest, "'"); end != -1 {
				pending = rest[:end]
			}
		} else if strings.HasSuffix(line, ": connected") {
			if pending != "" && pending != active && known[pending] {
				if active != "" {
					ctx.Store.UpsertEntry(networkEntry(ctx, active, false))
				}
				active = pending
				ctx.Store.UpsertEntry(networkEntry(ctx, active, true))
			}
			pending = ""
		} else if strings.HasSuffix(line, ": disconnected") {
			if active != "" {
				ctx.Store.UpsertEntry(networkEntry(ctx, active, false))
				active = ""
			}
			pending = ""
		}
	}

	cmd.Wait()
}

func networkEntry(ctx skill.Context, name string, connected bool) entry.Entry {
	terms := []entry.Term{
		{Text: "wifi " + name, MinScore: math.MinInt},
	}
	if connected {
		terms = append(terms, entry.Term{Text: "connected", MinScore: 15})
	}
	return entry.Entry{
		Key:   "wifi " + name,
		Label: "wifi " + name,
		Terms: terms,
		Run: func(ec entry.Context) {
			if err := ctx.Runtime.Command("nmcli", "connection", "up", "id", name).Run(); err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	}
}

func wifiIcon(ctx skill.Context) string {
	if activeSSID(ctx) != "" {
		return "󰤨"
	}
	return "󰤭"
}

func activeSSID(ctx skill.Context) string {
	out, err := ctx.Runtime.Command("nmcli", "-t", "-f", "ACTIVE,SSID", "dev", "wifi").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && parts[0] == "yes" {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

// activeConnectionName returns the profile name of the active Wi-Fi connection,
// which matches the names returned by wifiConnections().
func activeConnectionName(ctx skill.Context) string {
	out, err := ctx.Runtime.Command("nmcli", "-t", "-f", "NAME,TYPE", "connection", "show", "--active").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		typ := strings.ToLower(parts[1])
		if strings.Contains(typ, "wifi") || strings.Contains(typ, "wireless") {
			return parts[0]
		}
	}
	return ""
}

func wifiDeviceNames(ctx skill.Context) []string {
	out, err := ctx.Runtime.Command("nmcli", "-t", "-f", "DEVICE,TYPE", "device", "status").Output()
	if err != nil {
		return nil
	}
	var devices []string
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			typ := strings.ToLower(parts[1])
			if strings.Contains(typ, "wifi") || strings.Contains(typ, "wireless") {
				devices = append(devices, parts[0])
			}
		}
	}
	return devices
}

// wifiConnections returns all known wifi connection profile names and the currently active one
// in a single nmcli call, replacing the previous two separate queries.
func wifiConnections(ctx skill.Context) (names []string, active string, err error) {
	out, err := ctx.Runtime.Command("nmcli", "-t", "-f", "NAME,TYPE,ACTIVE", "connection", "show").Output()
	if err != nil {
		return nil, "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}
		name, typ, isActive := parts[0], strings.ToLower(parts[1]), parts[2]
		if strings.Contains(typ, "wifi") || strings.Contains(typ, "wireless") {
			names = append(names, name)
			if isActive == "yes" {
				active = name
			}
		}
	}
	return names, active, nil
}

