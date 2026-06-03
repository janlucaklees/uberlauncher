package battery

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"uberlauncher/internal/skill"
)

type BatterySkill struct{}

func New() skill.Skill {
	return &BatterySkill{}
}

func (s *BatterySkill) Id() string { return "battery" }

func (s *BatterySkill) Init(ctx skill.Context) {
	enabled, ok := ctx.Config["enabled"].(bool)
	if ok && !enabled {
		return
	}

	ctx.Status.Register("battery:icon", 60*time.Second, batteryIcon)
	ctx.Status.Register("battery:percentage", 60*time.Second, batteryPercentage)
}

func batteryBase() string {
	paths, _ := filepath.Glob("/sys/class/power_supply/BAT*/capacity")
	if len(paths) == 0 {
		return ""
	}
	return filepath.Dir(paths[0])
}

func batteryIcon() string {
	base := batteryBase()
	if base == "" {
		return ""
	}

	capacityData, err := os.ReadFile(filepath.Join(base, "capacity"))
	if err != nil {
		return ""
	}
	capacity, err := strconv.Atoi(strings.TrimSpace(string(capacityData)))
	if err != nil {
		return ""
	}

	statusData, _ := os.ReadFile(filepath.Join(base, "status"))
	if strings.TrimSpace(string(statusData)) == "Charging" || strings.TrimSpace(string(statusData)) == "Full" {
		return "󰂇"
	}

	icons := []string{"󰁺", "󰁻", "󰁼", "󰁽", "󰁾", "󰁿", "󰂀", "󰂁", "󰂂", "󰁹"}
	return icons[min(capacity/10, len(icons)-1)]
}

func batteryPercentage() string {
	base := batteryBase()
	if base == "" {
		return ""
	}

	data, err := os.ReadFile(filepath.Join(base, "capacity"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
