package config

import (
	"uberlauncher/internal/meta"
	"uberlauncher/internal/skill"

	"github.com/BurntSushi/toml"
)

type Config struct {
	configMap MapConfigStruct
}

func New() (*Config, error) {
	path, err := meta.GetConfigPath()
	if err != nil {
		return nil, err
	}

	err = validateConfig(path)
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

func validateConfig(path string) error {
	var configStruct ConfigStruct

	_, err := toml.DecodeFile(path, &configStruct)
	if err != nil {
		return err
	}

	return nil
}

func getConfigMap(path string) (MapConfigStruct, error) {
	var configMap MapConfigStruct

	_, err := toml.DecodeFile(path, &configMap)
	if err != nil {
		return MapConfigStruct{}, err
	}

	return configMap, nil
}

func (c *Config) GetGeneralConfig() ConfigMap {
	return c.configMap.General
}

func (c *Config) GetForSkill(skill skill.Skill) skill.ConfigMap {
	return c.configMap.Skills[skill.Id()]
}
