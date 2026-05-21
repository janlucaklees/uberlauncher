package config

import (
	defaultconfig "uberlauncher/config"
	"uberlauncher/internal/meta"
	"uberlauncher/internal/skill"

	"github.com/BurntSushi/toml"
)

type Config struct {
	configMap ConfigStruct
}

func New() (*Config, error) {
	path, err := meta.GetConfigPath(defaultconfig.DefaultTOML)
	if err != nil {
		return nil, err
	}

	configMap, err := getConfigMap(path)
	if err != nil {
		return nil, err
	}

	return &Config{
		configMap: configMap,
	}, nil
}

func getConfigMap(path string) (ConfigStruct, error) {
	var configMap ConfigStruct

	_, err := toml.DecodeFile(path, &configMap)
	if err != nil {
		return ConfigStruct{}, err
	}

	return configMap, nil
}

func (c *Config) GetGeneralConfig() skill.ConfigMap {
	return c.configMap.General
}

func (c *Config) GetForSkill(skill skill.Skill) skill.ConfigMap {
	return c.configMap.Skills[skill.Id()]
}
