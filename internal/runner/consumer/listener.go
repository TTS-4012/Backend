package consumer

import (
	"log"

	"github.com/nats-io/nats.go"
	"github.com/ocontest/backend/internal/runner"
	"github.com/ocontest/backend/pkg"
	"github.com/pkg/errors"
)

func (r RunnerConsumerImp) StartListen() {
	sub, err := r.queue.Subscribe()
	if err != nil {
		log.Fatal("couldn't subscribe", err)
	}

	for {
		msg, err := sub.NextMsg(runner.NatsTimeout)
		if errors.Is(err, nats.ErrTimeout) {
			continue
		}
		if err != nil {
			pkg.Log.Error("error on getting message from queue: ", err)
			continue
		}

		pkg.Log.Debug("got msg from nats")
		r.ProcessCode(msg)
	}
}
