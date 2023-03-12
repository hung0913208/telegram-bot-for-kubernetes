package bizfly

import (
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/platform"
)

type AccountModel struct {
	platform.BaseModel

	Email     string `gorm:"index:tbl_bizfly_account_idx_email" json:"email"`
	Password  string `json:"password"`
	ProjectId string `json:"project_id"`
}

func (AccountModel) TableName() string {
	return "tbl_bizfly_account"
}

type ClusterModel struct {
	platform.BaseModel

	Account string `gorm:"index:idx_bizfly_cluster_account_id" json:"account"`
	Name    string `gorm:"index:tbl_bizfly_cluster_idx_name" json:"name"`
	Status  string `json:"status"`
	Balance int    `json:"balance"`
	Locked  bool   `gorm:"index:idx_bizfly_cluster_locked" json:"locked"`
	Tags    string `json:"tags"`
}

func (ClusterModel) TableName() string {
	return "tbl_bizfly_cluster"
}

type ClusterStatModel struct {
	Cluster string `gorm:"primaryKey" json:"cluster"`
	Account string `gorm:"index:tbl_bizfly_cluster_stat_idx_account_id" json:"account"`
	Core    int    `json:"core"`
	Memory  int    `json:"memory"`
}

func (ClusterStatModel) TableName() string {
	return "tbl_bizfly_cluster_stat"
}

type PoolModel struct {
	platform.BaseModel

	Name              string `gorm:"index:tbl_bizfly_pool_idx_name" json:"name"`
	Account           string `gorm:"index:idx_bizfly_pool_account_id" json:"account"`
	Cluster           string `gorm:"index:idx_bizfly_pool_cluster_id" json:"cluster_id"`
	Zone              string `gorm:"index:idx_bizfly_pool_zone" json:"zone"`
	Status            string `json:"status"`
	Autoscale         string `json:"autoscale_group_id"`
	EnableAutoscaling bool   `json:"scaleable"`
	RequiredSize      int    `json:"required"`
	LimitSize         int    `json:"limit"`
}

func (PoolModel) TableName() string {
	return "tbl_bizfly_pool"
}

type PoolNodeModel struct {
	platform.BaseModel

	Name    string `json:"name"`
	Account string `gorm:"index:tbl_bizfly_pool_node_idx_account_id" json:"account"`
	Pool    string `gorm:"index:tbl_bizfly_pool_node_idx_pool_id" json:"pool"`
	Cluster string `gorm:"index:tbl_bizfly_pool_node_idx_cluster_id" json:"cluster"`
	Server  string `gorm:"index:tbl_bizfly_pool_node_idx_server_id" json:"physical_id"`
	Status  string `gorm:"index:tbl_bizfly_pool_node_idx_status" json:"status"`
	Reason  string `json:"reason"`
}

func (PoolNodeModel) TableName() string {
	return "tbl_bizfly_pool_node"
}

type ServerModel struct {
	platform.BaseModel

	Account string `gorm:"index:tbl_bizfly_server_idx_account_id" json:"account"`
	Status  string `gorm:"index:tbl_bizfly_server_idx_status" json:"status"`
	Cluster string `gorm:"index:tbl_bizfly_server_idx_cluster_id" json:"cluster"`
	Balance int    `json:"balance"`
	Locked  bool   `gorm:"index:tbl_bizfly_server_idx_locked" json:"locked"`
	Zone    string `gorm:"index:tbl_bizfly_server_idx_zone" json:"zone"`
}

func (ServerModel) TableName() string {
	return "tbl_bizfly_server"
}

type VolumeModel struct {
	platform.BaseModel

	Account     string `gorm:"index:tbl_bizfly_volume_idx_account_id" json:"account"`
	Zone        string `gorm:"index:tbl_bizfly_volume_idx_zone" json:"zone"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Size        int    `json:"size"`
}

func (VolumeModel) TableName() string {
	return "tbl_bizfly_volume"
}

type VolumeServerModel struct {
	Volume  string `gorm:"primaryKey" json:"volume_id"`
	Account string `gorm:"index:tbl_bizfly_volume_server_idx_account_id" json:"account"`
	Server  string `gorm:"index:tbl_bizfly_volume_server_idx_server_id" json:"server_id"`
}

func (VolumeServerModel) TableName() string {
	return "tbl_bizfly_volume_server"
}

type VolumeClusterModel struct {
	Pod        string `gorm:"primaryKey,index:tbl_bizfly_volume_cluster_idx_pod" json:"pod"`
	Cluster    string `gorm:"primaryKey,index:tbl_bizfly_volume_cluster_idx_cluser" json:"cluster"`
	Deployment string `gorm:"index:tbl_bizfly_volume_cluster_idx_deployment" json:"deployment"`
	Index      int    `gorm:"index:tbl_bizfly_volume_cluster_idx_index" json:"index"`
	Volume     string `gorm:"index:tbl_bizfly_volume_cluster_idx_volume_id" json:"volume_id"`
	Account    string `gorm:"index:tbl_bizfly_volume_cluster_idx_account_id" json:"account"`
	Size       int    `json:"size"`
}

func (VolumeClusterModel) TableName() string {
	return "tbl_bizfly_volume_cluster"
}

type FirewallModel struct {
	platform.BaseModel

	Account string `gorm:"index:tbl_bizfly_firewall_idx_account_id" json:"account"`
}

func (FirewallModel) TableName() string {
	return "tbl_bizfly_firewall"
}

type FirewallBoundEnum int

const (
	InBound FirewallBoundEnum = iota
	OutBound
)

type FirewallBoundModel struct {
	platform.BaseModel

	Account  string            `gorm:"index:tbl_bizfly_firewall_bound_idx_account_id" json:"account"`
	Firewall string            `gorm:"index:tbl_bizfly_firewall_bound_idx_firewall_id" json:"firewall"`
	Type     FirewallBoundEnum `gorm:"index:tbl_bizfly_firewall_bound_idx_bound_type" json:"type"`
	CIDR     string            `json:"cidr"`
}

func (FirewallBoundModel) TableName() string {
	return "tbl_bizfly_firewall_bound"
}

type SnapshotModel struct {
	platform.BackupModel
}

func (SnapshotModel) TableName() string {
	return "tbl_bizfly_snapshot"
}
