package config

import (
	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
)

type Config struct {
	K8sDeploy
	ConfigBuilder.Config
}

func Build(buildVersion string) (*Config, error) {
	cfg := &Config{}

	if err := BuildK8sDeploy(cfg); err != nil {
		return nil, logs.Errorf("k8sdeploy: %v", err)
	}

	c, err := ConfigBuilder.Build(ConfigBuilder.Local)
	if err != nil {
		return nil, logs.Errorf("configBuilder: %v", err)
	}
	cfg.Config = *c

	return cfg, nil
}
