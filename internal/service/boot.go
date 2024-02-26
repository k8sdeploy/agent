package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/k8sdeploy/agent/internal/agent"
	"github.com/k8sdeploy/agent/internal/agent/info"
	"github.com/k8sdeploy/agent/internal/config"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

type Boot struct {
	Config    *config.Config
	Context   context.Context
	ClientSet *kubernetes.Clientset
	ErrChan   chan error

	BootInfo *BootInfo
}

type BootInfo struct {
	Namespaces    []string
	NamespaceInfo []NamespaceInfo
}

type NamespaceInfo struct {
	Name         string
	Pods         []PodInfo
	Deployments  []DeploymentInfo
	ReplicaSets  []ReplicaSetInfo
	Services     []ServiceInfo
	Ingresses    []IngressInfo
	StatefulSets []StatefulSetInfo
	Jobs         []JobInfo
}

type PodInfo struct {
	Name   string
	Status string
	Image  string
}

type DeploymentInfo struct {
	Name      string
	Image     string
	ReadyPods int
	TotalPods int
}

type ReplicaSetInfo struct {
	Name      string
	ReadyPods int
	TotalPods int
	Image     string
}

type ServiceInfo struct {
	Name              string
	Type              string
	ClusterIP         string
	InternalEndpoints []string
	ExternalEndpoints []string
}

type IngressInfo struct {
	Name      string
	Hosts     []string
	Endpoints []string
}

type StatefulSetInfo struct {
	Name  string
	Pods  int
	Image string
}

type JobInfo struct {
	Name      string
	Image     string
	TotalPods int
	ReadyPods int
}

func NewBoot(cfg *config.Config, errChan chan error) *Boot {
	ctx := context.Background()

	a := agent.NewAgent(cfg)
	if err := a.GetKubernetesClient(); err != nil {
		errChan <- logs.Errorf("failed to get kubernetes client: %v", err)
		return nil
	}

	return &Boot{
		Config:    cfg,
		Context:   ctx,
		ClientSet: a.KubernetesClient.ClientSet,
		ErrChan:   errChan,
	}
}

func (b *Boot) GetInfo() *Boot {
	bi := &BootInfo{}

	b.GetNamespaces(bi)

	for _, namespace := range bi.Namespaces {
		ni := NamespaceInfo{
			Name:         namespace,
			Ingresses:    b.GetIngresses(namespace),
			Deployments:  b.GetDeployments(namespace),
			ReplicaSets:  b.GetReplicaSets(namespace),
			Pods:         b.GetPods(namespace),
			Services:     b.GetServices(namespace),
			StatefulSets: b.GetStatefulSets(namespace),
			Jobs:         b.GetJobs(namespace),
		}
		bi.NamespaceInfo = append(bi.NamespaceInfo, ni)
	}

	b.BootInfo = bi
	return b
}

func (b *Boot) GetNamespaces(bi *BootInfo) {
	namespaces := info.NewNamespaces(b.ClientSet, b.Context)
	names, err := namespaces.FetchAllNamespaces()
	if err != nil {
		b.ErrChan <- logs.Errorf("failed to fetch namespaces: %v", err)
	}

	bi.Namespaces = names
}

func (b *Boot) GetJobs(namespace string) []JobInfo {
	var jobInfo []JobInfo

	j := info.NewJobs(b.ClientSet, b.Context)
	jobs, err := j.GetJobs(namespace)
	if err != nil {
		b.ErrChan <- logs.Errorf("failed to get jobs: %v", err)
	}

	for _, j := range jobs {
		jobInfo = append(jobInfo, JobInfo{
			Name:  j.Name,
			Image: j.Image,
		})
	}

	return jobInfo

}

func (b *Boot) GetIngresses(namespace string) []IngressInfo {
	var ingressInfo []IngressInfo

	i := info.NewIngress(b.ClientSet, b.Context)
	ingress, err := i.GetIngress(namespace)
	if err != nil {
		b.ErrChan <- logs.Errorf("failed to get ingress: %v", err)
	}

	for _, i := range ingress {
		ingressInfo = append(ingressInfo, IngressInfo{
			Name:      i.Name,
			Hosts:     i.Hosts,
			Endpoints: i.Endpoints,
		})
	}
	return ingressInfo
}

