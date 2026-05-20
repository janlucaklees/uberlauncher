package config

import "uberlauncher/internal/skill"

type ConfigMap map[string]any

type MapConfigStruct struct {
	General ConfigMap             `toml:"general"`
	Skills  MapSkillsConfigStruct `toml:"skills"`
}

type MapSkillsConfigStruct map[string]skill.ConfigMap

type ConfigStruct struct {
	General GeneralConfigStruct `toml:"general"`
	Skills  SkillsConfigStruct  `toml:"skills"`
}

type GeneralConfigStruct struct{}

type SkillsConfigStruct struct {
	Todo TodoConfigStruct `toml:"todo"`
}

type TodoConfigStruct struct {
	Token string `toml:"token"`
}
