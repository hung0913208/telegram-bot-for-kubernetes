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
	Cluster string `gorm:"index:idx_cluster_id" json:"cluster"`
}

func (AliasModel) TableName() string {
	return "tbl_cluster_alias"
}

type DeploymentModel struct {
	Name string `gorm:"primaryKey" json:"name"`
	Kind string `gorm:"index:idx_kind" json:"kind"`
}

func (DeploymentModel) TableName() string {
	return "tbl_cluster_deployment"
}

type PodModel struct {
	Id            int       `gorm:"primaryKey,autoIncrement" json:"id"`
	Index         int       `gorm:"index:idx_pod_index" json:"index"`
	Deployment    string    `gorm:"index:idx_deployment" json:"deployment"`
	CpuLimit      int       `json:"cpu_limit"`
	MemoryLimit   int       `json:"memory_limit"`
	CpuRequest    int       `json:"cpu_request"`
	MemoryRequest int       `json:"memory_request"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"create_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"update_at"`
}

func (PodModel) TableName() string {
	return "tbl_cluster_pod"
}

type VolumeModel struct {
	platform.BaseModel

	Pod      int    `gorm:"index:idx_pod" json:"pod"`
	Name     string `gorm:"index:idx_name" json:"name"`
	Usage    int    `json:"usage"`
	Capacity int    `json:"capacity"`
}

func (VolumeModel) TableName() string {
	return "tbl_cluster_volume"
}

type VolumeSnapshotModel struct {
	platform.BaseModel

	Volume string `gorm:"index:idx_volume" json:"volume"`
}

func (VolumeSnapshotModel) TableName() string {
	return "tbl_cluster_volume_snapshot"
}
