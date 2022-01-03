package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)


type PushoverSender struct {
	app string
	user string
}

func NewPushoverSender(app, user string) (Sender, error) {
	return &PushoverSender{app, user}, nil
}

func (p *PushoverSender) Send(msg Message) error {
	if p.app == "" || p.user == "" {
		return SendError{err: fmt.Errorf("PushoverSender not initialised")}
	}

	if msg.Priority < -2 || msg.Priority > 2 {
		return SendError{err: fmt.Errorf("invalid priority: %d", msg.Priority)}
	}

	if len(msg.Title) > 250 {
		msg.Title = msg.Title[:250]
	}

	if len(msg.Text) > 1000 {
		msg.Text = msg.Text[:1000]
	}

	res, err := http.PostForm(endpoint, url.Values{
		"token": {p.app},
		"user": {p.user},
		"message": {msg.Text + "\n\n#" + Hostname},
		"title": {msg.Title},
		"priority": {strconv.Itoa(msg.Priority)},
		"timestamp": {strconv.FormatInt(msg.Timestamp.Unix(), 10)},
	})
	if err != nil {
		return SendError{err: err, temporary: true}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return SendError{
			err: fmt.Errorf("HTTP %d", res.StatusCode),
			temporary: res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500,
		}
	}

	return nil
}

func (p *PushoverSender) String() string {
	return "Pushover"
}

const endpoint = "https://api.pushover.net/1/messages.json"
