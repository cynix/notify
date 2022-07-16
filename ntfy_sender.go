package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)


type NtfySender struct {
	endpoint string
}

func NewNtfySender(topic, server string) (Sender, error) {
	if topic == "" {
		return nil, fmt.Errorf("invalid topic")
	}

	if server == "" {
		server = ntfyServer
	}

	return &NtfySender{fmt.Sprintf("https://%s/%s", server, topic)}, nil
}

func (n *NtfySender) Send(msg Message) error {
	if n.endpoint == "" {
		return SendError{err: fmt.Errorf("NtfySender not initialised")}
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

	req, err := http.NewRequest(http.MethodPost, n.endpoint, strings.NewReader(msg.Text))
	if err != nil {
		return SendError{err: err}
	}

	req.Header.Set("content-type", "text/plain")
	req.Header.Set("x-firebase", "no")

	req.Header.Set("x-title", msg.Title)
	req.Header.Set("x-priority", fmt.Sprintf("%d", msg.Priority + 3)) // [-2, 2] -> [1, 5]
	if msg.Hostname != "" {
		req.Header.Set("x-tags", msg.Hostname)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return SendError{err: err, temporary: true}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var r struct{
			Code int
			Error string
			Link string
		}

		json.NewDecoder(res.Body).Decode(&r)

		return SendError{
			err: fmt.Errorf("HTTP %d: %d: %s (%s)", res.StatusCode, r.Code, r.Error, r.Link),
			temporary: res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500,
		}
	}

	return nil
}

func (p *NtfySender) String() string {
	return "ntfy.sh"
}

const ntfyServer = "ntfy.sh"
