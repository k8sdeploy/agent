package info

import (
	"context"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NamespaceSendResponse struct {
	Namespaces []string `json:"namespaces"`
}

type NamespaceRequest struct {
	Clientset *kubernetes.Clientset
	Context   context.Context

	RequestID string
	Response  *NamespaceSendResponse
}

func NewNamespaces(cs *kubernetes.Clientset, ctx context.Context, rid string) *NamespaceRequest {
	return &NamespaceRequest{
		Clientset: cs,
		Context:   ctx,

		RequestID: rid,
	}
}

func (n *NamespaceRequest) ProcessRequest(id RequestDetails) error {
	namespaces, err := n.Clientset.CoreV1().Namespaces().List(n.Context, metav1.ListOptions{})
	if err != nil {
		return logs.Errorf("failed to get namespaces: %v", err)
	}

	ret := make([]string, 0)
	for _, namespace := range namespaces.Items {
		ret = append(ret, namespace.Name)
	}

	n.Response = &NamespaceSendResponse{
		Namespaces: ret,
	}

	return nil
}

func (n *NamespaceRequest) SendResponse() error {
	fmt.Printf("namespaces: %+v\n", n.Response)

	return nil
}
