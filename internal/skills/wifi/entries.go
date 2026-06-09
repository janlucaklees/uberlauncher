package wifi

import (
	"math"
	"strings"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

func networkEntry(ctx skill.Context, c *connection) entry.Entry {
	terms := []entry.Term{{Text: "wifi " + c.ssid, MinScore: math.MinInt}}
	if c.isConnected {
		terms = append(terms, entry.Term{Text: "connected", MinScore: 15})
	}
	return entry.Entry{
		Key:   c.ssid,
		Label: networkLabel(c),
		Terms: terms,
		Run: func(ec entry.Context) {
			var err error
			if c.isConnected {
				err = ctx.Runtime.Command("nmcli", "connection", "down", "id", c.ssid).Run()
			} else {
				err = ctx.Runtime.Command("nmcli", "connection", "up", "id", c.ssid).Run()
			}
			if err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	}
}

func (s *WifiSkill) upsertNetworkEntries(connections []connection) {
	for _, c := range connections {
		c := c
		s.ctx.Store.UpsertEntry(networkEntry(s.ctx, &c))
	}
}

func networkLabel(c *connection) string {
	parts := []string{"wifi " + c.ssid}
	if c.isConnected {
		parts = append(parts, getConnectionConnectedIcon(c))
	} else {
		parts = append(parts, getConnectionSavedIcon(c))
	}
	parts = append(parts, getConnectionSignalIcon(c))
	if icon := getConnectionSecuredIcon(c); icon != "" {
		parts = append(parts, icon)
	}
	return strings.Join(parts, " ")
}

func getConnectionConnectedIcon(c *connection) string {
	if c != nil && c.isConnected {
		return "🔗"
	}
	return ""
}

func getConnectionSavedIcon(c *connection) string {
	if c != nil && c.isSaved {
		return "󰆓"
	}
	return ""
}

func getConnectionSecuredIcon(c *connection) string {
	if c != nil && c.isSecured {
		return "󰒃"
	}
	return ""
}

func getConnectionSignalIcon(c *connection) string {
	if c == nil || c.signal == 0 {
		return "󰤭"
	}
	switch {
	case c.signal >= 75:
		return "󰤨"
	case c.signal >= 50:
		return "󰤥"
	case c.signal >= 25:
		return "󰤢"
	default:
		return "󰤟"
	}
}
