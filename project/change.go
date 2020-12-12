package main

import (
	"time"
)

type Change struct {
	kind string
	body string
	time time.Time
}
