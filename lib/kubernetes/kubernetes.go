package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeapi "k8s.io/client-go/kubernetes"
	kubeconfig "k8s.io/client-go/tools/clientcmd"
)

type Kubernetes interface {
	GetPods() (*corev1.PodList, error)
}

type kubernetesImpl struct {
	client *kubeapi.Clientset
}

func NewFromKubeconfig(config string) (Kubernetes, error) {
	kubeconfig, err := kubeconfig.RESTConfigFromKubeConfig([]byte(config))
	if err != nil {
		return nil, err
	}

	client, err := kubeapi.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return &kubernetesImpl{
		client: client,
	}, nil
}

func (self *kubernetesImpl) GetPods() (*corev1.PodList, error) {
	return self.client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
}
