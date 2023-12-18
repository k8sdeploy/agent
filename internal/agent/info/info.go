package info

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"

	"k8s.io/client-go/kubernetes"
)

type Info struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context
}

type InfoRequest struct {
	Namespace  NamespaceRequest
	Deployment DeploymentsRequest
	Version    VersionRequest
}

func NewInfo(clientset *kubernetes.Clientset, ctx context.Context) *Info {
	return &Info{
		ClientSet: clientset,
		Context:   ctx,
	}
}

type InfoSystem interface {
	ParseRequest(mappedString map[string]interface{}) error
	SendResponse() error
}

func (i Info) ParseInfoRequest(infoRequest string) error {
	var msgMap map[string]interface{}
	if err := json.Unmarshal([]byte(infoRequest), &msgMap); err != nil {
		return logs.Errorf("failed to unmarshal info request: %v", err)
	}

	if msgMap["info_type"] == nil {
		return logs.Error("info_type is required")
	}

	var is InfoSystem
	switch msgMap["info_type"] {
	case namespaceRequestType:
		is = NewNamespaces(i.ClientSet, i.Context)
	case deploymentsRequestType:
		is = NewDeployments()
	case versionRequestType:
		is = NewVersion(i.ClientSet, i.Context)
	}

	if is == nil {
		return fmt.Errorf("unknown info_type: %s", msgMap["info_type"])
	}

	if err := is.ParseRequest(msgMap); err != nil {
		return logs.Errorf("failed to parse request: %v", err)
	}

	if err := is.SendResponse(); err != nil {
		return logs.Errorf("failed to send response: %v", err)
	}

	return nil
}
