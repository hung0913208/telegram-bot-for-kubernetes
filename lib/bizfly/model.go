package bizfly

import (
    "time"
)

type BaseModel struct {
    ID        string        `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time     `gorm:"autoCreateTime" json:"create_at"`
    UpdatedAt time.Time     `gorm:"autoUpdateTime" json:"update_at"`
}

type AccountModel struct {
    BaseModel

    Email       string      `json:"email"`
    Password    string      `json:"password"`
    ProjectId   string      `json:"project_id"`
}

func (AccountModel) TableName() string {
    return "tbl_account"
}

type ClusterModel struct {
    BaseModel

    AccountId   string      `gorm:"index" json:"account"`
    Name        string      `gorm:"index" json:"name"`
}

func (ClusterModel) TableName() string {
    return "tbl_cluster"
}

type ServerModel struct {
}
