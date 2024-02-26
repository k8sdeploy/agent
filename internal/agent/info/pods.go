package info

import (
	"context"
	"encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

type PodsRequest struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	RequestID string
	Response  *PodsResponse
}

type PodInfo struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Image    string `json:"image"`
	Restarts int32  `json:"restarts"`

	StartedAt time.Time   `json:"started_at"`
	Metrics   interface{} `json:"metrics"`
}

type PodsResponse struct {
	RequestID string    `json:"request_id"`
	Namespace string    `json:"namespace"`
	Pods      []PodInfo `json:"pods"`
}

func NewPods(cs *kubernetes.Clientset, ctx context.Context) *PodsRequest {
	return &PodsRequest{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (d *PodsRequest) SetRequestID(rid string) {
	d.RequestID = rid
}

func (d *PodsRequest) ProcessRequest(details *RequestDetails) error {
	dp, err := d.GetPods(details.Namespace)
	if err != nil {
		return logs.Errorf("failed to get pods: %v", err)
	}

	d.Response = &PodsResponse{
		Namespace: details.Namespace,
		Pods:      dp,
	}

	return nil
}

func (d *PodsRequest) GetResponse() (string, error) {
	d.Response.RequestID = d.RequestID
	r, err := json.Marshal(d.Response)
	if err != nil {
		return "", logs.Errorf("failed to marshal response: %v", err)
	}

	return string(r), nil
}

func (d *PodsRequest) GetPods(namespace string) ([]PodInfo, error) {
	var pods []PodInfo

	podList, err := d.ClientSet.CoreV1().Pods(namespace).List(d.Context, metav1.ListOptions{})
	if err != nil {
		return nil, logs.Errorf("failed to get pods: %v", err)
	}

	for _, pod := range podList.Items {
		pods = append(pods, PodInfo{
			Name:      pod.Name,
			Status:    string(pod.Status.Phase),
			Image:     pod.Spec.Containers[0].Image,
			Restarts:  pod.Status.ContainerStatuses[0].RestartCount,
			StartedAt: pod.Status.StartTime.Time,
		})
	}

	return pods, nil
}
