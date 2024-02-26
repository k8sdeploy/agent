package deploy

//
//import (
//	"fmt"
//	"github.com/bugfixes/go-bugfixes/logs"
//	"github.com/hashicorp/vault/sdk/helper/pointerutil"
//	appsv1 "k8s.io/api/apps/v1"
//	apiv1 "k8s.io/api/core/v1"
//	netv1 "k8s.io/api/networking/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/util/intstr"
//)
//
//func (i *ImageRequest) createDeployment() error {
//	containerURL := i.RequestDetails.Image.ContainerURL
//	if i.RequestDetails.Image.Tag != "" {
//		containerURL = fmt.Sprintf("%s:%s", containerURL, i.RequestDetails.Image.Tag)
//	} else if i.RequestDetails.Image.Hash != "" {
//		containerURL = fmt.Sprintf("%s:%s", containerURL, i.RequestDetails.Image.Hash)
//	} else {
//		containerURL = fmt.Sprintf("%s:latest", containerURL)
//	}
//
//	deployment := &appsv1.Deployment{
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Deployment",
//			APIVersion: "apps/v1",
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      i.RequestDetails.Kube.Name,
//			Namespace: i.RequestDetails.Kube.Namespace,
//			Labels: map[string]string{
//				"app":  i.RequestDetails.Kube.Name,
//				"name": i.RequestDetails.Kube.Name,
//			},
//		},
//		Spec: appsv1.DeploymentSpec{
//			Replicas: func() *int32 {
//				i := int32(2)
//				return &i
//			}(),
//			Strategy: appsv1.DeploymentStrategy{
//				RollingUpdate: &appsv1.RollingUpdateDeployment{
//					MaxSurge: &intstr.IntOrString{
//						Type:   intstr.Int,
//						IntVal: 2,
//					},
//					MaxUnavailable: &intstr.IntOrString{
//						Type:   intstr.Int,
//						IntVal: 1,
//					},
//				},
//			},
//			Selector: &metav1.LabelSelector{
//				MatchLabels: map[string]string{
//					"app":  i.RequestDetails.Kube.Name,
//					"name": i.RequestDetails.Kube.Name,
//				},
//			},
//			Template: apiv1.PodTemplateSpec{
//				ObjectMeta: metav1.ObjectMeta{
//					Labels: map[string]string{
//						"app":  i.RequestDetails.Kube.Name,
//						"name": i.RequestDetails.Kube.Name,
//					},
//				},
//				Spec: buildPodSpec(containerURL, i.RequestDetails.Kube.Name, i.RequestDetails.DeploymentDetails),
//			},
//		},
//	}
//	_, err := i.ClientSet.AppsV1().Deployments(i.RequestDetails.Kube.Namespace).Create(i.Context, deployment, metav1.CreateOptions{})
//	if err != nil {
//		return logs.Errorf("failed to create deployment: %v", err)
//	}
//
//	return nil
//}
//
//func (i *ImageRequest) createService() error {
//	if i.RequestDetails.DeploymentDetails.Services == nil {
//		return nil
//	}
//
//	s := &apiv1.Service{
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Service",
//			APIVersion: "v1",
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      i.RequestDetails.Kube.Name,
//			Namespace: i.RequestDetails.Kube.Namespace,
//			Labels: map[string]string{
//				"app":  i.RequestDetails.Kube.Name,
//				"name": i.RequestDetails.Kube.Name,
//			},
//		},
//		Spec: apiv1.ServiceSpec{
//			Selector: map[string]string{
//				"app":  i.RequestDetails.Kube.Name,
//				"name": i.RequestDetails.Kube.Name,
//			},
//		},
//	}
//
//	for _, service := range i.RequestDetails.DeploymentDetails.Services {
//		s.Spec.Ports = append(s.Spec.Ports, apiv1.ServicePort{
//			Name:       service.Name,
//			Port:       int32(service.Port),
//			TargetPort: intstr.FromInt32(int32(service.TargetPort)),
//		})
//	}
//
//	_, err := i.ClientSet.CoreV1().Services(i.RequestDetails.Kube.Namespace).Create(i.Context, s, metav1.CreateOptions{})
//	if err != nil {
//		return logs.Errorf("failed to create service: %v", err)
//	}
//
//	return nil
//}
//
//func (i *ImageRequest) createIngress() error {
//	if i.RequestDetails.DeploymentDetails.Ingress == nil {
//		return nil
//	}
//
//	ing := &netv1.Ingress{
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Ingress",
//			APIVersion: "networking.k8s.io/v1",
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      i.RequestDetails.Kube.Name,
//			Namespace: i.RequestDetails.Kube.Namespace,
//			Labels: map[string]string{
//				"app":  i.RequestDetails.Kube.Name,
//				"name": i.RequestDetails.Kube.Name,
//			},
//		},
//		Spec: netv1.IngressSpec{
//			IngressClassName: pointerutil.StringPtr("nginx"),
//			TLS: []netv1.IngressTLS{
//				{
//					Hosts: []string{
//						i.RequestDetails.DeploymentDetails.Ingress.Ingresses[0].Hostname,
//					},
//					SecretName: fmt.Sprintf("%s-tls", i.RequestDetails.Kube.Name),
//				},
//			},
//		},
//	}
//
//	annotations := map[string]string{}
//	if i.RequestDetails.DeploymentDetails.Ingress.CertIssuer != "" {
//		annotations["cert-manager.io/cluster-issuer"] = i.RequestDetails.DeploymentDetails.Ingress.CertIssuer
//	}
//
//	if i.RequestDetails.DeploymentDetails.Ingress.RewriteTarget != "" {
//		annotations["nginx.ingress.kubernetes.io/rewrite-target"] = i.RequestDetails.DeploymentDetails.Ingress.RewriteTarget
//	}
//	ing.ObjectMeta.Annotations = annotations
//
//	for _, ingress := range i.RequestDetails.DeploymentDetails.Ingress.Ingresses {
//		if ingress.Port == 0 {
//			ingress.Port = i.RequestDetails.DeploymentDetails.Services[0].Port
//		}
//
//		ing.Spec.Rules = append(ing.Spec.Rules, netv1.IngressRule{
//			Host: ingress.Hostname,
//			IngressRuleValue: netv1.IngressRuleValue{
//				HTTP: &netv1.HTTPIngressRuleValue{
//					Paths: []netv1.HTTPIngressPath{
//						{
//							Path: ingress.Path,
//							PathType: func() *netv1.PathType {
//								p := netv1.PathTypePrefix
//								return &p
//							}(),
//							Backend: netv1.IngressBackend{
//								Service: &netv1.IngressServiceBackend{
//									Name: i.RequestDetails.Kube.Name,
//									Port: netv1.ServiceBackendPort{
//										Number: int32(ingress.Port),
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//		})
//	}
//
//	_, err := i.ClientSet.NetworkingV1().Ingresses(i.RequestDetails.Kube.Namespace).Create(i.Context, ing, metav1.CreateOptions{})
//	if err != nil {
//		return logs.Errorf("failed to create ingress: %v", err)
//	}
//
//	return nil
//}
//
//func buildPodSpec(containerURL, kubeName string, depDetails DeploymentDetails) apiv1.PodSpec {
//	podSpec := apiv1.PodSpec{
//		Containers: []apiv1.Container{
//			{
//				Name:            kubeName,
//				Image:           containerURL,
//				ImagePullPolicy: apiv1.PullAlways,
//			},
//		},
//	}
//
//	if len(depDetails.Ports) > 0 {
//		for _, port := range depDetails.Ports {
//			podSpec.Containers[0].Ports = append(podSpec.Containers[0].Ports, apiv1.ContainerPort{
//				ContainerPort: int32(port),
//				Protocol:      apiv1.ProtocolTCP,
//			})
//		}
//	}
//
//	if len(depDetails.Env) > 0 {
//		for _, env := range depDetails.Env {
//			if env.ValueFrom != nil {
//				podSpec.Containers[0].Env = append(podSpec.Containers[0].Env, apiv1.EnvVar{
//					Name: env.Name,
//					ValueFrom: &apiv1.EnvVarSource{
//						SecretKeyRef: &apiv1.SecretKeySelector{
//							LocalObjectReference: apiv1.LocalObjectReference{
//								Name: env.ValueFrom.SecretKeyRef.Name,
//							},
//							Key: env.ValueFrom.SecretKeyRef.Key,
//						},
//					},
//				})
//				continue
//			}
//
//			podSpec.Containers[0].Env = append(podSpec.Containers[0].Env, apiv1.EnvVar{
//				Name:  env.Name,
//				Value: env.Value,
//			})
//		}
//	}
//
//	if depDetails.ContainerPullSecret != "" {
//		podSpec.ImagePullSecrets = []apiv1.LocalObjectReference{
//			{
//				Name: depDetails.ContainerPullSecret,
//			},
//		}
//	}
//
//	return podSpec
//}
