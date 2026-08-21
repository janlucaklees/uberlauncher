package vpn

import (
	"math"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

func connectionEntry(ctx skill.Context, c *connection) entry.Entry {
	terms := []entry.Term{{Text: "vpn " + c.name, MinScore: math.MinInt}}
	if c.isConnected {
		terms = append(terms, entry.Term{Text: "connected", MinScore: 15})
	}
	return entry.Entry{
		Key:   c.name,
		Label: connectionLabel(c),
		Terms: terms,
		Run: func(ec entry.Context) {
			var err error
			if c.isConnected {
				err = ctx.Runtime.Command("nmcli", "connection", "down", "id", c.name).Run()
			} else {
				err = ctx.Runtime.Command("nmcli", "connection", "up", "id", c.name).Run()
			}
			if err != nil {
				ctx.Notifier.ReportError(err)
				ec.UI.KeepOpen()
			}
		},
	}
}

func (s *VpnSkill) upsertConnectionEntries(connections []connection) {
	for _, c := range connections {
		c := c
		s.ctx.Store.UpsertEntry(connectionEntry(s.ctx, &c))
	}
}

func connectionLabel(c *connection) string {
	label := "vpn " + c.name
	if c.isConnected {
		label += " 🔗"
	}
	return label
}
