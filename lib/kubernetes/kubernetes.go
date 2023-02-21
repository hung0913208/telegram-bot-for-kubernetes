package kubernetes

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeapi "k8s.io/client-go/kubernetes"
	kubeconfig "k8s.io/client-go/tools/clientcmd"
)

//type Platform interface {
//	GetPg(namespace string) ([]Pg, error)
//	GetRedis(namespace string) ([]Redis, error)
//	GetMongo(namespace string) ([]Mongo, error)
//}

type Application interface {
	GetApplication(namespace string) ([]corev1.Pod, error)
}

type Kubernetes interface {
	Application

	GetPods(namespace string) (*corev1.PodList, error)
	GetHelmPods(namespace string) (*corev1.PodList, error)

	GetNodes() (*corev1.NodeList, error)
	GetPVs() (*corev1.PersistentVolumeList, error)

	GetPodMetrics() (*PodMetricsList, error)
	GetNodeMetrics(node string) (*NodeMetricsList, error)

	Ping() bool
	Cron(
		name string,
		cmd []string,
		image string,
		schedule, namespace string,
	) error
	Do(
		name string,
		cmd []string,
		image string,
		namespace string,
		backOffLimit int32,
	) error
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

func (self *kubernetesImpl) GetHelmPods(namespace string) (*corev1.PodList, error) {
	return self.client.CoreV1().Pods(namespace).
		List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/managed-by=Helm",
			},
		)
}

func (self *kubernetesImpl) GetPVs() (*corev1.PersistentVolumeList, error) {
	return self.client.CoreV1().PersistentVolumes().
		List(
			context.TODO(),
			metav1.ListOptions{},
		)
}

func (self *kubernetesImpl) GetNodes() (*corev1.NodeList, error) {
	return self.client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
}

func (self *kubernetesImpl) Ping() bool {
	_, err := self.client.RESTClient().
		Get().
		AbsPath("readyz?verbose").
		DoRaw(context.TODO())
	return err == nil
}

func (self *kubernetesImpl) Cron(
	name string,
	cmd []string,
	image string,
	schedule, namespace string,
) error {
	cronjobs := self.client.BatchV1().CronJobs(namespace)
	spec := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    name,
									Image:   image,
									Command: cmd,
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
			Schedule:          schedule,
		},
	}

	_, err := cronjobs.Create(
		context.TODO(),
		spec,
		metav1.CreateOptions{},
	)
	if err != nil {
		return err
	}

	return nil
}

func (self *kubernetesImpl) Do(
	name string,
	cmd []string,
	image string,
	namespace string,
	backOffLimit int32,
) error {
	jobs := self.client.BatchV1().Jobs(namespace)
	spec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    name,
							Image:   image,
							Command: cmd,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	_, err := jobs.Create(
		context.TODO(),
		spec,
		metav1.CreateOptions{},
	)
	if err != nil {
		return err
	}

	return nil
}

func (self *kubernetesImpl) renderConfigMapExecScript(
	name string,
	namespace string,
	script string,
) error {
	configmaps := self.client.CoreV1().
		ConfigMaps(namespace)
	spec := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{
			"exec.sh": fmt.Sprintf(
				"#!/bin/bash\n"+
					"\n"+
					"%s\n"+
					"",
				script,
			),
		},
	}

	_, err := configmaps.Create(context.TODO(), &spec, metav1.CreateOptions{})
	return err
}
