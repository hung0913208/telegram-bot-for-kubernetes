package platform

type Maintenance interface {
    Expand(percent int) error
    Unlock() error
    Lock() error
}
