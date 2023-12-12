package main

import (
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"ocontest/internal/judge"
	"ocontest/pkg"
	"ocontest/pkg/configs"
)

type RunnerTaskHandler interface {
	StartListern()
	ProcessCode(msg *nats.Msg)
}

type RunnerImp struct {
	queue judge.JudgeQueue
}

func NewRunner(natsConfig configs.SectionNats) (RunnerTaskHandler, error) {
	queue, err := judge.NewJudgeQueue(natsConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return RunnerImp{
		queue: queue,
	}, nil
}

func (r RunnerImp) StartListern() {
	for {
		msg, err := r.queue.Get()
		if err != nil {
			pkg.Log.Error("error on getting message from queue: ", err)
		}
		r.ProcessCode(msg)
	}
}

func (r RunnerImp) ProcessCode(msg *nats.Msg) {
	//TODO implement me

	panic("implement me")
}
