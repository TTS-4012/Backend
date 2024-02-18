package consumer

import (
	"github.com/nats-io/nats.go"
	"github.com/ocontest/backend/internal/judge"
	"github.com/ocontest/backend/pkg/configs"
	"github.com/pkg/errors"
)

type RunnerConsumer interface {
	StartListen()
	ProcessCode(msg *nats.Msg)
}

type RunnerConsumerImp struct {
	queue judge.JudgeQueue
}

func NewRunnerScheduler(natsConfig configs.SectionNats) (RunnerConsumer, error) {
	queue, err := judge.NewJudgeQueue(natsConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return RunnerConsumerImp{
		queue: queue,
	}, nil
}
