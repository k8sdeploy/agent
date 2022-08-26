package agent

import (
	"context"
	"fmt"
	"github.com/k8sdeploy/agent/internal/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"
)

type KubernetesClient struct {
	Context   context.Context
	ClientSet *kubernetes.Clientset
}

type Agent struct {
	Config           *config.Config
	EventClient      *EventClient
	KubernetesClient *KubernetesClient
}

type EventClient struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Token        string `json:"token"`
	EventChannel string
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
		if err := <-errChan; err != nil {
			return err
		}
	}
}

func (a *Agent) connectOrchestrator() error {
	// TODO: remove change this
	a.EventClient = &EventClient{
		Token:        "CH6JBQS-QWwgIH4",
		EventChannel: fmt.Sprintf("%s/application/1/message", a.Config.K8sDeploy.SocketAddress),
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
