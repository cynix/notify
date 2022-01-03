package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)


const UserId = 100935633

type TelegramSender struct {
	endpoint string
	chatID string
}

func NewTelegramSender(token, chatID string) (Sender, error) {
	return &TelegramSender{fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token), chatID}, nil
}

func (t *TelegramSender) Send(msg Message) error {
	if t.endpoint == "" {
		return SendError{err: fmt.Errorf("TelegramSender not initialised")}
	}

	if len(msg.Title) + len(msg.Text) > 4000 {
		msg.Text = msg.Text[:4000-len(msg.Title)]
	}

	m := struct {
		ChatId string `json:"chat_id"`
		Text string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}{
		ChatId: t.chatID,
		Text: escape(msg.Text) + "\n\n",
		ParseMode: "MarkdownV2",
	}

	if msg.Title != "" {
		m.Text = "*" + escape(msg.Title) + "*\n\n" + m.Text
	}

	for _, hostname := range strings.Split(msg.Hostname, ".") {
		m.Text += "\\#" + hostname + " "
	}

	b, err := json.Marshal(m)
	if err != nil {
		return SendError{err: err}
	}

	res, err := http.Post(t.endpoint, "application/json", bytes.NewReader(b))
	if err != nil {
		s := strings.ReplaceAll(err.Error(), t.endpoint, "*")
		return SendError{err: fmt.Errorf("%s", s), temporary: true}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var r struct{
			Description string `json:"description"`
		}

		json.NewDecoder(res.Body).Decode(&r)

		return SendError{
			err: fmt.Errorf("HTTP %d: %s", res.StatusCode, r.Description),
			temporary: res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500,
		}
	}

	return nil
}

func (t *TelegramSender) String() string {
	return "Telegram"
}

var esc = regexp.MustCompile("([\\\\_*\\[\\]()~`>#+=|{}.!-])")

func escape(s string) string {
	return esc.ReplaceAllString(s, "\\$1")
}
