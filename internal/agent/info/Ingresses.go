package info

import (
	"context"
	"encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type IngressRequest struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	Response  *IngressResponse
	RequestID string
}

type IngressResponse struct {
	RequestID string        `json:"request_id"`
	Namespace string        `json:"namespace"`
	Ingresses []IngressInfo `json:"ingresses"`
}

type IngressInfo struct {
	Name      string   `json:"name"`
	Hosts     []string `json:"hosts"`
	Endpoints []string `json:"endpoints"`
}

func NewIngress(cs *kubernetes.Clientset, ctx context.Context) *IngressRequest {
	return &IngressRequest{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (i *IngressRequest) SetRequestID(rid string) {
	i.RequestID = rid
}

func (i *IngressRequest) ProcessRequest(details *RequestDetails) error {
	ing, err := i.GetIngress(details.Namespace)
	if err != nil {
		return logs.Errorf("failed to get ingress: %v", err)
	}

	i.Response = &IngressResponse{
		Namespace: details.Namespace,
		Ingresses: ing,
	}

	return nil
}

func (i *IngressRequest) GetResponse() (string, error) {
	i.Response.RequestID = i.RequestID

	jd, err := json.Marshal(i.Response)
	if err != nil {
		return "", logs.Errorf("failed to marshal response: %v", err)
	}

	return string(jd), nil
}

func (i *IngressRequest) GetIngress(namespace string) ([]IngressInfo, error) {
	ing, err := i.ClientSet.NetworkingV1().Ingresses(namespace).List(i.Context, metav1.ListOptions{})
	if err != nil {
		return nil, logs.Errorf("failed to get ingress: %v", err)
	}

	var ingresses []IngressInfo
	for _, i := range ing.Items {
		var hosts []string
		var endpoints []string
		for _, h := range i.Spec.Rules {
			hosts = append(hosts, h.Host)
		}
		for _, e := range i.Status.LoadBalancer.Ingress {
			endpoints = append(endpoints, e.IP)
		}
		ingresses = append(ingresses, IngressInfo{
			Name:      i.Name,
			Hosts:     hosts,
			Endpoints: endpoints,
		})
	}

	return ingresses, nil
}
