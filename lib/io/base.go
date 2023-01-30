package io

type Io interface {
    Scanf(format string, args ...interface{}) (int, error)
    Printf(format string, args ...interface{}) error
}

