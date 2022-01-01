package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
)


func LoadToken() {
	b, err := os.ReadFile(path.Join(EtcDir, TokenFile))
	if err == nil && len(b) > 0 {
		ConfigureTelegram(strings.TrimSpace(string(b)))
		return
	}

	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	b, err = os.ReadFile(path.Join(u.HomeDir, "." + TokenFile))
	if err == nil && len(b) > 0 {
		ConfigureTelegram(strings.TrimSpace(string(b)))
		return
	}

	fmt.Fprintf(os.Stderr, "could not load token from file")
	os.Exit(1)
}

const (
	EtcDir = "/usr/local/etc"
	TokenFile = "notify.token"
)
