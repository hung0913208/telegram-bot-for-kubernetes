package platform

type Pg interface {
	Backup
	Maintenance
}

type Mongo interface {
}

type Elasticsearch interface {
}
