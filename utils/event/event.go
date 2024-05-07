package event

import "time"

type Event struct {
	Name   string
	Time   time.Time
	Detail string
	Addon  interface{}
}
