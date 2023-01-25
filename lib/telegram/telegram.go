package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Telegram interface {
	ParseIncomingRequest(reader io.Reader) (*Update, error)
	ReplyMessage(chatId int, text string) error
}

type telegramImpl struct {
	token string
}

func NewTelegram(token string) Telegram {
	return &telegramImpl{
		token: token,
	}
}

func (self *telegramImpl) ParseIncomingRequest(reader io.Reader) (*Update, error) {
	var msgUpdate Update

	err := json.NewDecoder(reader).Decode(&msgUpdate)
	if err != nil {
		return nil, err
	}

	if msgUpdate.UpdateID == 0 {
		return nil, errors.New("invalid update request message")
	}

	return &msgUpdate, nil
}

func (self *telegramImpl) ReplyMessage(chatId int, text string) error {
	replyObj := map[string]string{
		"chat_id": strconv.Itoa(chatId),
		"text":    text,
	}
	replyMsg, err := json.Marshal(replyObj)
	if err != nil {
		return err
	}

	_, err = http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", self.token),
		"application/json",
		bytes.NewBuffer(replyMsg),
	)
	return err
}