func (b *Boot) GetDeployments(namespace string) []DeploymentInfo {
	var depInfo []DeploymentInfo

	d := info.NewDeployments(b.ClientSet, b.Context)
	deployments, err := d.GetDeployments(namespace)
	if err != nil {
		b.ErrChan <- logs.Errorf("failed to get deployments: %v", err)
	}

	for _, d := range deployments {
		depInfo = append(depInfo, DeploymentInfo{
			Name:      d.Name,
			Image:     d.Container,
			ReadyPods: int(d.ReadyReplicas),
			TotalPods: int(d.Replicas),
		})
	}

	return depInfo
}

func (b *Boot) GetReplicaSets(namespace string) []ReplicaSetInfo {
	var repInfo []ReplicaSetInfo

	r := info.NewReplicaSets(b.ClientSet, b.Context)
	replicaSets, err := r.GetReplicaSets(namespace)
	if err != nil {
		b.ErrChan <- logs.Errorf("failed to get replicasets: %v", err)
	}

	for _, r := range replicaSets {
		repInfo = append(repInfo, ReplicaSetInfo{
			Name:      r.Name,
			ReadyPods: int(r.ReadyReplicas),
			TotalPods: int(r.Replicas),
			Image:     r.Image,
		})
	}

	return repInfo
}

func (b *Boot) GetPods(namespace string) []PodInfo {
	var podInfo []PodInfo

	p := info.NewPods(b.ClientSet, b.Context)
	podList, err := p.GetPods(namespace)
	if err != nil {
		b.ErrChan <- logs.Errorf("failed to get pods: %v", err)
	}

	for _, pod := range podList {
		podInfo = append(podInfo, PodInfo{
			Name:   pod.Name,
			Status: pod.Status,
			Image:  pod.Image,
		})
	}

	return podInfo
}

func (b *Boot) GetStatefulSets(namespace string) []StatefulSetInfo {
	var statefulSetInfo []StatefulSetInfo

	s := info.NewStatefulSets(b.ClientSet, b.Context)
	statefulSets, err := s.GetStatefulSets(namespace)
	if err != nil {
		b.ErrChan <- logs.Errorf("failed to get statefulsets: %v", err)
	}

	for _, s := range statefulSets {
		statefulSetInfo = append(statefulSetInfo, StatefulSetInfo{
			Name:  s.Name,
			Pods:  int(s.ReadyReplicas),
			Image: s.Image,
		})
	}

	return statefulSetInfo
}

func (b *Boot) GetServices(namespace string) []ServiceInfo {
	var serviceInfo []ServiceInfo

	s := info.NewService(b.ClientSet, b.Context)
	services, err := s.GetServices(namespace)
	if err != nil {
		b.ErrChan <- logs.Errorf("failed to get services: %v", err)
	}

	for _, s := range services {
		serviceInfo = append(serviceInfo, ServiceInfo{
			Name:              s.Name,
			Type:              string(s.Type),
			ClusterIP:         s.ClusterIP,
			InternalEndpoints: s.InternalEndpoints,
			ExternalEndpoints: s.ExternalEndpoints,
		})
	}

	return serviceInfo
}

func (b *Boot) SendInfo() {
	c, err := json.Marshal(b.BootInfo.NamespaceInfo)
	if err != nil {
		b.ErrChan <- logs.Errorf("failed to marshal agent body: %v", err)
	}
	ci := string(c)
	logs.Infof("BootData: %s", ci)

	// Send the data to the orchestrator
	apiAddy := fmt.Sprintf("%s/agent/bootdata", b.Config.K8sDeploy.APIAddress)
	req, err := http.NewRequest("POST", apiAddy, bytes.NewBuffer(c))
	if err != nil {
		b.ErrChan <- logs.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Key", b.Config.K8sDeploy.Credentials.Agent.Key)
	req.Header.Set("X-Agent-Secret", b.Config.K8sDeploy.Credentials.Agent.Secret)

	if _, err := http.DefaultClient.Do(req); err != nil {
		b.ErrChan <- logs.Errorf("failed to send request: %v", err)
	}
	logs.Infof("Sent boot data to orchestrator: %s", apiAddy)
}
