package main

import "time"

type Message struct {
	Timestamp time.Time
	Hostname string
	Text string
	Title string
	Priority int
}
