package deploy

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Deployment struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context
}

type DeploymentInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	ImageURL  string `json:"image_url"`
}

func NewDeployment(clientset *kubernetes.Clientset, ctx context.Context) *Deployment {
	return &Deployment{
		ClientSet: clientset,
		Context:   ctx,
	}
}

func parseDeploymentInfo(deploymentInfo string) (*DeploymentInfo, error) {
	var msgMap map[string]interface{}
	if err := json.Unmarshal([]byte(deploymentInfo), &msgMap); err != nil {
		return nil, err
	}

	if msgMap["name"] == nil {
		return nil, fmt.Errorf("deployment name is required")
	}

	return &DeploymentInfo{
		Name:      msgMap["name"].(string),
		Namespace: msgMap["namespace"].(string),
		ImageURL:  msgMap["image_url"].(string),
	}, nil
}

func (d *Deployment) DeployImage(deploymentInfo string) error {
	di, err := parseDeploymentInfo(deploymentInfo)
	if err != nil {
		return err
	}

	deps := d.ClientSet.AppsV1().Deployments(di.Namespace)
	list, err := deps.List(d.Context, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, dep := range list.Items {
		if dep.ObjectMeta.Name == di.Name {
			dep.Spec.Template.Spec.Containers[0].Image = di.ImageURL
			//nolint:gosec
			_, err := deps.Update(d.Context, &dep, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
