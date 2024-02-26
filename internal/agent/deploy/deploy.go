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

type TypeDeploy string

const (
	imageRequestType TypeDeploy = "image"
)

type Deployment struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	Type      TypeDeploy
	RequestID string

	Response string
}

type Kube struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Image struct {
	Hash         string `json:"hash"`
	Tag          string `json:"tag"`
	ContainerURL string `json:"container_url"`
}

type Issuer struct {
	Service string `json:"service"`
	Key     string `json:"key"`
}

type RequestDetails struct {
	Kube   Kube   `json:"k8s"`
	Image  Image  `json:"image"`
	Issuer Issuer `json:"issuer"`
}

func NewDeployment(cs *kubernetes.Clientset, ctx context.Context) *Deployment {
	return &Deployment{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (d *Deployment) SetDeploymentType(dt TypeDeploy) {
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

func requestToDetails(deploymentRequest interface{}) (RequestDetails, error) {
	jd, err := json.Marshal(deploymentRequest)
	if err != nil {
		return RequestDetails{}, logs.Errorf("failed to marshal deployment request: %v", err)
	}

	var deployDetails RequestDetails
	if err := json.Unmarshal(jd, &deployDetails); err != nil {
		return RequestDetails{}, logs.Errorf("failed to unmarshal deployment request: %v", err)
	}

	return deployDetails, nil
}

func (d *Deployment) getSystem() (System, error) {
	var sys System

	switch d.Type {
	case imageRequestType:
		sys = NewImage(d.ClientSet, d.Context)
	default:
		return nil, logs.Errorf("unknown deployment_type: %s", d.Type)
	}

	sys.SetRequestID(d.RequestID)
	return sys, nil
}

func (d *Deployment) ParseRequest(deploymentRequest interface{}) error {
	deployDetails, err := requestToDetails(deploymentRequest)
	if err != nil {
		return logs.Errorf("failed to parse request: %v", err)
	}

	is, err := d.getSystem()
	if err != nil {
		return logs.Errorf("failed to get system: %v", err)
	}

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

func (d *Deployment) createPayload(cfg *config.Config) ([]byte, error) {
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
		return nil, logs.Errorf("failed to marshal payload: %v", err)
	}

	return payload, nil
}

func (d *Deployment) SendResponse(cfg *config.Config) error {
	payload, err := d.createPayload(cfg)
	if err != nil {
		return logs.Errorf("failed to create payload: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/exchanges/%s/amq.default/publish", cfg.K8sDeploy.RabbitHost, cfg.K8sDeploy.Queues.Agent), bytes.NewBuffer(payload))
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
