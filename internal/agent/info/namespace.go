package info

import (
	"context"
	"encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NamespaceSendResponse struct {
	RequestID  string   `json:"request_id"`
	Namespaces []string `json:"namespaces"`
}

type NamespaceRequest struct {
	Clientset *kubernetes.Clientset
	Context   context.Context

	RequestID string
	Response  *NamespaceSendResponse
}

func NewNamespaces(cs *kubernetes.Clientset, ctx context.Context) *NamespaceRequest {
	return &NamespaceRequest{
		Clientset: cs,
		Context:   ctx,
	}
}

func (n *NamespaceRequest) SetRequestID(rid string) {
	n.RequestID = rid
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

func (n *NamespaceRequest) GetResponse() (string, error) {
	n.Response.RequestID = n.RequestID

	jd, err := json.Marshal(n.Response)
	if err != nil {
		return "", logs.Errorf("failed to marshal response: %v", err)
	}

	return string(jd), nil
}
