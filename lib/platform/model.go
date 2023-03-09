package platform

import (
	"time"
)

type BaseModel struct {
	UUID      string    `gorm:"primaryKey" json:"uuid"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"create_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"update_at"`
}

type VersionModel struct {
	BaseModel

	Kind    string `gorm:"index:idx_kind" json:"kind"`
	Version string `gorm:"index:idx_version" json:"version"`
}

type BackupModel struct {
	BaseModel

	Namespace string      `gorm:"index:idx_namespace" json:"namespace"`
	Volume    string      `gorm:"index:idx_volume" json:"volume"`
	Image     string      `json:"image"`
	State     BackupState `json:"state"`
}

type PgModel struct {
	BaseModel

	Cluster string `gorm:"index:idx_cluster_id" json:"cluster"`
	Volume  string `json:"volume"`
	Label   string `json:"label"`
	Primary bool   `json:"primary"`
	Usage   int    `gorm:"index:idx_usage" json:"usage"`
}
