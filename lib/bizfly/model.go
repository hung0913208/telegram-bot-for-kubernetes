package bizfly

import (
	"time"
)

type BaseModel struct {
	ID        string    `gorm:"primaryKey" json:"id"`
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
	return "tbl_account"
}

type ClusterModel struct {
	BaseModel

	Account string `gorm:"index:idx_account_id" json:"account"`
	Name    string `gorm:"index:idx_name" json:"name"`
	Balance int    `json:"balance"`
}

func (ClusterModel) TableName() string {
	return "tbl_cluster"
}

type ServerModel struct {
	BaseModel

	Status  string `gorm:"index:idx_status" json:"status"`
	Cluster string `gorm:"index:idx_cluster_id" json:"cluster"`
	Balance int    `json:"balance"`
}
