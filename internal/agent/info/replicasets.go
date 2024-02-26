package info

import (
	"context"
	"encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ReplicaSetRequest struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	RequestID string
	Response  *ReplicaSetResponse
}

type ReplicaSetInfo struct {
	Name          string `json:"name"`
	ReadyReplicas int32  `json:"ready_replicas"`
	Replicas      int32  `json:"replicas"`
	Image         string `json:"image"`
}

type ReplicaSetResponse struct {
	RequestID   string           `json:"request_id"`
	Namespace   string           `json:"namespace"`
	ReplicaSets []ReplicaSetInfo `json:"replicasets"`
}

func NewReplicaSets(cs *kubernetes.Clientset, ctx context.Context) *ReplicaSetRequest {
	return &ReplicaSetRequest{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (d *ReplicaSetRequest) SetRequestID(rid string) {
	d.RequestID = rid
}

func (d *ReplicaSetRequest) ProcessRequest(details *RequestDetails) error {
	dp, err := d.GetReplicaSets(details.Namespace)
	if err != nil {
		return logs.Errorf("failed to get replicasets: %v", err)
	}

	d.Response = &ReplicaSetResponse{
		Namespace:   details.Namespace,
		ReplicaSets: dp,
	}

	return nil
}

func (d *ReplicaSetRequest) GetResponse() (string, error) {
	d.Response.RequestID = d.RequestID
	r, err := json.Marshal(d.Response)
	if err != nil {
		return "", logs.Errorf("failed to marshal response: %v", err)
	}

	return string(r), nil
}

func (d *ReplicaSetRequest) GetReplicaSets(namespace string) ([]ReplicaSetInfo, error) {
	var replicaSets []ReplicaSetInfo

	rs, err := d.ClientSet.AppsV1().ReplicaSets(namespace).List(d.Context, metav1.ListOptions{})
	if err != nil {
		return nil, logs.Errorf("failed to get replicasets: %v", err)
	}

	for _, r := range rs.Items {
		if r.Status.Replicas == 0 {
			continue
		}

		replicaSets = append(replicaSets, ReplicaSetInfo{
			Name:          r.Name,
			ReadyReplicas: r.Status.ReadyReplicas,
			Replicas:      r.Status.Replicas,
			Image:         r.Spec.Template.Spec.Containers[0].Image,
		})
	}

	return replicaSets, nil
}
