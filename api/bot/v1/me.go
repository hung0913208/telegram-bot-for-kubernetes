package handler

import (
    "encoding/json"
    "runtime/debug"
	"fmt"
	"net/http"
	"os"
    "strconv"
	"strings"
	"time"

	sentry "github.com/getsentry/sentry-go"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
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

var me telegram.Telegram

func init() {
    timeout, err := strconv.Atoi(os.Getenv("TIMEOUT"))
    if err != nil {
        timeout = 200
    }

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

	err = container.RegisterSimpleModule("cluster", cluster.NewModule(), timeout)
	if err != nil {
	    container.Terminate("Can't register module `cluster`", ErrorRegisterCluster)
	}

    // @TODO: implement database module
    //err = container.RegisterSimpleModule(
    //    "elephansql", 
    //    db.NewModule(), 
    //    timeout,
    //)
    //if err != nil {
    //    container.Terminate("Can't register module `elephansql`", ErrorRegisterSql)
    //}

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

    timeout, err := strconv.Atoi(os.Getenv("TIMEOUT"))
    if err != nil {
        timeout = 200
    }

	if r.Method == "GET" {
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
    input := strings.Trim(msg.Text, " ")
    outputs := make([]string, 0)
    session := toolbox.NewToolbox(input, &outputs)

    session.Init(time.Duration(timeout) * time.Millisecond)

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
        err = session.Execute(strings.Split(input, " "))
        if err != nil && len(outputs) == 0 {
            outputs = append(outputs, fmt.Sprintf("%v", err))
        }

        if len(outputs) == 0 {
            return
        }

        for _, output := range outputs {
            err = me.ReplyMessage(msg.Chat.ID, output)

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
                        output,
                    ),
                )
                return
            }
        }
	}
}
