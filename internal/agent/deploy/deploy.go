package deploy

import (
	"context"
	"encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"

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
		return nil, logs.Errorf("unmarshal deployment info: %v", err)
	}

	if msgMap["name"] == nil {
		return nil, logs.Error("deployment name is required")
	}

	return &DeploymentInfo{
		Name:      msgMap["name"].(string),
		Namespace: msgMap["namespace"].(string),
		ImageURL:  msgMap["image_url"].(string),
	}, nil
}

func (d *Deployment) DeployImage(deploymentInfo string) error {
	//di, err := parseDeploymentInfo(deploymentInfo)
	//if err != nil {
	//	return logs.Errorf("parse deployment info: %v", err)
	//}
	//
	//deps := d.ClientSet.AppsV1().Deployments(di.Namespace)
	//list, err := deps.List(d.Context, metav1.ListOptions{})
	//if err != nil {
	//	return logs.Errorf("list deployments: %v", err)
	//}
	//
	//for _, dep := range list.Items {
	//	if dep.ObjectMeta.Name == di.Name {
	//		dep.Spec.Template.Spec.Containers[0].Image = di.ImageURL
	//		//nolint:gosec
	//		_, err := deps.Update(d.Context, &dep, metav1.UpdateOptions{})
	//		if err != nil {
	//			return logs.Errorf("update deployment: %v", err)
	//		}
	//	}
	//}
	return nil
}

func (d *Deployment) DeleteDeployment(deploymentInfo string) error {
	//di, err := parseDeploymentInfo(deploymentInfo)
	//if err != nil {
	//  return logs.Errorf("parse deployment info: %v", err)
	//}
	//
	//deps := d.ClientSet.AppsV1().Deployments(di.Namespace)
	//err = deps.Delete(d.Context, di.Name, metav1.DeleteOptions{})
	//if err != nil {
	//  return logs.Errorf("delete deployment: %v", err)
	//}

	return nil
}

func (d *Deployment) GetDeploymentStatus(deploymentInfo string) error {
	return nil
}
