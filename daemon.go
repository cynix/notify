package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joncrlsn/dque"
)


func Daemon(ll []Listener, ss []Sender) {
	if (len(ll) == 0 || len(ss) == 0) {
		panic("need at least 1 listener and 1 sender")
	}

	var err error

	if err = os.MkdirAll(queueDir, 0700); err != nil {
		panic(err)
	}

	var q *dque.DQue

	if q, err = dque.NewOrOpen(queueName, queueDir, 32, func() interface{} { return new(Message) }); err != nil {
		panic(err)
	}
	defer q.Close()

	if err = q.TurboOn(); err != nil {
		panic(err)
	}

	log.Printf("daemon started (l=%d s=%d)", len(ll), len(ss))
	Send(Message{Timestamp: time.Now(), Hostname: Hostname, Text: "Daemon started", Priority: -1}, ss)

	ch := make(chan Message, 100)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1 + len(ll))

	go retry(ctx, &wg, q, ss)

	for _, l := range ll {
		go func(l Listener) {
			defer wg.Done()

			if err := l.Listen(ch); err != nil {
				log.Printf("listener error: %v", err)
			}
		}(l)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

_stop:
	for {
		select {
		case <-sig:
			break _stop

		case msg := <-ch:
			var ok, tmp bool

			if msg.Hostname != Hostname {
				msg.Hostname = msg.Hostname + "." + Hostname
			}

			for _, err := range Send(msg, ss) {
				if err == nil {
					ok = true
					continue
				}

				log.Printf("failed to send message: %s: %v", msg.Title, err)
				tmp = tmp || IsTemporary(err)
			}

			if ok {
				log.Printf("sent message: %v", msg.Title)
				continue
			}

			if !tmp {
				log.Printf("discarding message due to permanent failure: %s", msg.Title)
				log.Print(Indent(msg.Text))
				continue
			}

			log.Printf("enqueueing message: %v", msg.Title)
			q.Enqueue(msg)
		}
	}

	log.Print("daemon stopping")
	Send(Message{Timestamp: time.Now(), Hostname: Hostname, Text: "Daemon stopping", Priority: -1}, ss)

	for _, l := range ll {
		if err := l.Close(); err != nil {
			log.Printf("listener error during close: %v", err)
		}
	}

	cancel()
	wg.Wait()
}

func retry(ctx context.Context, wg *sync.WaitGroup, q *dque.DQue, ss []Sender) {
	defer wg.Done()

	var failures int

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			break
		}

		for i := 0; i < 30; i++ {
			v, err := q.Peek()
			if err == dque.ErrEmpty {
				break
			} else if err != nil {
				panic(err)
			}

			msg := v.(*Message)
			var ok, tmp bool

			for _, err := range Send(*msg, ss) {
				if err == nil {
					ok = true
					continue
				}

				log.Printf("failed to send queued message (originally received at %v): %s: %v", msg.Timestamp, msg.Title, err)
				tmp = tmp || IsTemporary(err)
			}

			if !ok && tmp {
				if failures++; failures >= 100 {
					tmp = false
				}
			}

			if ok || !tmp {
				if ok {
					log.Printf("dequeued message (originally received at %v)", msg.Timestamp)
				} else {
					log.Printf("discarding queued message due to permanent failure: %s", msg.Title)
					log.Print(Indent(msg.Text))
				}

				if _, err = q.Dequeue(); err != nil {
					panic(err)
				}

				failures = 0
				continue
			}

			break
		}
	}
}

const (
	queueDir = "/var/tmp/notify"
	queueName = "queue"
)
