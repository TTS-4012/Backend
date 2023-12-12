package main

import (
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"ocontest/internal/judge"
	"ocontest/pkg"
	"ocontest/pkg/configs"
)

type RunnerScheduler interface {
	StartListen()
	ProcessCode(msg *nats.Msg)
}

type RunnerSchedulerImp struct {
	queue judge.JudgeQueue
}

func NewRunnerScheduler(natsConfig configs.SectionNats) (RunnerScheduler, error) {
	queue, err := judge.NewJudgeQueue(natsConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return RunnerSchedulerImp{
		queue: queue,
	}, nil
}

func (r RunnerSchedulerImp) StartListen() {
	for {
		msg, err := r.queue.Get()
		if err != nil {
			pkg.Log.Error("error on getting message from queue: ", err)
		}
		r.ProcessCode(msg)
	}
}

func (r RunnerSchedulerImp) ProcessCode(msg *nats.Msg) {
	//TODO implement me

	panic("implement me")
}
