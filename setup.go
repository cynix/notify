package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"
)


func Setup(daemon bool) {
	file := tokenPath()
	stdin := int(os.Stdin.Fd())

	if !terminal.IsTerminal(stdin) {
		fmt.Fprintln(os.Stderr, "stdin is not a terminal")
		os.Exit(1)
	}

	fmt.Fprint(os.Stderr, "Telegram bot token: ")
	token, err := terminal.ReadPassword(stdin)
	fmt.Fprintln(os.Stderr)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read token: %v\n", err)
		os.Exit(1)
	}

	if len(token) == 0 {
		fmt.Fprintln(os.Stderr, "token empty")
		os.Exit(1)
	}

	os.Chmod(file, 0600)
	if err = os.WriteFile(file, token, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write token to %s: %v\n", file, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "token written to %s\n", file)

	if daemon {
		if daemonUser == "" {
			fmt.Fprintln(os.Stderr, "setting up as a daemon not supported on this OS")
			os.Exit(1)
		}

		if !strings.HasPrefix(file, EtcDir) {
			fmt.Fprintf(os.Stderr, "setting up as a daemon not supported with token file outisde %s\n", EtcDir)
			os.Exit(1)
		}

		exe, err := os.Executable()
		if err != nil || exe != "/usr/local/bin/notify" {
			fmt.Fprintf(os.Stderr, "setting up as a daemon not supported with executable at %s (%v)\n", exe, err)
			os.Exit(1)
		}

		u, err := user.Lookup(daemonUser)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to lookup user '%s': %v\n", daemonUser, err)
			os.Exit(1)
		}

		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			panic(err)
		}

		gid, err := strconv.Atoi(u.Gid)
		if err != nil {
			panic(err)
		}

		if err = os.Chown(file, uid, gid); err != nil {
			fmt.Fprintf(os.Stderr, "failed to chown %s: %v\n", file, err)
			os.Exit(1)
		}

		installService()
	}

	os.Exit(0)
}

func tokenPath() string {
	if unix.Access(EtcDir, unix.W_OK) == nil {
		return path.Join(EtcDir, TokenFile)
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	if unix.Access(dir, unix.W_OK) == nil {
		return path.Join(dir, TokenFile)
	}

	fmt.Fprintln(os.Stderr, "unable to determine token path")
	os.Exit(1)
	return ""
}
