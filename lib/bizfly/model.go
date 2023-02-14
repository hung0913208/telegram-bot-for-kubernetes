package bizfly

import (
	"time"
)

type BaseModel struct {
	UUID      string    `gorm:"primaryKey" json:"uuid"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"create_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"update_at"`
}

type AccountModel struct {
	BaseModel

	Email     string `gorm:"index:idx_email" json:"email"`
	Password  string `json:"password"`
	ProjectId string `json:"project_id"`
}

func (AccountModel) TableName() string {
	return "tbl_bizfly_account"
}

type ClusterModel struct {
	BaseModel

	Account string `gorm:"index:idx_account_id" json:"account"`
	Name    string `gorm:"index:idx_name" json:"name"`
	Status  string `json:"status"`
	Balance int    `json:"balance"`
	Locked  bool   `gorm:"index:idx_locked" json:"locked"`
}

func (ClusterModel) TableName() string {
	return "tbl_bizfly_cluster"
}

type PoolModel struct {
	BaseModel

	Name              string `gorm:"index:idx_name" json:"name"`
	Account           string `gorm:"index:idx_account_id" json:"account"`
	Cluster           string `gorm:"index:idx_cluster_id" json:"cluster_id"`
	Zone              string `gorm:"index:idx_zone" json:"zone"`
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
	BaseModel

	Name    string `json:"name"`
	Account string `gorm:"index:idx_account_id" json:"account"`
	Server  string `gorm:"index:idx_server_id" json:"physical_id"`
	Status  string `gorm:"index:idx_status" json:"status"`
	Reason  string `json:"reason"`
}

func (PoolNodeModel) TableName() string {
	return "tbl_bizfly_pool_node"
}

type ServerModel struct {
	BaseModel

	Account string `gorm:"index:idx_account_id" json:"account"`
	Status  string `gorm:"index:idx_status" json:"status"`
	Cluster string `gorm:"index:idx_cluster_id" json:"cluster"`
	Balance int    `json:"balance"`
	Locked  bool   `gorm:"index:idx_locked" json:"locked"`
	Zone    string `gorm:"index:idx_zone" json:"zone"`
}

func (ServerModel) TableName() string {
	return "tbl_bizfly_server"
}

type VolumeModel struct {
	BaseModel

	Account string `gorm:"index:idx_account_id" json:"account"`
	Type    string `json:"type"`
	Server  string `gorm:"index:idx_server_id" json:"server_id"`
	Status  string `json:"status"`
	Zone    string `gorm:"index:idx_zone" json:"zone"`
}

func (VolumeModel) TableName() string {
	return "tbl_bizfly_volume"
}

type FirewallModel struct {
	BaseModel

	Account string `gorm:"index:idx_account_id" json:"account"`
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
	BaseModel

	Account  string            `gorm:"index:idx_account_id" json:"account"`
	Firewall string            `gorm:"index:idx_firewall_id" json:"firewall"`
	Type     FirewallBoundEnum `gorm:"index:idx_bound_type" json:"type"`
	CIDR     string            `json:"cidr"`
}

func (FirewallBoundModel) TableName() string {
	return "tbl_bizfly_firewall_bound"
}
