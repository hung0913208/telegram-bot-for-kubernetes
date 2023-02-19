package kubernetes

import (
	"context"
	"encoding/json"
	"time"
)

type PodMetricsList struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		SelfLink string `json:"selfLink"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name              string    `json:"name"`
			Namespace         string    `json:"namespace"`
			SelfLink          string    `json:"selfLink"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
		} `json:"metadata"`
		Timestamp  time.Time `json:"timestamp"`
		Window     string    `json:"window"`
		Containers []struct {
			Name  string `json:"name"`
			Usage struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"usage"`
		} `json:"containers"`
	} `json:"items"`
}

func (self *kubernetesImpl) GetPodMetrics() (*PodMetricsList, error) {
	var podMetrics PodMetricsList

	data, err := self.client.RESTClient().
		Get().
		AbsPath("apis/metrics.k8s.io/v1beta1/pods").
		DoRaw(context.TODO())
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &podMetrics)
	if err != nil {
		return nil, err
	}

	return &podMetrics, nil
}
