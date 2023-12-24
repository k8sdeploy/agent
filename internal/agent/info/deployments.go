package info

import (
	"context"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type DeploymentsRequest struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	RequestID string
	Response  *DeploymentsSendResponse
}

type DeploymentInfo struct {
	Name      string `json:"name"`
	Container string `json:"container"`
}
type DeploymentsSendResponse struct {
	Namespace   string           `json:"namespace"`
	Deployments []DeploymentInfo `json:"deployments"`
}

func NewDeployments(cs *kubernetes.Clientset, ctx context.Context, rid string) *DeploymentsRequest {
	return &DeploymentsRequest{
		ClientSet: cs,
		Context:   ctx,

		RequestID: rid,
	}
}

func (d *DeploymentsRequest) ProcessRequest(details RequestDetails) error {
	dp, err := d.GetDeployments(details.Namespace)
	if err != nil {
		return logs.Errorf("failed to get deployments: %v", err)
	}

	d.Response = &DeploymentsSendResponse{
		Namespace:   details.Namespace,
		Deployments: dp,
	}

	return nil
}

func (d *DeploymentsRequest) SendResponse() error {
	fmt.Printf("deployments: %+v\n", d.Response)

	return nil
}

func (d *DeploymentsRequest) GetDeployments(namespace string) ([]DeploymentInfo, error) {
	dep, err := d.ClientSet.AppsV1().Deployments(namespace).List(d.Context, metav1.ListOptions{})
	if err != nil {
		return nil, logs.Errorf("failed to get deployments: %v", err)
	}

	var deployments []DeploymentInfo
	for _, dd := range dep.Items {
		deployments = append(deployments, DeploymentInfo{
			Name:      dd.Name,
			Container: dd.Spec.Template.Spec.Containers[0].Image,
		})
	}

	return deployments, nil
}
