package agent

import (
	"context"
	"fmt"

	"github.com/k8sdeploy/agent/internal/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Agent struct {
	Config *config.Config
}

func NewAgent(cfg *config.Config) *Agent {
	return &Agent{
		Config: cfg,
	}
}

func (a *Agent) Start() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	for {
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return err
		}
		fmt.Printf("Pods in cluster: %d\n", len(pods.Items))
	}
}
