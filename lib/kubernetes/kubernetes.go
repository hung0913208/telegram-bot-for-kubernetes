package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeapi "k8s.io/client-go/kubernetes"
	kubeconfig "k8s.io/client-go/tools/clientcmd"
)

type Kubernetes interface {
	GetPods(namespace string) (*corev1.PodList, error)
	GetAppPods(namespace string) (*corev1.PodList, error)
	GetInfraPods(namespace string) (*corev1.PodList, error)
}

type kubernetesImpl struct {
	client *kubeapi.Clientset
}

func NewFromKubeconfig(config []byte) (Kubernetes, error) {
	kubeconfig, err := kubeconfig.RESTConfigFromKubeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("fail read kubeconfig: %v", err)
	}

	client, err := kubeapi.NewForConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("fail new client: %v", err)
	}

	return &kubernetesImpl{
		client: client,
	}, nil
}

func (self *kubernetesImpl) GetPods(namespace string) (*corev1.PodList, error) {
	return self.client.CoreV1().Pods(namespace).
		List(context.TODO(), metav1.ListOptions{})
}

func (self *kubernetesImpl) GetAppPods(namespace string) (*corev1.PodList, error) {
	return self.client.CoreV1().Pods(namespace).
		List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/managed-by notin (Helm)",
			},
		)
}

func (self *kubernetesImpl) GetInfraPods(namespace string) (*corev1.PodList, error) {
	return self.client.CoreV1().Pods(namespace).
		List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/managed-by=Helm",
			},
		)
}
