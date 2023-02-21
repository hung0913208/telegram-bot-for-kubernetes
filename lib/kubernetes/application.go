package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
)

func (self *kubernetesImpl) GetApplication(namespace string) ([]corev1.Pod, error) {
	pods, err := self.GetPods(namespace)
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
