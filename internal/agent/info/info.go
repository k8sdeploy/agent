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

	Type      TypeInfo
	RequestID string

	Response string
}

type TypeInfo string

const (
	namespaceRequestType   TypeInfo = "namespaces"
	deploymentsRequestType TypeInfo = "deployments"
	deploymentRequestType  TypeInfo = "deployment"
)

func NewInfo(cs *kubernetes.Clientset, ctx context.Context) *Info {
	return &Info{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (i *Info) SetInfoType(it TypeInfo) {
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
	ProcessRequest(details *RequestDetails) error
	GetResponse() (string, error)
}

func requestToInfo(infoRequest interface{}) (*RequestDetails, error) {
	ir, err := json.Marshal(infoRequest)
	if err != nil {
		return nil, logs.Errorf("failed to marshal deployment request: %v", err)
	}

	jd := &RequestDetails{}
	if err := json.Unmarshal(ir, jd); err != nil {
		return nil, logs.Errorf("failed to unmarshal deployment request: %v", err)
	}

	return jd, nil
}

func (i *Info) createSystem(clientSet *kubernetes.Clientset, context context.Context, infoType TypeInfo) (System, error) {
	var is System
	switch infoType {
	case namespaceRequestType:
		is = NewNamespaces(clientSet, context)
	case deploymentsRequestType:
		is = NewDeployments(clientSet, context)
	case deploymentRequestType:
		is = NewDeployment(clientSet, context)
	default:
		return nil, logs.Errorf("unknown info type: %s", infoType)
	}

	is.SetRequestID(i.RequestID)

	return is, nil
}

func (i *Info) ParseRequest(infoRequest interface{}) error {
	infoDetails, err := requestToInfo(infoRequest)
	if err != nil {
		return logs.Errorf("failed to marshal deployment request: %v", err)
	}

	is, err := i.createSystem(i.ClientSet, i.Context, i.Type)
	if err != nil {
		return logs.Errorf("failed to create system: %v", err)
	}

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
