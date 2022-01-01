package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)


const UserId = 100935633
var Endpoint string

func ConfigureTelegram(token string) {
	Endpoint = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
}

func SendTelegram(text string) error {
	if Endpoint == "" {
		panic("Telegram not configured")
	}

	if len(text) > 4096 {
		text = text[:4096]
	}

	msg := struct {
		ChatId int `json:"chat_id"`
		Text string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}{
		ChatId: UserId,
		Text: text,
		ParseMode: "MarkdownV2",
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	res, err := http.Post(Endpoint, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", res.StatusCode)
	}

	return nil
}
