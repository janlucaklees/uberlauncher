package wifi

import (
	"strconv"
	"strings"
)

type connection struct {
	ssid        string
	signal      int
	isSecured   bool
	isConnected bool
	isSaved     bool
}

func (s *WifiSkill) loadSavedConnections() ([]connection, error) {
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
		name, typ, active := parts[0], strings.ToLower(parts[1]), parts[2]
		if !strings.Contains(typ, "wireless") {
			continue
		}
		result = append(result, connection{ssid: name, isSaved: true, isConnected: active == "yes"})
	}
	return result, nil
}

func (s *WifiSkill) getActiveConnection() *connection {
	out, err := s.ctx.Runtime.Command("nmcli", "-t", "-f", "ACTIVE,SSID,SIGNAL,SECURITY", "dev", "wifi", "list", "--rescan", "no").Output()
	if err != nil {
		return nil
	}
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "yes:") {
			continue
		}
		rest := line[4:]

		lastColon := strings.LastIndex(rest, ":")
		if lastColon == -1 {
			continue
		}
		security := rest[lastColon+1:]
		rest = rest[:lastColon]

		prevColon := strings.LastIndex(rest, ":")
		if prevColon == -1 {
			continue
		}
		signalStr := rest[prevColon+1:]
		ssid := strings.ReplaceAll(rest[:prevColon], `\:`, ":")

		if ssid == "" {
			continue
		}
		signal, err := strconv.Atoi(signalStr)
		if err != nil {
			continue
		}
		return &connection{
			ssid:        ssid,
			signal:      signal,
			isSecured:   security != "" && security != "--",
			isConnected: true,
			isSaved:     true,
		}
	}
	return nil
}
