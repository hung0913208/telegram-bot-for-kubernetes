package platform

import (
	"fmt"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

type Backup interface {
	Backup(interval ...string) error
}

type backupImpl struct {
	client kubernetes.Kubernetes
	pods   []corev1.Pod
	name   string
}

func NewBackup(
	name string,
	client kubernetes.Kubernetes,
	podListArgs ...*corev1.PodList,
) (Backup, error) {
	var podList *corev1.PodList
	var err error

	pods := make([]corev1.Pod, 0)

	if len(podListArgs) == 0 {
		podList, err = client.GetPods("")

		if err != nil {
			return nil, err
		}
	} else {
		podList = podListArgs[0]
	}

	for _, pod := range podList.Items {
		pods = append(pods, pod)
	}

	return &backupImpl{
		client: client,
		pods:   pods,
	}, nil
}

func (self *backupImpl) Backup(interval ...string) error {
	if len(interval) > 0 {
		err := self.client.Cron(
			fmt.Sprintf("rsync-%s-backup", self.name),
			"",
			"ubuntu",
			interval[0],
			"default",
		)
		if err != nil {
			return err
		}
	}

	err := self.client.Do(
		fmt.Sprintf("rsync-%s-backup", self.name),
		"",
		"ubuntu",
		"default",
		int32(0),
	)
	if err != nil {
		return err
	}
	return nil
}
