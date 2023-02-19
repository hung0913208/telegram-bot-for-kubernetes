package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	sentry "github.com/getsentry/sentry-go"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/logs"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/telegram"
	"github.com/hung0913208/telegram-bot-for-kubernetes/modules/cluster"
	"github.com/hung0913208/telegram-bot-for-kubernetes/modules/toolbox"
)

const (
	NoError              = 0
	ErrorInitContainer   = 1
	ErrorInitSentry      = 2
	ErrorRegisterCluster = 3
	ErrorRegisterBot     = 4
	ErrorRegisterSql     = 5
)

var (
	me      telegram.Telegram
	input   string
	outputs []string
)

func init() {
    timeouts := []string{"100","2","1000"}

    if len(os.Getenv("TIMEOUT")) > 0 {
        timeouts = strings.Split(os.Getenv("TIMEOUT"), ",")
    }

    timeoutDb, err := strconv.Atoi(timeouts[0])
    if err != nil {
        timeoutDb = 100
    }

    timeoutModule, err := strconv.Atoi(timeouts[0])
    if err != nil {
        timeoutModule = 1000
    }

	outputs = make([]string, 0)

	err = container.Init()
	if err != nil {
		container.Terminate(
			"Can't setup container to store modules",
			ErrorInitContainer,
		)
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Debug:            true,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		container.Terminate(fmt.Sprintf("sentry.Init: %v", err), ErrorInitSentry)
	}
	defer sentry.Flush(2 * time.Second)

	port, err := strconv.Atoi(os.Getenv("ELEPHANSQL_PORT"))
	if err != nil {
		container.Terminate("Can't register module `elephansql`", ErrorRegisterSql)
	}

	elephansql, err := db.NewPgModule(
		os.Getenv("ELEPHANSQL_HOST"),
		port,
		os.Getenv("ELEPHANSQL_USERNAME"),
		os.Getenv("ELEPHANSQL_PASSWORD"),
		os.Getenv("ELEPHANSQL_DATABASE"),
	)
	if err != nil {
		container.Terminate("Can't register module `elephansql`", ErrorRegisterSql)
	}

	err = container.RegisterSimpleModule(
		"elephansql",
		elephansql,
        timeoutDb,
	)
	if err != nil {
		container.Terminate("Can't register module `elephansql`", ErrorRegisterSql)
	}

	clusterModule, err := cluster.NewModule()
	if err != nil {
		container.Terminate(fmt.Sprintf("new cluster fail: %v", err), ErrorInitContainer)
	}

	err = container.RegisterSimpleModule(
        "cluster", 
        clusterModule, 
        timeoutModule,
    )
	if err != nil {
		container.Terminate(
			fmt.Sprintf("Can't register module `cluster`: %v", err),
			ErrorRegisterCluster,
		)
	}

	err = container.RegisterSimpleModule(
		"toolbox",
		toolbox.NewToolbox(input, &outputs),
        timeoutModule,
	)
	if err != nil {
		container.Terminate(fmt.Sprintf("Can't register module `toolbox`: %v", err),
			ErrorRegisterBot)
	}

	me = telegram.NewTelegram(os.Getenv("TELEGRAM_TOKEN"))

	if len(os.Getenv("TELEGRAM_WEBHOOK")) > 0 {
		webhook := fmt.Sprintf("https://%s/api/bot/v1/me", os.Getenv("TELEGRAM_WEBHOOK"))

		webhookRegistered, err := me.GetWebhook()
		if err != nil {
			container.Terminate(fmt.Sprint(err), ErrorRegisterBot)
		}

		if webhookRegistered != webhook {
			me.SetWebhook(webhook)
		}
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	var msg *telegram.Message

	if r.Method == "GET" {
		return
	}

	bot, err := container.Pick("toolbox")
	if err != nil {
		return
	}

	logger := logs.NewLogger()
	updateMsg, err := me.ParseIncomingRequest(r.Body)

	defer func() {
		err := recover()

		if err != nil {
			if updateMsg != nil {
				out, _ := json.Marshal(updateMsg)

				if os.Getenv("VERCEL_ENV") != "production" || os.Getenv("DEBUG") == "true" {
					logger.Warn(fmt.Sprintf(
						"Crash %v when execute msg: \n%s\n\nStacktrace:\n%s",
						err,
						string(out),
						string(debug.Stack()),
					))
				}
			}

			if msg != nil {
				me.ReplyMessage(
					msg.Chat.ID,
					fmt.Sprintf("Crash with reason: %v", err),
				)
			}
		}

		sentry.Flush(2 * time.Second)
	}()

	if err != nil {
		logger.Error(fmt.Sprintf("Fail parsing: %v", err))
		return
	}

	if msg == nil && updateMsg.Message != nil {
		msg = updateMsg.Message
	}
	if msg == nil && updateMsg.EditedMessage != nil {
		msg = updateMsg.EditedMessage
	}
	if msg == nil && updateMsg.ChannelPost != nil {
		msg = updateMsg.ChannelPost
	}
	if msg == nil && updateMsg.EditedChannelPost != nil {
		msg = updateMsg.EditedChannelPost
	}

	needAnswer := false
	input = strings.Trim(msg.Text, " ")
	outputs = make([]string, 0)

	if msg.Chat.Type == "private" {
		needAnswer = true
	}

	if strings.HasPrefix(input, os.Getenv("TELEGRAM_ALIAS")) {
		needAnswer = true
		input = strings.Trim(
			strings.TrimPrefix(input, os.Getenv("TELEGRAM_ALIAS")),
			" ",
		)
	}

	if needAnswer {
		err = bot.Execute(strings.Split(input, " "))
		if err != nil && len(outputs) == 0 {
			outputs = append(outputs, fmt.Sprintf("%v", err))
		}

		if len(outputs) == 0 {
			return
		}

		for _, output := range outputs {
			rendered := ""

			for _, line := range strings.Split(output, "\n") {
				firstChar := true
				hashCnt := 0
				starCnt := 3
				fixed := make([]byte, 0)

				for i := 0; i < len(line); i++ {
					if firstChar {
						if line[i] == '#' {
							hashCnt++
							continue
						} else if hashCnt > 0 || line[i] != ' ' {
							firstChar = false

							if hashCnt > 2 {
								starCnt = 2
							} else if hashCnt == 0 {
								starCnt = 0
							}

							for j := starCnt; j > 0; j-- {
								fixed = append(fixed, '*')
							}
						}
					}

					if !firstChar {
						for _, c := range []byte("#([])-.=") {
							if c == line[i] {
								fixed = append(fixed, '\\')
							}
						}
					}

					fixed = append(fixed, line[i])
				}

				if len(fixed) > 0 {
					for j := starCnt; j > 0; j-- {
						fixed = append(fixed, '*')
					}
				}

				rendered += (string(fixed) + "\n")
			}

			err = me.ReplyMessage(
				msg.Chat.ID,
				rendered,
			)

			if err != nil {
				if len(output) > 0 {
					me.ReplyMessage(
						msg.Chat.ID,
						fmt.Sprintf("Failt reply message size %d", len(output)),
					)
				}

				logger.Error(
					fmt.Sprintf(
						"reply message to %d fail: \n\n%v\n\nOutput:\n%s",
						updateMsg.Message.Chat.ID,
						err,
						rendered,
					),
				)
				return
			}
		}
	}
}
