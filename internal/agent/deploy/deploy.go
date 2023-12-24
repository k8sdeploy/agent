package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"

	"k8s.io/client-go/kubernetes"
)

type DeployType string

const (
	imageRequestType DeployType = "image"
)

type Deployment struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context
	Type      DeployType
	RequestID string
}

type RequestDetails struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	ContainerURL string `json:"container_url"`
	Hash         string `json:"hash"`
	Tag          string `json:"tag"`
}

func NewDeployment(cs *kubernetes.Clientset, ctx context.Context, dt DeployType, rid string) *Deployment {
	return &Deployment{
		ClientSet: cs,
		Context:   ctx,

		Type:      dt,
		RequestID: rid,
	}
}

type System interface {
	ProcessRequest(details RequestDetails) error
	SendResponse() error
}

func (d *Deployment) ParseRequest(deploymentRequest interface{}) error {
	jd, err := json.Marshal(deploymentRequest)
	if err != nil {
		return logs.Errorf("failed to marshal deployment request: %v", err)
	}

	var deployDetails RequestDetails
	if err := json.Unmarshal(jd, &deployDetails); err != nil {
		return logs.Errorf("failed to unmarshal deployment request: %v", err)
	}

	var is System
	switch d.Type {
	case imageRequestType:
		is = NewImage(d.ClientSet, d.Context, d.RequestID)
	default:
		return fmt.Errorf("unknown deployment_type: %s", d.Type)
	}

	if err := is.ProcessRequest(deployDetails); err != nil {
		return logs.Errorf("failed to parse request: %v", err)
	}

	if err := is.SendResponse(); err != nil {
		return logs.Errorf("failed to send response: %v", err)
	}

	return nil
}
