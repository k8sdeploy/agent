package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	EventClient      *EventClient
	SelfUpdate       *EventClient
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

	if err := a.connectOrchestrator(); err != nil {
		return err
	}
	if err := a.getKubernetesClient(); err != nil {
		return err
	}
	for {
		// todo dictate this number by billing
		time.Sleep(10 * time.Second)

		go a.listenForEvents(errChan)

		if a.Config.SelfUpdate {
			go a.listenForSelfUpdate(errChan)
		}

		if err := <-errChan; err != nil {
			return err
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
		Key:       a.Config.K8sDeploy.Credentials.Key,
		Secret:    a.Config.K8sDeploy.Credentials.Secret,
		CompanyID: a.Config.K8sDeploy.Credentials.CompanyID,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/agent", a.Config.K8sDeploy.APIAddress), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("failed to connect to orchestrator")
	}

	type AgentChannelDetails struct {
		Token   string `json:"token"`
		Channel string `json:"channel"`
	}

	type orchestratorResponse struct {
		Update AgentChannelDetails `json:"update"`
		Event  AgentChannelDetails `json:"event"`
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			panic(err)
		}
	}()
	var resp orchestratorResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return err
	}
	a.EventClient = &EventClient{
		ID:    resp.Event.Channel,
		Token: resp.Event.Token,
		Name:  "eventChannel",
	}
	a.SelfUpdate = &EventClient{
		ID:    resp.Update.Channel,
		Token: resp.Update.Token,
		Name:  "updateChannel",
	}

	return nil
}

func (a *Agent) getKubernetesClient() error {
	// get kubernetes config
	if a.Config.Development {
		cfgPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		cfg, err := clientcmd.BuildConfigFromFlags("", cfgPath)
		if err != nil {
			return err
		}

		clientSet, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return err
		}
		a.KubernetesClient = &KubernetesClient{
			Context:   context.Background(),
			ClientSet: clientSet,
		}
		return nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	a.KubernetesClient = &KubernetesClient{
		Context:   context.Background(),
		ClientSet: clientset,
	}
	return nil
}
