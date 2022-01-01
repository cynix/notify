//go:build freebsd
package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

func installService() {
	if err := os.WriteFile("/usr/local/etc/rc.d/notify", []byte(strings.ReplaceAll(rc, "%%USERNAME%%", daemonUser)), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write to /usr/local/etc/rc.d/notify: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("/usr/local/etc/rc.conf.d/notify", []byte("notify_enable=\"YES\"\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write to /usr/local/etc/rc.conf.d/notify: %v\n", err)
		os.Exit(1)
	}

	if dst, err := os.Readlink("/usr/local/bin/notify-sendmail"); err != nil || dst != "/usr/local/bin/notify" {
		os.Remove("/usr/local/bin/notify-sendmail")

		if err := os.Symlink("notify", "/usr/local/bin/notify-sendmail"); err != nil {
			fmt.Fprintf(os.Stderr, "failed to symlink /usr/local/bin/notify-sendmail: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Fprintln(os.Stderr, "notify service installed")
}

const daemonUser = "mailnull"

//go:embed freebsd.in
var rc string
