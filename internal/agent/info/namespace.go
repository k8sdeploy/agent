package info

import (
	"context"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	namespaceRequestType = "namespaces"
)

type NamespaceSendResponse struct {
	Namespaces []string `json:"namespaces"`
}

type NamespaceRequest struct {
	InfoType  string `json:"info_type"`
	Clientset *kubernetes.Clientset
	Context   context.Context
}

func NewNamespaces(clientset *kubernetes.Clientset, ctx context.Context) *NamespaceRequest {
	return &NamespaceRequest{
		InfoType:  namespaceRequestType,
		Clientset: clientset,
		Context:   ctx,
	}
}

func (n *NamespaceRequest) ParseRequest(msgMap map[string]interface{}) error {
	return nil
}

func (n *NamespaceRequest) SendResponse() error {
	rs, err := n.getNamespaces()
	if err != nil {
		return logs.Errorf("failed to get namespaces: %v", err)
	}
	fmt.Printf("%+v\n", rs)

	return nil
}

func (n *NamespaceRequest) getNamespaces() (*NamespaceSendResponse, error) {
	namespaces, err := n.Clientset.CoreV1().Namespaces().List(n.Context, metav1.ListOptions{})
	if err != nil {
		return nil, logs.Errorf("failed to get namespaces: %v", err)
	}

	ret := make([]string, 0)
	for _, namespace := range namespaces.Items {
		ret = append(ret, namespace.Name)
	}

	return &NamespaceSendResponse{
		Namespaces: ret,
	}, nil
}
