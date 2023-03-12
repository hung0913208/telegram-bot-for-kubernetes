package cluster

import (
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/platform"

	"time"
)

type ClusterModel struct {
	Name           string       `gorm:"primaryKey" json:"name"`
	CreatedAt      time.Time    `gorm:"autoCreateTime" json:"create_at"`
	UpdatedAt      time.Time    `gorm:"autoUpdateTime" json:"update_at"`
	Provider       ProviderEnum `json:"provider"`
	Metadata       string       `json:"metadata"`
	Kubeconfig     string       `json:"kubeconfig"`
	Expire         int64        `json:"expired_time"`
	Infrastructure string       `json:"infrastructure"`
}

func (ClusterModel) TableName() string {
	return "tbl_cluster_cluster"
}

type AliasModel struct {
	Alias   string `gorm:"primaryKey" json:"name"`
	Cluster string `gorm:"index:idx_cluster_alias_cluster" json:"cluster"`
}

func (AliasModel) TableName() string {
	return "tbl_cluster_alias"
}

type DeploymentModel struct {
	Name    string `gorm:"primaryKey" json:"name"`
	Cluster string `gorm:"primaryKey,index:idx_cluster_deployment_cluster" json:"cluster"`
	Kind    string `gorm:"index:idx_cluster_deployment_kind" json:"kind"`
}

func (DeploymentModel) TableName() string {
	return "tbl_cluster_deployment"
}

type PodModel struct {
	platform.BaseModel

	Deployment    string `gorm:"index:idx_cluster_pod_deployment" json:"deployment"`
	Status        int    `gorm:"index:idx_cluster_pod_status" json:"status"`
	Version       string `json:"version"`
	CpuLimit      int    `gorm:"index:idx_cluster_pod_cpu_limit" json:"cpu_limit"`
	MemoryLimit   int    `gorm:"index:idx_cluster_pod_memory_limit" json:"memory_limit"`
	CpuRequest    int    `gorm:"index:idx_cluster_pod_cpu_request" json:"cpu_request"`
	MemoryRequest int    `gorm:"index:idx_cluster_pod_memory_request" json:"memory_request"`
}

func (PodModel) TableName() string {
	return "tbl_cluster_pod"
}

type VolumeModel struct {
	platform.BaseModel

	Name     string `json:"name"`
	Pod      string `gorm:"index:idx_cluster_volume_pod" json:"pod"`
	Cluster  string `gorm:"index:idx_cluster_volume_cluster" json:"cluster"`
	Usage    int    `gorm:"index:idx_cluster_volume_usage" json:"usage"`
	Capacity int    `json:"capacity"`
}

func (VolumeModel) TableName() string {
	return "tbl_cluster_volume"
}

type VolumeSnapshotModel struct {
	platform.BaseModel

	Volume  string `gorm:"index:idx_volume_snapshot_volume" json:"volume"`
	Version int    `gorm:"index:idx_volume_snapshot_version" json:"version"`
}

func (VolumeSnapshotModel) TableName() string {
	return "tbl_cluster_volume_snapshot"
}
