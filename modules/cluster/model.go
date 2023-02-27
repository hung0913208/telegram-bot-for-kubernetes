package cluster

type ClusterModel struct {
	Name          string       `gorm:"primaryKey" json:"name"`
	Provider      ProviderEnum `json:"provider"`
	Metadata      string       `json:"metadata"`
	Kubeconfig    string       `json:"kubeconfig"`
	Expire        int64        `json:"expired_time"`
	Infrascruture string       `json:"infrastructure"`
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

