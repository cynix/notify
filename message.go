package main

import "time"

type Message struct {
	Timestamp time.Time
	Text string
	Title string
	Priority int
}
