package config

import "github.com/caarlos0/env/v6"

type K8sDeploy struct {
	APIAddress    string `env:"API_ADDRESS" envDefault:"https://api.k8sdeploy.dev/v1"`
	SocketAddress string `env:"SOCKET_ADDRESS" envDefault:"https://sockets.chewedfeed.com"`
	Key           string `env:"K8SDEPLOY_API_KEY" envDefault:""`
	Secret        string `env:"K8SDEPLOY_API_SECRET" envDefault:""`
	CompanyID     string `env:"K8SDEPLOY_COMPANY_ID" envDefault:""`
	SelfUpdate    bool   `env:"K8SDEPLOY_SELF_UPDATE" envDefault:"false"`
}

func BuildK8sDeploy(c *Config) error {
	cfg := &K8sDeploy{}

	if err := env.Parse(cfg); err != nil {
		return err
	}
	c.K8sDeploy = *cfg

	return nil
}
