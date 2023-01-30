package toolbox

import (
    "context"
    "sync"
    "errors"
    "time"
    "strings"
    "strconv"
    "os"

    "github.com/spf13/cobra"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/io"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/logs"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/bizfly"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/search"
)

type toolboxImpl struct {
    io        io.Io
    wg        *sync.WaitGroup
    timeout   time.Duration
    logger    logs.Logger
    bizflyApi map[string]bizfly.Api
}

func NewToolbox(input string, output *string) Toolbox {
    return &toolboxImpl{
        io:        io.NewStringStream(input, output),
        wg:        &sync.WaitGroup{},
        logger:    logs.NewLoggerWithStacktrace(),
        bizflyApi: make(map[string]bizfly.Api),
    }
}

func (self *toolboxImpl) Init(timeout time.Duration) error {
    self.timeout = timeout
	return nil
}

func (self *toolboxImpl) Deinit() error {
	// @TODO: please fill this one
	return nil
}

func (self *toolboxImpl) Execute(args []string) error {
    waitResponse := make(chan struct{})
    ctx, cancel := context.WithTimeout(
        context.Background(), 
        self.timeout,
    )

    defer cancel()

    parser := self.newRootParser()

    if args[0] == "sre" {
        args = args[1:]
    }

    parser.SetArgs(args)
    parser.SetErr(self.io)
    parser.SetOut(self.io)

    cmd, _, err := parser.Find(args)

    if err != nil || cmd == nil {
        records, err := search.Search(nil, strings.Join(args, " "))
        if err != nil {
            return err
        }

        limit := -1
        if len(os.Getenv("SEARCH_LIMIT")) > 0 {
            limit, err = strconv.Atoi(os.Getenv("SEARCH_LIMIT"))

            if err != nil {
                limit = -1
            }
        }

        for i, record := range records {
            if limit > 0 && limit <= i {
                break
            }

            self.Ok(record.Description)
        }

        return nil
    } else {
        err := parser.ExecuteContext(ctx)
        if err != nil {
            return err
        }

        go func() {
            defer close(waitResponse)
            self.wg.Wait()
        }()

        select {
        case <-waitResponse:
            return nil

        case <-time.After(self.timeout):
            return errors.New("Timeout waiting response from sub-command")
        }
    }
}

func (self *toolboxImpl) GenerateSafeCallback(
    callback func(cmd *cobra.Command, args []string),
) func(cmd *cobra.Command, args []string) {
    return func(cmd *cobra.Command, args []string) {
        defer self.wg.Done()
        self.wg.Add(1)

        callback(cmd, args)
    }
}

func (self *toolboxImpl) Ok(msg string) {
    self.io.Print(msg + "\n")
}

func (self *toolboxImpl) Fail(msg string) {
    self.io.Print(msg + "\n")

    if self.logger != nil {
        self.logger.Error(msg)
    }
}
