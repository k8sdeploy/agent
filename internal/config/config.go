package config

import (
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/caarlos0/env/v6"
	ConfigBuilder "github.com/keloran/go-config"
)

type Config struct {
	K8sDeploy
	ConfigBuilder.Config
}

func Build(buildVersion string) (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, logs.Error(err)
	}

	if err := BuildK8sDeploy(cfg); err != nil {
		return nil, logs.Error(err)
	}

	c, err := ConfigBuilder.Build(ConfigBuilder.Local)
	if err != nil {
		return nil, logs.Error(err)
	}
	cfg.Config = *c

	return cfg, nil
}
