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

type Ingress interface {
}

type Kubernetes interface {
	GetPods(namespace string) (*corev1.PodList, error)
	GetHelmPods(namespace string) (*corev1.PodList, error)

	GetNodes() (*corev1.NodeList, error)
	GetPVs() (*corev1.PersistentVolumeList, error)

	GetPodMetrics() (*PodMetricsList, error)
	GetNodeMetrics(node string) (*NodeMetricsList, error)

	Ping() bool
	Cron(
		name string,
		script string,
		image string,
		schedule, namespace string,
	) error
	Do(
		name string,
		script string,
		image string,
		namespace string,
		backOffLimit int32,
	) error
}

type kubernetesImpl struct {
	client       *kubeapi.Clientset
	beforeScript string
	afterScript  string
}

func NewFromKubeconfig(config []byte, scripts ...[]string) (Kubernetes, error) {
	kubeconfig, err := kubeconfig.RESTConfigFromKubeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("fail read kubeconfig: %v", err)
	}

	client, err := kubeapi.NewForConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("fail new client: %v", err)
	}

	if len(scripts) > 0 {
		return &kubernetesImpl{
			client:       client,
			beforeScript: scripts[0][0],
			afterScript:  scripts[0][1],
		}, nil
	} else {
		return &kubernetesImpl{
			client:       client,
			beforeScript: "",
			afterScript:  "",
		}, nil
	}
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
	script string,
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
									Command: []string{"/bin/bash", "-c", "/data/exec.sh"},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: fmt.Sprintf("%s-data-cm", name),
											},
										},
									},
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

	err := self.renderConfigMapExecScript(
		fmt.Sprintf("%s-data-cm", name),
		namespace,
		script,
	)
	if err != nil {
		return err
	}

	_, err = cronjobs.Create(
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
	script string,
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
							Command: []string{"/bin/bash", "-c", "/data/exec.sh"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("%s-data-cm", name),
									},
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	err := self.renderConfigMapExecScript(
		fmt.Sprintf("%s-data-cm", name),
		namespace,
		script,
	)
	if err != nil {
		return err
	}

	_, err = jobs.Create(
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
					"%s\n"+
					"%s\n"+
					"%s\n",
				self.beforeScript,
				script,
				self.afterScript,
			),
		},
	}

	_, err := configmaps.Create(context.TODO(), &spec, metav1.CreateOptions{})
	return err
}
