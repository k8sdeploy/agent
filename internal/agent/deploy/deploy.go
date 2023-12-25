package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/k8sdeploy/agent/internal/config"
	"k8s.io/client-go/kubernetes"
	"net/http"
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

	Response string
}

type RequestDetails struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	ContainerURL string `json:"container_url"`
	Hash         string `json:"hash"`
	Tag          string `json:"tag"`
}

func NewDeployment(cs *kubernetes.Clientset, ctx context.Context) *Deployment {
	return &Deployment{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (d *Deployment) SetDeploymentType(dt DeployType) {
	d.Type = dt
}

func (d *Deployment) SetRequestID(rid string) {
	d.RequestID = rid
}

type System interface {
	SetRequestID(rid string)
	ProcessRequest(details RequestDetails) error
	GetResponse() (string, error)
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
		is = NewImage(d.ClientSet, d.Context)
	default:
		return fmt.Errorf("unknown deployment_type: %s", d.Type)
	}

	is.SetRequestID(d.RequestID)

	if err := is.ProcessRequest(deployDetails); err != nil {
		return logs.Errorf("failed to parse request: %v", err)
	}

	resp, err := is.GetResponse()
	if err != nil {
		return logs.Errorf("failed to get response: %v", err)
	}
	d.Response = resp

	return nil
}

func (d *Deployment) SendResponse(cfg *config.Config) error {
	type Props struct {
		RequestID string `json:"request_id"`
	}

	type Payload struct {
		Props           Props  `json:"properties"`
		PayloadEncoding string `json:"payload_encoding"`
		RoutingKey      string `json:"routing_key"`
		Payload         string `json:"payload"`
	}

	payload, err := json.Marshal(Payload{
		Props: Props{
			RequestID: d.RequestID,
		},
		PayloadEncoding: "string",
		RoutingKey:      cfg.K8sDeploy.Queues.Response,
		Payload:         d.Response,
	})
	if err != nil {
		return logs.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/exchanges/%s/amq.default/publish", cfg.Rabbit.Host, cfg.K8sDeploy.Queues.Agent), bytes.NewBuffer(payload))
	if err != nil {
		return logs.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(cfg.K8sDeploy.Credentials.Queue.Key, cfg.K8sDeploy.Credentials.Queue.Secret)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return logs.Errorf("failed to get events: %v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			_ = logs.Errorf("failed to close queue body: %v", err)
		}
	}()
	if res.StatusCode != http.StatusOK {
		return logs.Errorf("failed get events: %s", res.Status)
	}

	return nil
}
