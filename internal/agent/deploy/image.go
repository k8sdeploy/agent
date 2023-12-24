package deploy

import (
	"context"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ImageRequest struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	RequestDetails RequestDetails
	RequestID      string

	UpdateStatus bool
}

func NewImage(cs *kubernetes.Clientset, ctx context.Context, rid string) *ImageRequest {
	return &ImageRequest{
		ClientSet: cs,
		Context:   ctx,
		RequestID: rid,
	}
}

func (i *ImageRequest) ProcessRequest(details RequestDetails) error {
	if details.Name == "" {
		return logs.Error("name is required")
	}

	if details.Namespace == "" {
		return logs.Error("namespace is required")
	}

	if details.ContainerURL == "" {
		return logs.Error("container_url is required")
	}

	if details.Hash == "" && details.Tag == "" {
		return logs.Error("hash or tag is required")
	}

	i.RequestDetails = details

	deps := i.ClientSet.AppsV1().Deployments(i.RequestDetails.Namespace)
	deployment, err := deps.Get(i.Context, i.RequestDetails.Name, metav1.GetOptions{})
	if err != nil {
		return logs.Errorf("failed to get deployment: %v", err)
	}

	if i.RequestDetails.Tag != "" {
		deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", i.RequestDetails.ContainerURL, i.RequestDetails.Tag)
	}
	if i.RequestDetails.Hash != "" {
		deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", i.RequestDetails.ContainerURL, i.RequestDetails.Hash)
	}

	_, err = deps.Update(i.Context, deployment, metav1.UpdateOptions{})
	if err != nil {
		return logs.Errorf("failed to update deployment: %v", err)
	}

	i.UpdateStatus = true

	return nil
}

func (i *ImageRequest) SendResponse() error {
	fmt.Printf("send response: %v\n", i.UpdateStatus)

	return nil
}
