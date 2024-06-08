package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cynix/notify"
	"github.com/on2itsecurity/parsemail"
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
		fmt.Println(Version)
		return
	}

	if *daemon {
		ss, err := notify.LoadSenders()
		if err != nil {
			panic(err)
		}

		u, err := notify.NewUnixListener()
		if err != nil {
			panic(err)
		}

		notify.Daemon([]notify.Listener{u}, ss)
		return
	}

	msg := notify.Message{Title: *title, Priority: *priority}

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

	if err := notify.Notify(msg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Will be overwritten by goreleaser
var Version string = "dev"
