/*
On GKE this needs "client certificate enabled" and correct RBAC setup
*/
package k8sservices

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var clientset *kubernetes.Clientset

func init() {
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("ERROR: init(): Could not get kube config in cluster. Error: %v", err.Error())
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("ERROR: init(): Could not connect to kube cluster with config. Error: %v", err.Error())
	}
}

// GetService - get information on the given service in k8s service struct
func GetService(serviceName string, k8sClient typev1.CoreV1Interface) (*corev1.Service, string, error) {
	listOptions := metav1.ListOptions{}
	serviceSlice := strings.Split(serviceName, ".")
	if len(serviceSlice) < 2 {
		return nil, "", fmt.Errorf("Service name not according to convention defined in README. Service name: %s", serviceName)
	}
	namespace := serviceSlice[1]
	svcs, err := k8sClient.Services(namespace).List(context.Background(), listOptions)
	if err != nil {
		log.Fatal(err)
	}
	svcComponents := strings.Split(serviceName, ".")
	for _, svc := range svcs.Items {
		if svc.Name == svcComponents[0] {
			return &svc, namespace, nil
		}
	}
	return nil, namespace, errors.New("cannot find service")
}

// GetPodsForSvc - List the pods for a given service
func GetPodsForSvc(svc *corev1.Service, namespace string, k8sClient typev1.CoreV1Interface) (*corev1.PodList, error) {
	set := labels.Set(svc.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := k8sClient.Pods(namespace).List(context.Background(), listOptions)
	return pods, err
}

// GetPods - Wraps GetService and GetPodsForSvc in singular function for ease of use
func GetPods(serviceName string) (*corev1.PodList, error) {
	svc, namespace, err := GetService(serviceName, clientset.CoreV1())
	if err != nil {
		return nil, err
	}
	pods, podErr := GetPodsForSvc(svc, namespace, clientset.CoreV1())
	if podErr != nil {
		return nil, err
	}
	return pods, nil
}
