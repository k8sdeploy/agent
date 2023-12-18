package config

import (
	"github.com/caarlos0/env/v6"
)

type Credentials struct {
	Agent struct {
		Key       string `env:"K8SDEPLOY_KEY" envDefault:""`
		Secret    string `env:"K8SDEPLOY_SECRET" envDefault:""`
		CompanyID string `env:"K8SDEPLOY_COMPANY_ID" envDefault:""`
	}
	Queue struct {
		Key    string `env:"K8SDEPLOY_QUEUE_KEY" envDefault:""`
		Secret string `env:"K8SDEPLOY_QUEUE_SECRET" envDefault:""`
	}
}

//type AgentCredentials struct {
//	Key       string `env:"K8SDEPLOY_KEY" envDefault:""`
//	Secret    string `env:"K8SDEPLOY_SECRET" envDefault:""`
//	CompanyID string `env:"K8SDEPLOY_COMPANY_ID" envDefault:""`
//}
//
//type QueueCredentials struct {
//	Key    string `env:"K8SDEPLOY_QUEUE_KEY" envDefault:""`
//	Secret string `env:"K8SDEPLOY_QUEUE_SECRET" envDefault:""`
//}

type Queues struct {
	Master   string `env:"K8SDEPLOY_MASTER_QUEUE" envDefault:""`
	Agent    string `env:"K8SDEPLOY_AGENT_QUEUE" envDefault:""`
	Response string `env:"K8SDEPLOY_RESPONSE_QUEUE" envDefault:""`
}

type K8sDeploy struct {
	APIAddress string `env:"API_ADDRESS" envDefault:"https://api.k8sdeploy.dev/v1"`

	SelfUpdate   bool   `env:"K8SDEPLOY_SELF_UPDATE" envDefault:"false"`
	BuildVersion string `env:"BUILD_VERSION" envDefault:""`

	Queues
	Credentials
}

func BuildK8sDeploy(c *Config) error {
	cfg := &K8sDeploy{}

	if err := env.Parse(cfg); err != nil {
		return err
	}
	c.K8sDeploy = *cfg

	return nil
}
