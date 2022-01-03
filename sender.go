package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)


type Sender interface {
	Send(Message) error
}

type SendError struct {
	err error
	temporary bool
}


func LoadSenders() ([]Sender, error) {
	if ss, err := loadSenders("/usr/local/etc/notify.conf"); err == nil {
		return ss, nil
	}

	if dir, err := os.UserConfigDir(); err == nil {
		if ss, err := loadSenders(path.Join(dir, "notify.conf")); err == nil {
			return ss, nil
		}
	}

	return nil, fmt.Errorf("failed to load system/user config")
}

func Send(msg Message, ss []Sender) []error {
	errs := make([]error, 0, len(ss))
	ch := make(chan error)

	for _, s := range ss {
		go func(s Sender) {
			ch <- s.Send(msg)
		}(s)
	}

	for err := range ch {
		if errs = append(errs, err); len(errs) == len(ss) {
			break
		}
	}

	return errs
}

func (e SendError) Unwrap() error {
	return e.err
}

func (e SendError) Error() string {
	return e.err.Error()
}

func (e SendError) Temporary() bool {
	return e.temporary
}

func loadSenders(file string) ([]Sender, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	var ss []Sender

	for scanner := bufio.NewScanner(f); scanner.Scan(); {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		if len(parts) != 2 {
			continue
		}

		switch parts[0] {
		case "pushover":
			parts = strings.Split(parts[1], ",")
			if len(parts) != 2 {
				continue
			}

			s, err := NewPushoverSender(parts[0], parts[1])
			if err != nil {
				continue
			}

			ss = append(ss, s)

		case "telegram":
			parts = strings.Split(parts[1], ",")
			if len(parts) != 2 {
				continue
			}

			s, err := NewTelegramSender(parts[0], parts[1])
			if err != nil {
				continue
			}

			ss = append(ss, s)
		}
	}

	if len(ss) == 0 {
		return nil, fmt.Errorf("no valid senders")
	}

	return ss, nil
}
