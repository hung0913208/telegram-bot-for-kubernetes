package bizfly

import (
	"time"
)

type BaseModel struct {
    Id        int       `gorm:"autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"create_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"update_at"`
}

type AccountModel struct {
	BaseModel

    UUID      string `gorm:"unique,index:idx_uuid" json:"uuid"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	ProjectId string `json:"project_id"`
}

func (AccountModel) TableName() string {
	return "tbl_bizfly_account"
}

type ClusterModel struct {
	BaseModel

    UUID    string `gorm:"unique,index:idx_uuid" json:"uuid"`
	Account string `gorm:"index:idx_account_id" json:"account"`
	Name    string `gorm:"index:idx_name" json:"name"`
    Status  string `json: "status"`
	Balance int    `json:"balance"`
}

func (ClusterModel) TableName() string {
	return "tbl_bizfly_cluster"
}

type ServerModel struct {
	BaseModel

    UUID    string `gorm:"unique,index:idx_uuid" json:"uuid"`
	Status  string `gorm:"index:idx_status" json:"status"`
	Cluster string `gorm:"index:idx_cluster_id" json:"cluster"`
	Balance int    `json:"balance"`
}

func (ServerModel) TableName() string {
	return "tbl_bizfly_server"
}
