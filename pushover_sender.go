package notify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

	if msg.Hostname != "" {
		if msg.Title != "" {
			msg.Title = "[" + msg.Hostname + "] " + msg.Title
		} else {
			msg.Title = msg.Hostname
		}
	}

	if len(msg.Title) > 250 {
		msg.Title = msg.Title[:250]
	}

	if len(msg.Text) > 1000 {
		msg.Text = msg.Text[:1000]
	}

	res, err := http.PostForm(pushoverEndpoint, url.Values{
		"token": {p.app},
		"user": {p.user},
		"message": {msg.Text},
		"title": {msg.Title},
		"priority": {strconv.Itoa(msg.Priority)},
		"timestamp": {strconv.FormatInt(msg.Timestamp.Unix(), 10)},
	})
	if err != nil {
		return SendError{err: err, temporary: true}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var r struct{
			Errors []string `json:"errors"`
		}

		json.NewDecoder(res.Body).Decode(&r)

		return SendError{
			err: fmt.Errorf("HTTP %d: %s", res.StatusCode, strings.Join(r.Errors, "; ")),
			temporary: res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500,
		}
	}

	return nil
}

func (p *PushoverSender) String() string {
	return "Pushover"
}

const pushoverEndpoint = "https://api.pushover.net/1/messages.json"
