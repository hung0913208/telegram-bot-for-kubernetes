package kubernetes

import (
	"context"
	"errors"
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
	GetClient() *kubeapi.Clientset
	GetPods(namespace string) (*corev1.PodList, error)
	GetHelmPods(namespace string) (*corev1.PodList, error)

	GetNodes() (*corev1.NodeList, error)
	GetPVs() (*corev1.PersistentVolumeList, error)

	GetPodMetrics() (*PodMetricsList, error)
	GetNodeMetrics(node string) (*NodeMetricsList, error)

	SetReadable(flag bool) error
	SetWriteable(flag bool) error

	Ping() bool
	Cron(
		name string,
		command string,
		schedule, namespace string,
		extendVolumes ...[]corev1.PersistentVolumeClaim,
	) error
	Do(
		name string,
		command string,
		namespace string,
		extendVolumes ...[]corev1.PersistentVolumeClaim,
	) error
}

type Hook struct {
	Header     string
	Exec       []string
	Name       string
	Image      string
	MountPoint string
	PreHook    string
	PostHook   string
}

type kubernetesImpl struct {
	client    *kubeapi.Clientset
	hook      Hook
	writeable bool
	readable  bool
}

func NewFromKubeconfig(config []byte, hook ...Hook) (Kubernetes, error) {
	kubeconfig, err := kubeconfig.RESTConfigFromKubeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("fail read kubeconfig: %v", err)
	}

	client, err := kubeapi.NewForConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("fail new client: %v", err)
	}

	if len(hook) > 0 {
		return &kubernetesImpl{
			client:    client,
			hook:      hook[0],
			readable:  true,
			writeable: false,
		}, nil
	} else {
		return &kubernetesImpl{
			client:    client,
			readable:  false,
			writeable: false,
			hook: Hook{
				Name:       "exec",
				Exec:       []string{"/bin/bash", "-c", "/data/exec"},
				MountPoint: "/data",
			},
		}, nil
	}
}

func NewFromClient(client Kubernetes, hook ...Hook) Kubernetes {
	if len(hook) > 0 {
		return &kubernetesImpl{
			client: client.GetClient(),
			hook:   hook[0],
		}
	} else {
		return &kubernetesImpl{
			client:    client.GetClient(),
			readable:  false,
			writeable: false,
			hook: Hook{
				Name:       "exec",
				Exec:       []string{"/bin/bash", "-c", "/data/exec"},
				MountPoint: "/data",
			},
		}
	}
}

func (self *kubernetesImpl) SetReadable(flag bool) error {
	if flag && !self.readable {
		self.readable = true
	}
	if flag == self.readable {
		return nil
	}

	return errors.New("can't disable immutable flag `readable` " +
		"when it has been turn on")
}

func (self *kubernetesImpl) SetWriteable(flag bool) error {
	if flag && !self.readable {
		self.readable = true
	}
	if flag == self.readable {
		return nil
	}

	return errors.New("can't disable immutable flag `writeable` " +
		"when it has been turn on")
}

func (self *kubernetesImpl) GetClient() *kubeapi.Clientset {
	return self.client
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
	command string,
	schedule, namespace string,
	extendVolumes ...[]corev1.PersistentVolumeClaim,
) error {
	cronjobs := self.client.BatchV1().CronJobs(namespace)
	volumes := make([]corev1.Volume, 0)

	if len(extendVolumes) > 0 {
		for _, pvc := range extendVolumes[0] {
			volumes = append(volumes,
				corev1.Volume{
					Name: pvc.Name,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc.Name,
						},
					},
				},
			)
		}
	}

	volumes = append(volumes,
		corev1.Volume{
			Name: self.hook.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-%s-cm", name, self.hook.Name),
					},
				},
			},
		},
	)

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
									Image:   self.hook.Image,
									Command: self.hook.Exec,
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      self.hook.Name,
											MountPath: self.hook.MountPoint,
										},
									},
								},
							},
							Volumes:       volumes,
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
			Schedule:          schedule,
		},
	}

	err := self.renderConfigMapExecHook(
		fmt.Sprintf("%s-%s-cm", name, self.hook.Name),
		namespace,
		command,
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
	command string,
	namespace string,
	extendVolumes ...[]corev1.PersistentVolumeClaim,
) error {
	volumes := make([]corev1.Volume, 0)

	if len(extendVolumes) > 0 {
		for _, pvc := range extendVolumes[0] {
			volumes = append(volumes,
				corev1.Volume{
					Name: pvc.Name,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc.Name,
						},
					},
				},
			)
		}
	}

	volumes = append(volumes,
		corev1.Volume{
			Name: self.hook.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-%s-cm", name, self.hook.Name),
					},
				},
			},
		},
	)

	backOffLimit := int32(0)
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
							Image:   self.hook.Image,
							Command: self.hook.Exec,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      self.hook.Name,
									MountPath: self.hook.MountPoint,
								},
							},
						},
					},
					Volumes:       volumes,
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	err := self.renderConfigMapExecHook(
		fmt.Sprintf("%s-%s-cm", name, self.hook.Name),
		namespace,
		command,
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

func (self *kubernetesImpl) renderConfigMapExecHook(
	name string,
	namespace string,
	command string,
) error {
	configmaps := self.client.CoreV1().
		ConfigMaps(namespace)
	spec := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{
			self.hook.Name: fmt.Sprintf(
				"%s\n%s\n%s\n%s\n",
				self.hook.Header,
				self.hook.PreHook,
				command,
				self.hook.PostHook,
			),
		},
	}

	_, err := configmaps.Create(context.TODO(), &spec, metav1.CreateOptions{})
	return err
}
