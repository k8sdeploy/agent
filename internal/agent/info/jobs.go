package info

import (
	"context"
	"encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type JobsRequest struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context

	RequestID string
	Response  *JobsResponse
}

type JobInfo struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	Completions int32  `json:"completions"`
	Parallelism int32  `json:"parallelism"`
	Active      int32  `json:"active"`
	Succeeded   int32  `json:"succeeded"`
	Failed      int32  `json:"failed"`
	Age         string `json:"age"`
}

type JobsResponse struct {
	RequestID string    `json:"request_id"`
	Namespace string    `json:"namespace"`
	Jobs      []JobInfo `json:"jobs"`
}

func NewJobs(cs *kubernetes.Clientset, ctx context.Context) *JobsRequest {
	return &JobsRequest{
		ClientSet: cs,
		Context:   ctx,
	}
}

func (d *JobsRequest) SetRequestID(rid string) {
	d.RequestID = rid
}

func (d *JobsRequest) ProcessRequest(details *RequestDetails) error {
	dp, err := d.GetJobs(details.Namespace)
	if err != nil {
		return logs.Errorf("failed to get jobs: %v", err)
	}

	d.Response = &JobsResponse{
		Namespace: details.Namespace,
		Jobs:      dp,
	}

	return nil
}

func (d *JobsRequest) GetResponse() (string, error) {
	d.Response.RequestID = d.RequestID
	r, err := json.Marshal(d.Response)
	if err != nil {
		return "", logs.Errorf("failed to marshal response: %v", err)
	}
	return string(r), nil
}

func (d *JobsRequest) GetJobs(namespace string) ([]JobInfo, error) {
	var jobs []JobInfo

	jobList, err := d.ClientSet.BatchV1().Jobs(namespace).List(d.Context, metav1.ListOptions{})
	if err != nil {
		return jobs, logs.Errorf("failed to get jobs: %v", err)
	}

	for _, job := range jobList.Items {
		jobs = append(jobs, JobInfo{
			Name:        job.Name,
			Image:       job.Spec.Template.Spec.Containers[0].Image,
			Completions: *job.Spec.Completions,
			Parallelism: *job.Spec.Parallelism,
			Active:      job.Status.Active,
			Succeeded:   job.Status.Succeeded,
			Failed:      job.Status.Failed,
			Age:         job.CreationTimestamp.String(),
		})
	}

	return jobs, nil
}
