package info

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/k8sdeploy/agent/internal/config"
	"net/http"

	"k8s.io/client-go/kubernetes"
)

type Info struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	Type      InfoType
	RequestID string

	Response string
}

type InfoType string

const (
	namespaceRequestType   InfoType = "namespaces"
	deploymentsRequestType InfoType = "deployments"
	deploymentRequestType  InfoType = "deployment"
)

func NewInfo(cs *kubernetes.Clientset, ctx context.Context) *Info {
	return &Info{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (i *Info) SetInfoType(it InfoType) {
	i.Type = it
}

func (i *Info) SetRequestID(rid string) {
	i.RequestID = rid
}

type RequestDetails struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type System interface {
	SetRequestID(rid string)
	ProcessRequest(details RequestDetails) error
	GetResponse() (string, error)
}

func (i *Info) ParseRequest(infoRequest interface{}) error {
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
		is = NewNamespaces(i.ClientSet, i.Context)
	case deploymentsRequestType:
		is = NewDeployments(i.ClientSet, i.Context)
	case deploymentRequestType:
		is = NewDeployment(i.ClientSet, i.Context)
	}

	is.SetRequestID(i.RequestID)

	if err := is.ProcessRequest(infoDetails); err != nil {
		return logs.Errorf("failed to parse request: %v", err)
	}

	resp, err := is.GetResponse()
	if err != nil {
		return logs.Errorf("failed to get response: %v", err)
	}
	i.Response = resp

	return nil
}

func (i *Info) SendResponse(cfg *config.Config) error {
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
			RequestID: i.RequestID,
		},
		PayloadEncoding: "string",
		RoutingKey:      cfg.K8sDeploy.Queues.Response,
		Payload:         i.Response,
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
