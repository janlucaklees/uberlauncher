package vpn

import "strings"

type connection struct {
	name        string
	isConnected bool
}

func (s *VpnSkill) loadVPNConnections() ([]connection, error) {
	out, err := s.ctx.Runtime.Command("nmcli", "-t", "-f", "NAME,TYPE,ACTIVE", "connection", "show").Output()
	if err != nil {
		return nil, err
	}
	var result []connection
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}
		name, typ, active := parts[0], parts[1], parts[2]
		if typ != "vpn" {
			continue
		}
		result = append(result, connection{name: name, isConnected: active == "yes"})
	}
	return result, nil
}

func (s *VpnSkill) hasActiveConnection() bool {
	connections, err := s.loadVPNConnections()
	if err != nil {
		return false
	}
	for _, c := range connections {
		if c.isConnected {
			return true
		}
	}
	return false
}
