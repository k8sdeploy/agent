package info

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceRequest struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	Response  *ServiceResponse
	RequestID string
}

type ServiceResponse struct {
	RequestID string        `json:"request_id"`
	Namespace string        `json:"namespace"`
	Services  []ServiceInfo `json:"services"`
}

type ServiceInfo struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	ClusterIP         string   `json:"cluster_ip"`
	InternalEndpoints []string `json:"internal_endpoints"`
	ExternalEndpoints []string `json:"external_endpoints"`
}

func NewService(cs *kubernetes.Clientset, ctx context.Context) *ServiceRequest {
	return &ServiceRequest{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (s *ServiceRequest) SetRequestID(rid string) {
	s.RequestID = rid
}

func (s *ServiceRequest) ProcessRequest(details *RequestDetails) error {
	svc, err := s.GetServices(details.Namespace)
	if err != nil {
		return logs.Errorf("failed to get services: %v", err)
	}

	s.Response = &ServiceResponse{
		Namespace: details.Namespace,
		Services:  svc,
	}

	return nil
}

func (s *ServiceRequest) GetResponse() (string, error) {
	s.Response.RequestID = s.RequestID

	jd, err := json.Marshal(s.Response)
	if err != nil {
		return "", logs.Errorf("failed to marshal response: %v", err)
	}

	return string(jd), nil
}

func (s *ServiceRequest) GetServices(namespace string) ([]ServiceInfo, error) {
	svc, err := s.ClientSet.CoreV1().Services(namespace).List(s.Context, metav1.ListOptions{})
	if err != nil {
		return nil, logs.Errorf("failed to get services: %v", err)
	}

	var services []ServiceInfo
	for _, s := range svc.Items {
		ies := []string{}
		for _, ie := range s.Spec.Ports {
			ies = append(ies, fmt.Sprintf("%s.%s:%d", s.ObjectMeta.Name, s.ObjectMeta.Namespace, ie.Port))
		}
		ies = append(ies, fmt.Sprintf("%s.%s:0", s.ObjectMeta.Name, s.ObjectMeta.Namespace))

		services = append(services, ServiceInfo{
			Name:              s.Name,
			Type:              string(s.Spec.Type),
			ClusterIP:         s.Spec.ClusterIP,
			InternalEndpoints: ies,
			ExternalEndpoints: s.Spec.ExternalIPs,
		})
	}

	return services, nil
}
