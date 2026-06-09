package wifi

import (
	"errors"
	"time"

	"uberlauncher/internal/entry"
	"uberlauncher/internal/skill"
)

type WifiSkill struct {
	ctx skill.Context
}

func New() skill.Skill {
	return &WifiSkill{}
}

func (s *WifiSkill) Id() string { return "wifi" }

func (s *WifiSkill) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("nmcli") {
		ctx.Notifier.ReportError(errors.New("nmcli not found"))
		return
	}

	// Save the context
	s.ctx = ctx

	// Get the current connection and set the statuses
	ctx.Status.Register("wifi", 5*time.Second, func() string {
		ac := s.getActiveConnection()
		if ac == nil {
			return "Not connected"
		}
		return ac.ssid
	})
	ctx.Status.Register("wifi:icon", 10*time.Second, func() string {
		ac := s.getActiveConnection()
		return getConnectionSignalIcon(ac)
	})

	// Upsert static entries first
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

	// List all known connections
	ctx.Runtime.Go(func() {
		s.upsertKnownNetworks()
	})
}

func (s *WifiSkill) upsertKnownNetworks() {
	saved, err := s.loadSavedConnections()
	if err != nil {
		s.ctx.Notifier.ReportError(err)
		return
	}
	s.upsertNetworkEntries(saved)
}
