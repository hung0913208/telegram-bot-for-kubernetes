package toolbox

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/bizfly"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/io"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/logs"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/search"
	"github.com/spf13/cobra"
)

type toolboxImpl struct {
	io        io.Io
	timeout   time.Duration
	wg        *sync.WaitGroup
	logger    logs.Logger
	enable    bool
	bizflyApi map[string][]bizfly.Api
}

func NewToolbox(input string, outputs *[]string) Toolbox {
	logger := logs.NewLoggerWithStacktrace()

	return &toolboxImpl{
		io:        io.NewStringStream(input, outputs),
		wg:        &sync.WaitGroup{},
		bizflyApi: make(map[string][]bizfly.Api),
		logger:    logger,
	}
}

func (self *toolboxImpl) Init(timeout time.Duration) error {
	var setting SettingModel

	self.timeout = timeout
	self.enable = false

	if len(self.bizflyApi) == 0 {
		clients, err := bizfly.NewApiFromDatabase(
			"https://manage.bizflycloud.vn",
			"HN",
			self.timeout,
		)
		if err != nil {
			self.logger.Errorf("load bizfly account fail: %v", err)
		}

		if clients != nil {
			for _, client := range clients {
				if _, ok := self.bizflyApi[client.GetAccount()]; !ok {
					self.bizflyApi[client.GetAccount()] = make([]bizfly.Api, 0)
				}

				self.bizflyApi[client.GetAccount()] = append(
					self.bizflyApi[client.GetAccount()],
					client,
				)
			}
		}
	}

	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	migrator := dbConn.Migrator()
	migrator.CreateTable(&SettingModel{})

	result := dbConn.Where("name = ?", "timeout").
		First(&setting)

	if result.Error == nil {
		timeoutFromDb, err := strconv.Atoi(setting.Value)
		if err != nil {
			self.logger.Errorf(
				"convert timeout value fail: %v\n\nValue = %v (type %d)",
				err,
				setting.Value,
				setting.Type,
			)
		} else {
			self.timeout = time.Duration(timeoutFromDb) * time.Millisecond
		}
	}

	result = dbConn.Where("name = ?", "enable").
		First(&setting)
	if result.Error == nil {
		enableFromDb, err := strconv.Atoi(setting.Value)

		if err != nil {
			self.logger.Errorf(
				"convert enable value fail: %v\n\nValue = %v (type %d)",
				err,
				setting.Value,
				setting.Type,
			)
		} else if enableFromDb == 1 {
			self.enable = true
		} else {
			self.enable = false
		}
	}

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

	if !self.enable && args[len(args)-1] != "--help" && args[0] != "setting" {
		return nil
	}

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
	name string,
	callback func(cmd *cobra.Command, args []string),
) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		defer self.wg.Done()
		self.wg.Add(1)

		callback(cmd, args)
	}
}

func (self *toolboxImpl) Ok(msg string, data ...interface{}) {
	self.io.Print(msg+"\n", data...)
}

func (self *toolboxImpl) Fail(msg string, data ...interface{}) {
	self.io.Print(msg+"\n", data...)

	if self.logger != nil {
		self.logger.Error(fmt.Sprintf(msg, data...))
	}
}

func (self *toolboxImpl) Flush() {
	self.io.Flush()
}
