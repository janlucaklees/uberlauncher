package config

import "uberlauncher/internal/skill"

type Config struct{}

func New() *Config {
	return &Config{}
}

func (c *Config) Get(key string) string     { return "" }
func (c *Config) GetInt(key string) int     { return 0 }
func (c *Config) GetBool(key string) bool   { return false }
func (c *Config) Set(key string, value any) {}

func (c *Config) GetForSkill(_ skill.Skill) *Config { return c }
