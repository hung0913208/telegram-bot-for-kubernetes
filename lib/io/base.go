package io

import (
	"io"
)

type Io interface {
	io.Writer

	Scan(format string, args ...interface{}) (int, error)
	Print(msg string) error
	Flush()
}
