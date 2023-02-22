package platform

type Database interface {
	GetPg(namespace string) ([]Pg, error)
	GetMongo(namespace string) ([]Mongo, error)
	GetElasticsearch(namespace string) ([]Elasticsearch, error)
}

type Eventuate interface {
	GetKafka(namespace string) ([]Kafka, error)
	GetRabbitMQ(namespace string) ([]RabbitMQ, error)
}

type Platform interface {
	Database

	GetCdc(namespace string) ([]Cdc, error)
	GetRedis(namespace string) ([]Redis, error)
}
