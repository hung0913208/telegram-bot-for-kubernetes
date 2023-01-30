package toolbox

import (
    "time"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/io"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/logs"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/bizfly"
)

type toolboxImpl struct {
    io        io.Io
    timeout   time.Duration
    logger    logs.Logger
    bizflyApi map[string]bizfly.Api
}

func NewToolbox(input, output string) Toolbox {
    return &toolboxImpl{
        io:        io.NewStringStream(input, output),
        logger:    logs.NewLoggerWithStacktrace(),
        bizflyApi: make(map[string]bizfly.Api),
    }
}

func (self *toolboxImpl) Init(timeout time.Duration) error {
	// @TODO: please fill this one
    self.timeout = timeout
	return nil
}

func (self *toolboxImpl) Deinit() error {
	// @TODO: please fill this one
	return nil
}

func (self *toolboxImpl) Execute(args []string) error {
    return self._tmain(args)
}

func (self *toolboxImpl) Okf(format string, args ...interface{}) {
    self.io.Printf(format, args)
}

func (self *toolboxImpl) Failf(format string, args ...interface{}) {
    self.io.Printf(format, args)

    if self.logger != nil {
        self.logger.Errorf(format, args)
    }
}
