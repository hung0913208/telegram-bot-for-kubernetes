package handler

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	sentry "github.com/getsentry/sentry-go"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/logs"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/telegram"
	"github.com/hung0913208/telegram-bot-for-kubernetes/modules/cluster"
	"github.com/hung0913208/telegram-bot-for-kubernetes/modules/toolbox"
)

var input string
var output string

const (
    NoError              = 0
    ErrorInitContainer   = 1
    ErrorInitSentry      = 2
    ErrorRegisterCluster = 3
    ErrorRegisterBot     = 4
)

var me telegram.Telegram

func init() {
    timeout := 10
	err := container.Init()
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

    err = container.RegisterSimpleModule(
        "toolbox", 
        toolbox.NewToolbox(input, output),
        timeout,
    )
    if err != nil {
        container.Terminate("Can't register module `toolbox`", ErrorRegisterBot)
    }

    me := telegram.NewTelegram(os.Getenv("TELEGRAM_TOKEN"))

    if len(os.Getenv("VERCEL_URL")) > 0 {
        webhook := fmt.Sprintf("https://%s/api/bot/v1/me", os.Getenv("VERCEL_URL"))

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
	defer sentry.Flush(2 * time.Second)

	if r.Method == "GET" {
	    return
	}

    logger := logs.NewLogger()
	updateMsg, err := me.ParseIncomingRequest(r.Body)

	if err != nil {
	    logger.Errorf("Fail parsing: %v", err)
	    return
	}

    input = strings.Trim(updateMsg.Message.Text, " ")
    needAnswer := false

    if updateMsg.Message.Chat.Type == "private" {
        needAnswer = true
    }

    if strings.HasPrefix(input, os.Getenv("TELEGRAM_ALIAS")) {
        needAnswer = true
    }

    if needAnswer {
        if len(output) > 0 {
            err = me.ReplyMessage(updateMsg.Message.Chat.ID, output)
        }

        if err != nil {
            logger.Errorf(
                "reply message to %d fail: \n\n%v",
                updateMsg.Message.Chat.ID,
                err,
            )
            return
        }
	}
}
