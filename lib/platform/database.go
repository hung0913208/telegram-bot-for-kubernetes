package platform

type Pg interface {
    Maintenance
	Backup
}

type Mongo interface {
}

type Elasticsearch interface {
}
