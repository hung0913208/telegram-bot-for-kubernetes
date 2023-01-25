package handler

import (
	"fmt"
	"net/http"
	"os"
	"time"

	sentry "github.com/getsentry/sentry-go"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/telegram"
	"github.com/hung0913208/telegram-bot-for-kubernetes/modules/cluster"
)

func init() {
	err := container.Init()
	if err != nil {
		container.Terminate("Can't setup container to store modules", 1)
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:              "",
		Debug:            true,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		container.Terminate(fmt.Sprintf("sentry.Init: %v", err), 2)
	}

	err = container.Register("cluster", cluster.NewModule())
	if err != nil {
		container.Terminate("Can't register module `cluster`", 3)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	defer sentry.Flush(2 * time.Second)
	me := telegram.NewTelegram(os.Getenv("TELEGRAM_TOKEN"))

	updateMsg, err := me.ParseIncomingRequest(r.Body)
	if err != nil {
		sentry.CaptureException(err)
	}

	err = me.ReplyMessage(updateMsg.Message.Chat.ID, "")
	if err != nil {
		sentry.CaptureException(fmt.Errorf(
			"reply message to %d fail: \n\n%v",
			updateMsg.Message.Chat.ID,
			err,
		))
	}
}