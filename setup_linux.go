//go:build linux
package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)


func installService() {
	if unix.Access("/sbin/openrc-run", unix.X_OK) != nil {
		fmt.Fprintf(os.Stderr, "setting up as daemon not supported without OpenRC")
		os.Exit(1)
	}

	if err := os.WriteFile("/etc/init.d/notify", []byte(strings.ReplaceAll(rc, "%%USERNAME%%", daemonUser)), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create /etc/init.d/notify: %v\n", err)
		os.Exit(1)
	}

	if dst, err := os.Readlink("/etc/runlevels/default/notify"); err != nil || dst != "/etc/init.d/notify" {
		os.Remove("/etc/runlevels/default/notify")

		if err := os.Symlink("/etc/init.d/notify", "/etc/runlevels/default/notify"); err != nil {
			fmt.Fprintf(os.Stderr, "failed to symlink /etc/runlevels/default/notify: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Fprintln(os.Stderr, "notify service installed at default runlevel")
}

const daemonUser = "mail"

//go:embed openrc.in
var rc string
