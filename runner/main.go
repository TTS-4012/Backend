package main

import (
	"bytes"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"ocontest/internal/judge"
	"ocontest/pkg"
	"ocontest/pkg/configs"
	"ocontest/pkg/structs"
	"strings"
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
	logger := pkg.Log.WithFields(logrus.Fields{
		"module":  "runner",
		"subject": msg.Subject,
	})

	var task structs.JudgeRequest
	var resp structs.JudgeResponse

	err := json.Unmarshal(msg.Data, &task)
	if err != nil {
		pkg.Log.Error("error on unmarshal message: ", err)
		msg.Respond([]byte("error on unmarshal message"))
	}

	for ind := range task.Testcases {
		testCase := task.Testcases[ind]
		resp.TestResults[ind].SubmissionID = task.SubmissionID
		resp.TestResults[ind].TestcaseID = testCase.ID

		input := bytes.NewReader([]byte(testCase.Input))
		var output, stderr bytes.Buffer
		verdict, err := RunTask(TimeLimit, MemoryLimit, task.Code, input, &output, &stderr)
		if err != nil {
			logger.Error("error on running code: ", err)
			verdict = structs.VerdictUnknown
			resp.ServerError = err.Error()
		}
		outputStr := output.String()
		stderrStr := stderr.String()
		if stderrStr != "" {
			logger.Warning("stderr is not empty: ", stderrStr)
		}
		resp.TestResults[ind].RunnerError = stderrStr
		resp.TestResults[ind].RunnerOutput = outputStr

		if verdict != structs.VerdictOK {
			continue
		}
		if verdict == structs.VerdictOK && r.checkOutput(outputStr, task.Testcases[ind].ExpectedOutput) {

		}

	}

	panic("implement me")
}

func (r RunnerSchedulerImp) checkOutput(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	return actual == expected

}
