package common

import (
	"log"
	"time"
)

type Worker struct {
	Stopped         bool
	ShutdownChannel chan int
	Interval        time.Duration
	Action          func()
}

func (t *Worker) Run() {
	log.Print("Garbage colleciton worker started running")

	t.Stopped = false

	for {
		select {
		case <-t.ShutdownChannel:
			return
		case <-time.After(t.Interval):
			break
		}

		t.Action()
	}
}
