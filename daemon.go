package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joncrlsn/dque"
)


type Message struct {
	Timestamp time.Time
	Text string
}

type Daemon struct {
	ctx context.Context
	conn net.PacketConn
	queue *dque.DQue
	wg sync.WaitGroup
}

func (d *Daemon) Serve() {
	if d.conn != nil {
		panic("already started")
	}

	os.Remove(Socket)

	var cancel context.CancelFunc
	d.ctx, cancel = context.WithCancel(context.Background())

	var err error

	if d.conn, err = new(net.ListenConfig).ListenPacket(d.ctx, "unixgram", Socket); err != nil {
		panic(err)
	}
	defer func() {
		d.conn.Close()
		os.Remove(Socket)
	}()

	if err = os.Chmod(Socket, 0777); err != nil {
		panic(err)
	}

	if err = os.MkdirAll(queueDir, 0700); err != nil {
		panic(err)
	}

	if d.queue, err = dque.NewOrOpen(queueName, queueDir, 32, func() interface{} { return new(Message) }); err != nil {
		panic(err)
	}
	defer d.queue.Close()

	if err = d.queue.TurboOn(); err != nil {
		panic(err)
	}

	log.Println("Daemon started")
	SendTelegram(fmt.Sprintf("Daemon started on %s", Hostname))

	d.wg.Add(2)
	go d.listen()
	go d.retry()

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	cancel()
	d.wg.Wait()

	SendTelegram(fmt.Sprintf("Daemon stopped on %s", Hostname))
	log.Println("Daemon stopped")
}

func (d *Daemon) listen() {
	defer d.wg.Done()

	b := make([]byte, 4096)

	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			break
		}

		d.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, _, err := d.conn.ReadFrom(b)

		if err != nil {
			if err.(net.Error).Timeout() {
				continue
			}

			panic(err)
		}

		if n <= 0 {
			continue
		}

		msg := Message{time.Now(), string(b[:n])}

		if err = SendTelegram(msg.Text); err != nil {
			log.Printf("failed to send, queueing for retry: %v", err)
			d.queue.Enqueue(&msg)
		}
	}
}

func (d *Daemon) retry() {
	defer d.wg.Done()

	var failures int

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			break
		}

		for i := 0; i < 30; i++ {
			v, err := d.queue.Peek()
			if err == dque.ErrEmpty {
				break
			} else if err != nil {
				panic(err)
			}

			msg := v.(*Message)

			if err = SendTelegram(msg.Text); err == nil {
				failures = 0
				if _, err = d.queue.Dequeue(); err != nil {
					panic(err)
				}
				log.Printf("dequeued message originally received at %v", msg.Timestamp)
			} else {
				log.Printf("failed to send: %v", err)

				if failures++; failures >= 100 {
					failures = 0
					log.Printf("discarding %d items after 100 failed attempts", d.queue.Size())

					for {
						v, err := d.queue.Dequeue()
						if err == dque.ErrEmpty {
							break
						} else if err != nil {
							panic(err)
						}

						msg := v.(*Message)
						log.Printf("message originally received at %v\n%s", msg.Timestamp, Indent(msg.Text))
					}
				}

				break
			}
		}
	}
}

const (
	Socket = "/var/run/notify/sock"
)

const (
	queueDir = "/var/tmp/notify"
	queueName = "queue"
)
