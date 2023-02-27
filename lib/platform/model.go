package platform

type VersionModel struct {
    UUID    string `gorm:"primaryKey" json:"uuid"`
    Kind    string `gorm:"index:idx_kind" json:"kind"`
    Version string `gorm:"index:idx_version" json:"version"`
}

type PgModel struct {
    UUID    string  `gorm:"primaryKey" json:"uuid"`
    Cluster string  `gorm:"index:idx_cluster_id" json:"cluster"`
    Volume  string  `json:"volume"`
    Label   string  `json:"label"`
    Primary bool    `json:"primary"`
    Usage   int     `gorm:"index:idx_usage" json:"usage"`
}

