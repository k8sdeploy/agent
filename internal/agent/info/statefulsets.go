package info

import (
	"context"
	"encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type StatefulSetsRequest struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	RequestID string
	Response  *StatefulSetsResponse
}

type StatefulSetInfo struct {
	Name            string `json:"name"`
	ReadyReplicas   int32  `json:"ready_replicas"`
	CurrentReplicas int32  `json:"current_replicas"`
	Image           string `json:"image"`
}

type StatefulSetsResponse struct {
	RequestID    string            `json:"request_id"`
	Namespace    string            `json:"namespace"`
	StatefulSets []StatefulSetInfo `json:"statefulsets"`
}

func NewStatefulSets(cs *kubernetes.Clientset, ctx context.Context) *StatefulSetsRequest {
	return &StatefulSetsRequest{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (d *StatefulSetsRequest) SetRequestID(rid string) {
	d.RequestID = rid
}

func (d *StatefulSetsRequest) ProcessRequest(details *RequestDetails) error {
	dp, err := d.GetStatefulSets(details.Namespace)
	if err != nil {
		return logs.Errorf("failed to get statefulsets: %v", err)
	}

	d.Response = &StatefulSetsResponse{
		Namespace:    details.Namespace,
		StatefulSets: dp,
	}

	return nil
}

func (d *StatefulSetsRequest) GetResponse() (string, error) {
	d.Response.RequestID = d.RequestID
	r, err := json.Marshal(d.Response)
	if err != nil {
		return "", logs.Errorf("failed to marshal statefulsets response: %v", err)
	}

	return string(r), nil
}

func (d *StatefulSetsRequest) GetStatefulSets(namespace string) ([]StatefulSetInfo, error) {
	var statefulSets []StatefulSetInfo
	sts, err := d.ClientSet.AppsV1().StatefulSets(namespace).List(d.Context, metav1.ListOptions{})
	if err != nil {
		return statefulSets, logs.Errorf("failed to get statefulsets: %v", err)
	}

	for _, s := range sts.Items {
		statefulSets = append(statefulSets, StatefulSetInfo{
			Name:            s.ObjectMeta.Name,
			ReadyReplicas:   s.Status.ReadyReplicas,
			CurrentReplicas: s.Status.CurrentReplicas,
			Image:           s.Spec.Template.Spec.Containers[0].Image,
		})
	}

	return statefulSets, nil
}
