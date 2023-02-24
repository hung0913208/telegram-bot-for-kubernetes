package platform

import (
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

type Backup interface {
	Perform() error
}

type backupImpl struct {
	client kubernetes.Kubernetes
	pod    corev1.Pod
}

func NewBackup(
	client kubernetes.Kubernetes,
	podList ...*corev1.PodList,
) (Backup, error) {
	if len(podList) > 0 {
		return &backupImpl{client: client}, nil
	} else {
		return &backupImpl{client: client}, nil
	}
}

func (self *backupImpl) Perform() error {
	return nil
}
