package cluster

type DatabasePlatform interface {
}

type ReplicationPlatform interface {
	DatabasePlatform
}

type CachingPlatform interface {
}

type EventPlatform interface {
}

type MonitorPlatform interface {
}
