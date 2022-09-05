package info

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

const (
	deploymentsRequestType = "deployments"
)

type DeploymentsRequest struct {
	Namespace string `json:"namespace"`
	Clientset *kubernetes.Clientset
	Context   context.Context
}

type DeploymentInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type DeploymentsSendResponse struct {
	Namespace   string           `json:"namespace"`
	Deployments []DeploymentInfo `json:"deployments"`
}

func NewDeployments() *DeploymentsRequest {
	return &DeploymentsRequest{}
}

func (n *DeploymentsRequest) ParseRequest(msgMap map[string]interface{}) error {
	namespace := msgMap["namespace"]
	if namespace == nil {
		return fmt.Errorf("namespace is required")
	}
	n.Namespace = namespace.(string)
	return nil
}

func (n *DeploymentsRequest) SendResponse() error {
	return nil
}
