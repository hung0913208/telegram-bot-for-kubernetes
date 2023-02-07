package bizfly

import (
	"time"
)

type BaseModel struct {
	Id        int       `gorm:"autoIncrement" json:"id"`
	UUID      string    `gorm:"unique,index:idx_uuid" json:"uuid"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"create_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"update_at"`
}

type AccountModel struct {
	BaseModel

	Email     string `json:"email"`
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
}

func (ClusterModel) TableName() string {
	return "tbl_bizfly_cluster"
}

type ServerModel struct {
	BaseModel

	Account string `gorm:"index:idx_account_id" json:"account"`
	Status  string `gorm:"index:idx_status" json:"status"`
	Cluster string `gorm:"index:idx_cluster_id" json:"cluster"`
	Balance int    `json:"balance"`
}

func (ServerModel) TableName() string {
	return "tbl_bizfly_server"
}

type VolumeModel struct {
	BaseModel

	Account string `gorm:"index:idx_account_id" json:"account"`
	Type    string `json:"type"`
	Status  string `json:"status"`
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

	Firewall string            `gorm:"index:idx_firewall_id" json:"firewall"`
	Type     FirewallBoundEnum `gorm:"index:idx_bound_type" json:"type"`
	CIDR     string            `json:"cidr"`
}

func (FirewallBoundModel) TableName() string {
	return "tbl_bizfly_firewall_bound"
}
