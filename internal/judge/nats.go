package judge

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/ocontest/backend/pkg/configs"
	"github.com/ocontest/backend/pkg/structs"
)

type JudgeQueue interface {
	Send(req structs.JudgeRequest) (structs.JudgeResponse, error)
	Subscribe() (*nats.Subscription, error)
}

type JudgeQueueImp struct {
	conn   *nats.Conn
	sub    *nats.Subscription
	config configs.SectionNats
}

func NewJudgeQueue(c configs.SectionNats) (JudgeQueue, error) {
	conn, err := nats.Connect(c.Url)
	if err != nil {
		return nil, err
	}
	ans := JudgeQueueImp{
		conn:   conn,
		config: c,
	}
	return ans, err
}

func (j JudgeQueueImp) Send(req structs.JudgeRequest) (resp structs.JudgeResponse, err error) {
	data, err := json.Marshal(req)
	if err != nil {
		return
	}
	msg, err := j.conn.Request(j.config.Subject, data, j.config.ReplyTimeout)
	if err != nil {
		return
	}
	err = json.Unmarshal(msg.Data, &resp)
	return
}
func (j JudgeQueueImp) Subscribe() (*nats.Subscription, error) {

	return j.conn.QueueSubscribeSync(j.config.Subject, j.config.Queue)

}
