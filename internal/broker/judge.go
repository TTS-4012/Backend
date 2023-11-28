package broker

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"ocontest/pkg/configs"
	"ocontest/pkg/structs"
)

type JudgeQueue interface {
	Send(s structs.JudgeSubmissions) error
	Sub() (<-chan structs.JudgeSubmissions, <-chan error, error)
}

type JudgeQueueImp struct {
	conn   *nats.Conn
	config configs.SectionNats
}

func NewJudgeQueue(c configs.SectionNats) (JudgeQueue, error) {
	conn, err := nats.Connect(c.Url)
	if err != nil {
		return nil, err
	}
	return JudgeQueueImp{
		conn:   conn,
		config: c,
	}, nil
}

func (j JudgeQueueImp) Send(s structs.JudgeSubmissions) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return j.conn.Publish(j.config.Subject, data)
}

func (j JudgeQueueImp) Sub() (<-chan structs.JudgeSubmissions, <-chan error, error) {
	ans := make(chan structs.JudgeSubmissions, j.config.SubscribeChanSize)
	errors := make(chan error, j.config.SubscribeChanSize)

	_, err := j.conn.QueueSubscribe(j.config.Subject, j.config.Queue, func(m *nats.Msg) {
		var submission structs.JudgeSubmissions
		err := json.Unmarshal(m.Data, &submission)
		if err != nil {
			errors <- err
		} else {
			ans <- submission
		}
	})
	if err != nil {
		return nil, nil, err
	}
	return ans, errors, nil
}
