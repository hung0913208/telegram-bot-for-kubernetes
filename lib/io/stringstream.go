package io

import (
    "fmt"
)

type ioStringStreamImpl struct {
    input, output string
}

func NewStringStream(input string, output string) Io {
    return &ioStringStreamImpl{
        input:  input,
        output: output,
    }
}

func (self *ioStringStreamImpl) Scanf(
    format string,
    args ...interface{},
) (int, error) {
    return fmt.Sscanf(self.input, format, args)    
}

func (self *ioStringStreamImpl) Printf(
    format string,
    args ...interface{},
) error {
    self.output += fmt.Sprintf(format, args)
    return nil
}
