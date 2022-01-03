package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/rfizzle/parsemail"
	"github.com/spf13/pflag"
)


func main() {
	sendmail := strings.Contains(os.Args[0], "sendmail")

	flags := pflag.NewFlagSet("flags", pflag.ExitOnError)
	flags.ParseErrorsWhitelist.UnknownFlags = true
	flags.SetInterspersed(true)

	ver := flags.BoolP("version", "v", false, "print version and exit")
	daemon := flags.Bool("daemon", false, "run as a daemon")
	title := flags.String("title", "", "notification title")
	priority := flags.Int("priority", 0, "notification priority")

	if err := flags.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	if *ver {
		fmt.Printf("%s (%s)\n", version, commit[:7])
		return
	}

	if *daemon {
		ss, err := LoadSenders()
		if err != nil {
			panic(err)
		}

		u, err := NewUnixListener(UnixSocket)
		if err != nil {
			panic(err)
		}

		Daemon([]Listener{u}, ss)
		return
	}

	msg := Message{Timestamp: time.Now(), Hostname: Hostname, Title: *title, Priority: *priority}

	if sendmail {
		mail, err := parsemail.Parse(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse mail: %v\n", err)
			os.Exit(1)
		}

		msg.Text = strings.TrimRight(mail.TextBody, "\n")

		if msg.Title == "" {
			msg.Title = strings.TrimSpace(mail.Subject)
		}
	} else {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read from stdin: %v\n", err)
			os.Exit(1)
		}

		msg.Text = strings.TrimRight(string(b), "\n")
	}

	if conn, err := net.Dial("unixgram", UnixSocket); err == nil {
		defer conn.Close()

		if err := json.NewEncoder(conn).Encode(msg); err == nil {
			return
		}
	}

	ss, err := LoadSenders()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, err = range Send(msg, ss) {
		if err == nil {
			return
		}

		fmt.Fprintf(os.Stderr, "failed to send: %v\n", err)
	}

	os.Exit(1)
}

var Hostname string

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	Hostname = strings.SplitN(hostname, ".", 2)[0]
}

// Will be overwritten by goreleaser
var version string = "dev"
var commit string = "0000000"
