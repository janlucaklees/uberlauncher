package vpn

import (
	"errors"
	"time"

	"uberlauncher/internal/skill"
)

type VpnSkill struct {
	ctx skill.Context
}

func New() skill.Skill {
	return &VpnSkill{}
}

func (s *VpnSkill) Id() string { return "vpn" }

func (s *VpnSkill) Init(ctx skill.Context) {
	if !ctx.Runtime.HasCommand("nmcli") {
		ctx.Notifier.ReportError(errors.New("nmcli not found"))
		return
	}

	// Save the context
	s.ctx = ctx

	// Set the status icon
	ctx.Status.Register("vpn:icon", 5*time.Second, func() string {
		if s.hasActiveConnection() {
			return "󰌾"
		}
		return "󰌿"
	})

	// List all known VPN connections
	ctx.Runtime.Go(func() {
		s.upsertKnownConnections()
	})
}

func (s *VpnSkill) upsertKnownConnections() {
	connections, err := s.loadVPNConnections()
	if err != nil {
		s.ctx.Notifier.ReportError(err)
		return
	}
	s.upsertConnectionEntries(connections)
}
