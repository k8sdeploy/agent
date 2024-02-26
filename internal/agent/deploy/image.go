package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

type ImageRequest struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	RequestDetails RequestDetails
	RequestID      string

	UpdateStatus bool
}

func NewImage(cs *kubernetes.Clientset, ctx context.Context) *ImageRequest {
	return &ImageRequest{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (i *ImageRequest) SetRequestID(rid string) {
	i.RequestID = rid
}

func validateImageRequest(details RequestDetails) error {
	if details.Kube.Name == "" {
		return logs.Error("name is required")
	}

	if details.Kube.Namespace == "" {
		return logs.Error("namespace is required")
	}

	if details.Image.ContainerURL == "" {
		return logs.Error("container_url is required")
	}

	if details.Image.Hash == "" && details.Image.Tag == "" {
		return logs.Error("hash or tag is required")
	}

	return nil
}

func (i *ImageRequest) ProcessRequest(details RequestDetails) error {
	if err := validateImageRequest(details); err != nil {
		return logs.Errorf("failed to validate request: %v", err)
	}

	i.RequestDetails = details

	deps := i.ClientSet.AppsV1().Deployments(details.Kube.Namespace)
	deployment, err := deps.Get(i.Context, details.Kube.Name, metav1.GetOptions{})
	if err != nil {
		return logs.Errorf("failed to get deployment: %v", err)
	}

	if details.Image.Tag != "" {
		deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", details.Image.ContainerURL, details.Image.Tag)
	}
	if details.Image.Hash != "" {
		deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", details.Image.ContainerURL, details.Image.Hash)
	}

	_, err = deps.Update(i.Context, deployment, metav1.UpdateOptions{})
	if err != nil {
		return logs.Errorf("failed to update deployment: %v", err)
	}

	i.UpdateStatus = true

	return nil
}

func (i *ImageRequest) GetResponse() (string, error) {
	type Resp struct {
		Updated    bool      `json:"updated"`
		UpdateTime time.Time `json:"update_time"`
		RequestID  string    `json:"request_id"`
	}

	resp, err := json.Marshal(Resp{
		Updated:    i.UpdateStatus,
		UpdateTime: time.Now(),
		RequestID:  i.RequestID,
	})

	if err != nil {
		return "", logs.Errorf("failed to marshal response: %v", err)
	}

	return string(resp), nil
}
