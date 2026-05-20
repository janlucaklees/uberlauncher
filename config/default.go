package config

import _ "embed"

// DefaultTOML is the default config created on first launch.
//
//go:embed default.toml
var DefaultTOML []byte
