package info

import (
	"context"
	"encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"

	"k8s.io/client-go/kubernetes"
)

type Info struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	Type      InfoType
	RequestID string
}

type InfoType string

const (
	namespaceRequestType   InfoType = "namespaces"
	deploymentsRequestType InfoType = "deployments"
	deploymentRequestType  InfoType = "deployment"
)

func NewInfo(cs *kubernetes.Clientset, ctx context.Context, it InfoType, rid string) *Info {
	return &Info{
		ClientSet: cs,
		Context:   ctx,

		Type:      it,
		RequestID: rid,
	}
}

type RequestDetails struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type System interface {
	ProcessRequest(details RequestDetails) error
	SendResponse() error
}

func (i Info) ParseRequest(infoRequest interface{}) error {
	jd, err := json.Marshal(infoRequest)
	if err != nil {
		return logs.Errorf("failed to marshal deployment request: %v", err)
	}

	var infoDetails RequestDetails
	if err := json.Unmarshal(jd, &infoDetails); err != nil {
		return logs.Errorf("failed to unmarshal deployment request: %v", err)
	}

	var is System
	switch i.Type {
	case namespaceRequestType:
		is = NewNamespaces(i.ClientSet, i.Context, i.RequestID)
	case deploymentsRequestType:
		is = NewDeployments(i.ClientSet, i.Context, i.RequestID)
	case deploymentRequestType:
		is = NewDeployment(i.ClientSet, i.Context, i.RequestID)
	}

	if err := is.ProcessRequest(infoDetails); err != nil {
		return logs.Errorf("failed to parse request: %v", err)
	}

	if err := is.SendResponse(); err != nil {
		return logs.Errorf("failed to send response: %v", err)
	}

	return nil
}
