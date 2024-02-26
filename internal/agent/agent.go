package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	"net/http"
	"path/filepath"
	"time"

	"github.com/k8sdeploy/agent/internal/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type KubernetesClient struct {
	Context   context.Context
	ClientSet *kubernetes.Clientset
}

type Agent struct {
	Config           *config.Config
	KubernetesClient *KubernetesClient
}

type EventClient struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Token        string `json:"token"`
	EventChannel string `json:"channel"`
}

func NewAgent(cfg *config.Config) *Agent {
	return &Agent{
		Config: cfg,
	}
}

func (a *Agent) Start() error {
	errChan := make(chan error)
	billingTime := 5

	if err := a.connectOrchestrator(); err != nil {
		return logs.Errorf("failed to connect to orchestrator: %v", err)
	}
	if err := a.GetKubernetesClient(); err != nil {
		return logs.Errorf("failed to get kubernetes client: %v", err)
	}
	for {
		select {
		case err := <-errChan:
			if err != nil {
				logs.Infof("error in agent loop: %v", err)
				continue
			}
		case <-time.After(time.Duration(billingTime) * time.Second):
			go a.listenForEvents(errChan)

			if a.Config.SelfUpdate {
				go a.listenForSelfUpdate(errChan)
			}
			continue
		}
	}
}

func (a *Agent) connectOrchestrator() error {
	type AgentBody struct {
		Key       string `json:"key"`
		Secret    string `json:"secret"`
		CompanyID string `json:"company_id"`
	}
	b, err := json.Marshal(&AgentBody{
		Key:       a.Config.K8sDeploy.Credentials.Agent.Key,
		Secret:    a.Config.K8sDeploy.Credentials.Agent.Secret,
		CompanyID: a.Config.K8sDeploy.Credentials.Agent.CompanyID,
	})
	if err != nil {
		return logs.Errorf("failed to marshal agent body: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/agent", a.Config.K8sDeploy.APIAddress), bytes.NewBuffer(b))
	if err != nil {
		return logs.Errorf("failed to create request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return logs.Errorf("failed to connect to orchestrator: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return logs.Error("failed to connect to orchestrator")
	}

	type QueueName string
	const (
		AgentQueue    QueueName = "agent"
		ResponseQueue QueueName = "response"
		MasterQueue   QueueName = "master"
	)

	type Queue struct {
		Name QueueName `json:"name"`
		Path string    `json:"path"`
	}

	type Credentials struct {
		Key    string `json:"key"`
		Secret string `json:"secret"`
	}

	type AgentDetails struct {
		Credentials Credentials `json:"credentials"`
		Queues      []Queue     `json:"queues"`
	}

	var agentDetails AgentDetails
	if err := json.NewDecoder(res.Body).Decode(&agentDetails); err != nil {
		return logs.Errorf("failed to decode agent details: %v", err)
	}

	a.Config.K8sDeploy.Credentials.Queue.Key = agentDetails.Credentials.Key
	a.Config.K8sDeploy.Credentials.Queue.Secret = agentDetails.Credentials.Secret

	for _, queue := range agentDetails.Queues {
		switch queue.Name {
		case AgentQueue:
			a.Config.K8sDeploy.Queues.Agent = queue.Path
		case ResponseQueue:
			a.Config.K8sDeploy.Queues.Response = queue.Path
		case MasterQueue:
			a.Config.K8sDeploy.Queues.Master = queue.Path
		}
	}

	return nil
}

func (a *Agent) GetKubernetesClient() error {
	// get kubernetes config
	if a.Config.Development {
		cfgPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		cfg, err := clientcmd.BuildConfigFromFlags("", cfgPath)
		if err != nil {
			return logs.Errorf("failed to build config from flags: %v", err)
		}

		clientSet, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return logs.Errorf("failed to create clientset: %v", err)
		}
		a.KubernetesClient = &KubernetesClient{
			Context:   context.Background(),
			ClientSet: clientSet,
		}
		return nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return logs.Errorf("failed to get in cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return logs.Errorf("failed to create clientset: %v", err)
	}
	a.KubernetesClient = &KubernetesClient{
		Context:   context.Background(),
		ClientSet: clientset,
	}
	return nil
}
