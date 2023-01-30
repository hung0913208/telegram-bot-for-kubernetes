package io

import (
    "fmt"
)

type ioStringStreamImpl struct {
    input  string
    output *string
}

func NewStringStream(input string, output *string) Io {
    return &ioStringStreamImpl{
        input:  input,
        output: output,
    }
}

func (self *ioStringStreamImpl) Scan(
    format string,
    args ...interface{},
) (int, error) {
    return fmt.Sscanf(self.input, format, args)    
}

func (self *ioStringStreamImpl) Print(msg string) error {
    *(self.output) += msg
    return nil
}

func (self *ioStringStreamImpl) Write(b []byte) (int, error) {
    *(self.output) += string(b)
    return len(b), nil
}
