package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/rfizzle/parsemail"
	"github.com/spf13/pflag"
)


func main() {
	sendmail := strings.Contains(os.Args[0], "sendmail")

	flags := pflag.NewFlagSet("flags", pflag.ExitOnError)
	flags.ParseErrorsWhitelist.UnknownFlags = true
	flags.SetInterspersed(true)

	ver := flags.BoolP("version", "v", false, "print version and exit")
	setup := flags.Bool("setup", false, "initial setup")
	daemon := flags.Bool("daemon", false, "run (or setup) as a daemon")
	markdown := flags.Bool("markdown", false, "use markdown formatting")

	var from, to string
	flags.StringVarP(&from, "from", "f", "", "sender")

	if err := flags.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	if *ver {
		fmt.Printf("%s (%s)\n", version, commit[:7])
		os.Exit(0)
	}

	if *setup {
		Setup(*daemon)
		return
	}

	if *daemon {
		LoadToken()

		var d Daemon
		d.Serve()
		return
	}

	var text string

	if sendmail {
		mail, err := parsemail.Parse(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse mail: %v\n", err)
			os.Exit(1)
		}

		if from == "" {
			if mail.Sender != nil {
				from = mail.Sender.String()
			}

			if from == "" {
				for _, f := range mail.From {
					from = f.String()

					if from != "" {
						break
					}
				}
			}
		}

		if to == "" {
			for _, t := range mail.To {
				to = t.String()

				if to != "" {
					break
				}
			}
		}

		if subject := strings.TrimSpace(mail.Subject); subject != "" {
			text = "*" + Escaper.Replace(subject) + "*\n\n"
		}

		text = fmt.Sprintf(
			"%s%s\n\n%s \u2192 %s \\#%s",
			text, Escaper.Replace(strings.TrimRight(mail.TextBody, "\n")),
			Escaper.Replace(from), Escaper.Replace(to), Hostname)
	} else {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read from stdin: %v\n", err)
			os.Exit(1)
		}

		text = strings.TrimRight(string(b), "\n")

		if !*markdown {
			text = Escaper.Replace(text)
		}

		if from != "" {
			text = fmt.Sprintf("%s\n\n\\#%s \\#%s", text, Escaper.Replace(from), Hostname)
		} else {
			text = fmt.Sprintf("%s\n\n\\#%s", text, Hostname)
		}
	}

	if conn, err := net.Dial("unixgram", Socket); err == nil {
		defer conn.Close()

		if n, err := conn.Write([]byte(text)); err != nil || n < len(text) {
			fmt.Fprintf(os.Stderr, "failed to send message: %v (%d)\n", err, n)
			os.Exit(1)
		}
	} else {
		LoadToken()

		if err := SendTelegram(text); err != nil {
			fmt.Fprintf(os.Stderr, "failed to send message: %v\n", err)
			os.Exit(1)
		}
	}
}

var Hostname string
var Escaper = strings.NewReplacer(
	"\\", "\\\\",
	"_", "\\_",
	"*", "\\*",
	"[", "\\[",
	"]", "\\]",
	"(", "\\(",
	")", "\\)",
	"~", "\\~",
	"`", "\\`",
	">", "\\>",
	"#", "\\#",
	"+", "\\+",
	"-", "\\-",
	"=", "\\=",
	"|", "\\|",
	"{", "\\{",
	"}", "\\}",
	".", "\\.",
	"!", "\\!")

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
