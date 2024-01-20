package notify

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)


func Notify(msg Message) error {
	if msg.Text == "" {
		return fmt.Errorf("no text given")
	}

	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	if msg.Hostname == "" {
		msg.Hostname = Hostname
	}

	if conn, err := net.Dial("unixgram", UnixSocket); err == nil {
		defer conn.Close()

		if err := json.NewEncoder(conn).Encode(msg); err == nil {
			return nil
		}
	}

	ss, err := LoadSenders()
	if err != nil {
		return err
	}

	errs := Send(msg, ss)

	for _, err = range errs {
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to send: %v", errs)
}

var Hostname string

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	Hostname = strings.SplitN(hostname, ".", 2)[0]
}
