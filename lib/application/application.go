package application

import (
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

type Application interface {
	GetApplication(namespace string) ([]corev1.Pod, error)
}

type applicationImpl struct {
	kubectl kubernetes.Kubernetes
}

func NewApplicationManager(
	kubectl kubernetes.Kubernetes,
) Application {
	return &applicationImpl{
		kubectl: kubectl,
	}
}

func (self *applicationImpl) GetApplication(namespace string) ([]corev1.Pod, error) {
	pods, err := self.kubectl.GetPods(namespace)
	if err != nil {
		return nil, err
	}

	ret := make([]corev1.Pod, 0)

	for _, pod := range pods.Items {
		_, foundApp := pod.Labels["app"]
		_, foundHash := pod.Labels["pod-template-hash"]

		if !foundApp || !foundHash || len(pod.Labels) > 2 {
			continue
		}

		ret = append(ret, pod)
	}
	return ret, nil
}
