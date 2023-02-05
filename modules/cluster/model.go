package cluster

type ClusterModel struct {
	Id         int          `gorm:"primaryKey" json:"id"`
	Name       string       `json:"name"`
	Provider   ProviderEnum `json:"provider"`
	Metadata   string       `json:"metadata"`
	Kubeconfig string       `json:"kubeconfig"`
}

func (ClusterModel) TableName() string {
	return "tbl_cluster_cluster"
}
