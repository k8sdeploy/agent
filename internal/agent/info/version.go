package info

import (
	"context"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

const (
	versionRequestType = "version"
)

type VersionRequest struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	ClientSet *kubernetes.Clientset
	Context   context.Context
}

type VersionSendResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Image     string `json:"image"`
	Version   string `json:"version"`
}

func NewVersion(cs *kubernetes.Clientset, ctx context.Context) *VersionRequest {
	return &VersionRequest{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (v *VersionRequest) ParseRequest(msgMap map[string]interface{}) error {
	namespace := msgMap["namespace"]
	if namespace == nil {
		return logs.Error("namespace is required")
	}

	name := msgMap["name"]
	if name == nil {
		return fmt.Errorf("name is required")
	}
	v.Namespace = namespace.(string)
	v.Name = name.(string)

	return nil
}

func (v *VersionRequest) SendResponse() error {
	vi, err := v.getVersion()
	if err != nil {
		return logs.Errorf("failed to get version: %v", err)
	}
	fmt.Printf("version: %+v\n", vi)

	return nil
}

func (v *VersionRequest) getVersion() (*VersionSendResponse, error) {
	dep, err := v.ClientSet.AppsV1().Deployments(v.Namespace).Get(v.Context, v.Name, metav1.GetOptions{})
	if err != nil {
		return nil, logs.Errorf("failed to get deployment: %v", err)
	}

	i := dep.Spec.Template.Spec.Containers[0].Image
	version := strings.Split(i, ":")[1]

	return &VersionSendResponse{
		Name:      v.Name,
		Namespace: v.Namespace,
		Version:   version,
		Image:     i,
	}, nil
}
