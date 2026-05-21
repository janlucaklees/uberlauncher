package config

import "uberlauncher/internal/skill"

type ConfigStruct struct {
	General skill.ConfigMap    `toml:"general"`
	Skills  SkillsConfigStruct `toml:"skills"`
}

type SkillsConfigStruct map[string]skill.ConfigMap
