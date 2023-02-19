package kubernetes

import (
	"context"
	"encoding/json"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Timestamp  time.Time   `json:"timestamp"`
		Window     string      `json:"window"`
		Containers []Container `json:"containers"`
	} `json:"items"`
}

type Container struct {
	Name  string `json:"name"`
	Usage Usage  `json:"usage"`
}
type Usage struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type NodeMetricsList struct {
	Pods []*PodRef `json:"pods"`
}

type PodRef struct {
	PodRef struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"podRef"`
	Volumes []*Volume `json:"volume"`
}

type Volume struct {
	// The time at which these stats were updated.
	Time metav1.Time `json:"time"`

	// Used represents the total bytes used by the Volume.
	// Note: For block devices this maybe more than the total size of the files.
	UsedBytes int64 `json:"usedBytes"` // TODO: use uint64 here as well?

	// Capacity represents the total capacity (bytes) of the volume's
	// underlying storage. For Volumes that share a filesystem with the host
	// (e.g. emptydir, hostpath) this is the size of the underlying storage,
	// and will not equal Used + Available as the fs is shared.
	CapacityBytes int64 `json:"capacityBytes"`

	// Available represents the storage space available (bytes) for the
	// Volume. For Volumes that share a filesystem with the host (e.g.
	// emptydir, hostpath), this is the available space on the underlying
	// storage, and is shared with host processes and other Volumes.
	AvailableBytes int64 `json:"availableBytes"`

	// InodesUsed represents the total inodes used by the Volume.
	InodesUsed uint64 `json:"inodesUsed"`

	// Inodes represents the total number of inodes available in the volume.
	// For volumes that share a filesystem with the host (e.g. emptydir, hostpath),
	// this is the inodes available in the underlying storage,
	// and will not equal InodesUsed + InodesFree as the fs is shared.
	Inodes uint64 `json:"inodes"`

	// InodesFree represent the inodes available for the volume.  For Volumes that share
	// a filesystem with the host (e.g. emptydir, hostpath), this is the free inodes
	// on the underlying storage, and is shared with host processes and other volumes
	InodesFree uint64 `json:"inodesFree"`

	Name   string `json:"name"`
	PvcRef struct {
		PvcName      string `json:"name"`
		PvcNamespace string `json:"namespace"`
	} `json:"pvcRef"`
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

func (self *kubernetesImpl) GetNodeMetrics(node string) (*NodeMetricsList, error) {
	var metric NodeMetricsList

	resp := self.client.CoreV1().
		RESTClient().
		Get().
		Resource("nodes").Name(node).
		SubResource("proxy").
		Suffix("stats/summary").
		Do(context.TODO())

	raw, err := resp.Raw()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(raw, &metric)
	if err != nil {
		return nil, err
	}

	return &metric, nil
}
