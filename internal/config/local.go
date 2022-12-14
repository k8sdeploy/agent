package config

import "github.com/caarlos0/env/v6"

type Local struct {
	KeepLocal    bool `env:"LOCAL_ONLY" envDefault:"false" json:"keep_local,omitempty"`
	Development  bool `env:"DEVELOPMENT" envDefault:"false" json:"development,omitempty"`
	HTTPPort     int  `env:"HTTP_PORT" envDefault:"3000" json:"port,omitempty"`
	SelfUpdate   bool `env:"SELF_UPDATE" envDefault:"true" json:"self_update,omitempty"`
	BuildVersion string
}

func BuildLocal(buildVersion string, cfg *Config) error {
	local := &Local{}
	if err := env.Parse(local); err != nil {
		return err
	}
	cfg.BuildVersion = buildVersion
	cfg.Local = *local
	return nil
}
