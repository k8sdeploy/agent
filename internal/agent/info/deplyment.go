package info

import (
	"context"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"
)

type DeploymentRequest struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	ClientSet *kubernetes.Clientset
	Context   context.Context

	Response  *DeploymentResponse
	RequestID string
}

type Replicas struct {
	SetName     string `json:"set_name"`
	Available   int32  `json:"available"`
	Ready       int32  `json:"ready"`
	Total       int32  `json:"total"`
	Unavailable int32  `json:"unavailable"`
}

type DeploymentResponse struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Image     string    `json:"image"`
	Version   string    `json:"version"`
	Replicas  Replicas  `json:"replicas"`
	Pods      []PodInfo `json:"pods"`
}

type PodInfo struct {
	Name      string      `json:"name"`
	Restarts  int32       `json:"restarts"`
	StartedAt time.Time   `json:"started_at"`
	Metrics   interface{} `json:"metrics"`
}

func NewDeployment(cs *kubernetes.Clientset, ctx context.Context, rid string) *DeploymentRequest {
	return &DeploymentRequest{
		ClientSet: cs,
		Context:   ctx,

		RequestID: rid,
	}
}

func (v *DeploymentRequest) ProcessRequest(details RequestDetails) error {
	if details.Namespace == "" {
		return logs.Error("namespace is required")
	}

	if details.Name == "" {
		return fmt.Errorf("name is required")
	}

	deps, err := v.getDeployment(details.Name, details.Namespace)
	if err != nil {
		return logs.Errorf("failed to get deployment: %v", err)
	}
	v.Response = deps

	return nil
}

func (v *DeploymentRequest) SendResponse() error {
	fmt.Printf("deployment: %+v\n", v.Response)

	return nil
}

func (v *DeploymentRequest) getDeployment(name, namespace string) (*DeploymentResponse, error) {
	dep, err := v.ClientSet.AppsV1().Deployments(namespace).Get(v.Context, name, metav1.GetOptions{})
	if err != nil {
		return nil, logs.Errorf("failed to get deployment: %v", err)
	}

	i := dep.Spec.Template.Spec.Containers[0].Image
	version := strings.Split(i, ":")[1]

	repName, err := v.getReplicaSet(name, namespace)
	if err != nil {
		return nil, logs.Errorf("failed to get replica set: %v", err)
	}

	pods, err := v.getPods(namespace, repName)
	if err != nil {
		return nil, logs.Errorf("failed to get pods: %v", err)
	}

	return &DeploymentResponse{
		Name:      v.Name,
		Namespace: v.Namespace,
		Version:   version,
		Image:     i,
		Replicas: Replicas{
			SetName:     repName,
			Available:   dep.Status.AvailableReplicas,
			Ready:       dep.Status.ReadyReplicas,
			Total:       dep.Status.Replicas,
			Unavailable: dep.Status.UnavailableReplicas,
		},
		Pods: pods,
	}, nil
}

func (v *DeploymentRequest) getReplicaSet(name, ns string) (string, error) {
	reps, err := v.ClientSet.AppsV1().ReplicaSets(ns).List(v.Context, metav1.ListOptions{})
	if err != nil {
		return "nil", logs.Errorf("failed to get replica set: %v", err)
	}
	var repName string
	for _, rep := range reps.Items {
		for _, owner := range rep.OwnerReferences {
			if owner.Name == name {
				repName = rep.Name
			}
		}
	}

	return repName, nil
}

func (v *DeploymentRequest) getPods(ns, replicaSetName string) ([]PodInfo, error) {
	allPods, err := v.ClientSet.CoreV1().Pods(ns).List(v.Context, metav1.ListOptions{})
	if err != nil {
		return nil, logs.Errorf("failed to get pods: %v", err)
	}

	pods := make([]PodInfo, 0)
	for _, pod := range allPods.Items {
		for _, owner := range pod.OwnerReferences {
			if owner.Name == replicaSetName {
				pod, err := v.ClientSet.CoreV1().Pods(ns).Get(v.Context, pod.Name, metav1.GetOptions{})
				if err != nil {
					return nil, logs.Errorf("failed to get pod: %v", err)
				}
				podInfo := PodInfo{
					Name:      pod.Name,
					Restarts:  pod.Status.ContainerStatuses[0].RestartCount,
					StartedAt: pod.Status.StartTime.Time,
				}
				pods = append(pods, podInfo)
			}
		}
	}

	for _, pod := range pods {
		metric, err := v.getMetrics(pod)
		if err != nil {
			return nil, logs.Errorf("failed to get metrics: %v", err)
		}
		pod.Metrics = metric
	}

	return pods, nil
}

func (v *DeploymentRequest) getMetrics(info PodInfo) (interface{}, error) {
	return nil, nil
}
