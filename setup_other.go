//go:build !(freebsd || linux)
package main

func installService() {
	panic("setting up as a daemon not supported on this OS")
}

const daemonUser = ""
