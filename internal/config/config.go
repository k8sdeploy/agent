package config

import (
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Local
	K8sDeploy
}

func Build(buildVersion string) (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, bugLog.Error(err)
	}
	if err := BuildLocal(buildVersion, cfg); err != nil {
		return nil, bugLog.Error(err)
	}
	if err := BuildK8sDeploy(cfg); err != nil {
		return nil, bugLog.Error(err)
	}

	return cfg, nil
}
