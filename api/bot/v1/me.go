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
)

func init() {
	err := container.Init()
	if err != nil {
		container.Terminate("Can't setup container to store modules", 1)
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Debug:            true,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		container.Terminate(fmt.Sprintf("sentry.Init: %v", err), 2)
	}
	defer sentry.Flush(2 * time.Second)

	err = container.Register("cluster", cluster.NewModule())
	if err != nil {
		container.Terminate("Can't register module `cluster`", 3)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	defer sentry.Flush(2 * time.Second)

	if r.Method == "GET" {
		return
	}

	me := telegram.NewTelegram(os.Getenv("TELEGRAM_TOKEN"))
	logger := logs.NewLogger()
	updateMsg, err := me.ParseIncomingRequest(r.Body)

	if err != nil {
		logger.Errorf("Fail parsing: %v", err)
		return
	}

    command := strings.Trim(updateMsg.Message.Text, " ")
    needAnswer := false
    
    if updateMsg.Message.Chat.Type == "private" {
        needAnswer = true
    }

    if strings.Contains(command, os.Getenv("TELEGRAM_ALIAS")) {
        needAnswer = true
    }

    if needAnswer {
		err = me.ReplyMessage(updateMsg.Message.Chat.ID, "test test test")

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
