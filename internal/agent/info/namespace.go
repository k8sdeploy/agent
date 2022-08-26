package info

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const NAMESPACES_REQUERST_TYPE = "namespaces"

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
		InfoType:  NAMESPACES_REQUERST_TYPE,
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
		return err
	}

	fmt.Sprintf("%s", rs)

	return nil
}

func (n *NamespaceRequest) getNamespaces() (*NamespaceSendResponse, error) {
	namespaces, err := n.Clientset.CoreV1().Namespaces().List(n.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	ret := make([]string, 0)
	for _, namespace := range namespaces.Items {
		ret = append(ret, namespace.Name)
	}

	return &NamespaceSendResponse{
		Namespaces: ret,
	}, nil
}
