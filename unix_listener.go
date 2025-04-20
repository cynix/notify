package notify

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"time"
)


const UnixSocket = "/var/run/notify/sock"

type UnixListener struct {
	socket string
	conn net.PacketConn
	closing bool
}

func NewUnixListener() (Listener, error) {
	if err := os.Remove(UnixSocket); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err := os.MkdirAll(path.Dir(UnixSocket), 0755); err != nil {
		return nil, err
	}

	conn, err := net.ListenPacket("unixgram", UnixSocket)
	if err != nil {
		return nil, err
	}

	if err = os.Chmod(UnixSocket, 0777); err != nil {
		return nil, err
	}

	return &UnixListener{socket: UnixSocket, conn: conn}, nil
}

func (u *UnixListener) Listen(ch chan<- Message) error {
	if u.conn == nil {
		return fmt.Errorf("UnixListener not initialised")
	}

	defer os.Remove(u.socket)

	b := make([]byte, 4096)

	for {
		n, _, err := u.conn.ReadFrom(b)

		if err != nil {
			if u.closing {
				return nil
			}

			return err
		}

		if n <= 0 {
			continue
		}

		var msg Message

		if err = json.Unmarshal(b[:n], &msg); err != nil {
			log.Printf("failed to decode message on %s: %v", u.socket, err)
			log.Println(Indent(hex.Dump(b[:n])))
			continue
		}

		now := time.Now()

		if msg.Timestamp.After(now) {
			msg.Timestamp = now
		} else if now.Sub(msg.Timestamp) > 1 * time.Second {
			msg.Timestamp = now.Add(-1 * time.Second)
		}

		ch <- msg
	}
}

func (u *UnixListener) Close() error {
	u.closing = true
	return u.conn.Close()
}
