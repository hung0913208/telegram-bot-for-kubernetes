package platform

import (
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

type pgImpl struct {
	client kubernetes.Kubernetes
	pod    corev1.Pod
}

func GetPgFromPodList(
	client kubernetes.Kubernetes,
	podList ...*corev1.PodList,
) ([]Pg, error) {
	var (
		list *corev1.PodList
		err  error
	)

	if len(podList) > 0 {
		list = podList[0]
	} else {
		list, err = client.GetPods("")

		if err != nil {
			return nil, err
		}
	}

	ret := make([]Pg, 0)
	for _, pod := range list.Items {
		name, ok := pod.Labels["app.kubernetes.io/name"]
		if !ok || name != "postgresql" {
			continue
		}

		instance, err := NewPg(client, pod)
		if err != nil {
			return nil, err
		}

		ret = append(ret, instance)
	}

	return ret, nil
}

func NewPg(client kubernetes.Kubernetes, pod corev1.Pod) (Pg, error) {
	return &pgImpl{
		client: client,
		pod:    pod,
	}, nil
}

func (self *pgImpl) Backup() error {
	return nil
}
