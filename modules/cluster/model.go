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

type NodeModel struct {
	Id      string `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"index:idx_name" json:"name"`
	Pool    string `gorm:"index:idx_pool" json:"pool_id"`
	Cluster string `gorm:"index:idx_cluster_id" json:"cluster"`
}

func (NodeModel) TableName() string {
	return "tbl_cluster_node"
}

type PodModel struct {
	Id            string `gorm:"primaryKey" json:"id"`
	Name          string `gorm:"index:idx_name" json:"name"`
	Cluster       string `gorm:"index:idx_cluster_id" json:"cluster"`
	Node          string `gorm:"index:idx_node_id" json:"node"`
	Status        string `gorm:"index:idx_status" json:"status"`
	CpuLimit      int    `json:"cpu_max"`
	MemoryLimit   int    `json:"memory_max"`
	CpuRequest    int    `json:"cpu_min"`
	MemoryRequest int    `json:"memory_min"`
}

func (PodModel) TableName() string {
	return "tbl_cluster_pod"
}

type IngressModel struct {
	Id string `gorm:"primaryKey" json:"id"`
}

func (IngressModel) TableName() string {
	return "tbl_cluster_ingress"
}

type PlatformModel struct {
}

func (PlatformModel) TableName() string {
	return "tbl_cluster_platform"
}

type AppModel struct {
}

func (AppModel) TableName() string {
	return "tbl_cluster_app"
}
