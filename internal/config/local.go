package config

import "github.com/caarlos0/env/v6"

type OrchestratorService struct {
	Key     string `env:"AGENT_KEY"`
	ID      string `env:"AGENT_ID"`
	Address string `env:"ORCHESTRATOR_ADDRESS" envDefault:"https://api.k8sdeploy.dev"`
}

type Services struct {
	OrchestratorService
}

type Local struct {
	KeepLocal   bool `env:"LOCAL_ONLY" envDefault:"false" json:"keep_local,omitempty"`
	Development bool `env:"DEVELOPMENT" envDefault:"false" json:"development,omitempty"`
	HTTPPort    int  `env:"HTTP_PORT" envDefault:"3000" json:"port,omitempty"`
	Services    `json:"services"`
}

func BuildLocal(cfg *Config) error {
	local := &Local{}
	if err := env.Parse(local); err != nil {
		return err
	}
	cfg.Local = *local
	return nil
}
