package judge

import (
	"encoding/json"
	"ocontest/pkg"
	"ocontest/pkg/configs"
	"ocontest/pkg/structs"

	"github.com/nats-io/nats.go"
)

type JudgeQueue interface {
	Send(req structs.JudgeRequest) (structs.JudgeResponse, error)
	Get() (*nats.Msg, error)
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
	err = ans.StartSubscribe()
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

func (j JudgeQueueImp) StartSubscribe() error {

	if j.sub != nil {
		pkg.Log.Error("subscribe called when j.sub is not nil")
		return pkg.ErrInternalServerError
	}

	sub, err := j.conn.QueueSubscribeSync(j.config.Subject, j.config.Queue)
	if err != nil {
		return err
	}
	j.sub = sub
	return nil
}

func (j JudgeQueueImp) Get() (msg *nats.Msg, err error) {
	if j.sub == nil {
		pkg.Log.Error("GetByID Next called when there is no subscription")
		return nil, pkg.ErrBadRequest
	}
	return j.sub.NextMsg(j.config.ReplyTimeout)
}
