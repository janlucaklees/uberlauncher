package config

import "uberlauncher/internal/skill"

type StatusBarConfig struct {
	Enabled bool   `toml:"enabled"`
	Left    string `toml:"left"`
	Center  string `toml:"center"`
	Right   string `toml:"right"`
}

type ConfigStruct struct {
	General   skill.ConfigMap    `toml:"general"`
	StatusBar StatusBarConfig    `toml:"statusbar"`
	Skills    SkillsConfigStruct `toml:"skills"`
}

type SkillsConfigStruct map[string]skill.ConfigMap
