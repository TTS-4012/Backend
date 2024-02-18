package consumer

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/ocontest/backend/internal/runner"
	taskRunner "github.com/ocontest/backend/internal/runner/task-runner"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
	"github.com/sirupsen/logrus"
)

func (r RunnerConsumerImp) ProcessCode(msg *nats.Msg) {
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
	logger.Debug("Recieved task ", task.SubmissionID, " number of tests:", len(task.Testcases))
	resp.TestResults = make([]structs.TestResult, len(task.Testcases))
	for ind := range task.Testcases {
		testCase := task.Testcases[ind]
		resp.TestResults[ind].SubmissionID = task.SubmissionID
		resp.TestResults[ind].TestcaseID = testCase.ID

		input := bytes.NewReader([]byte(testCase.Input))
		var output, stderr bytes.Buffer
		verdict, err := taskRunner.RunTask(runner.TimeLimit, runner.MemoryLimit, task.Code, input, &output, &stderr)
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
		if verdict == structs.VerdictOK {
			if r.checkOutput(outputStr, task.Testcases[ind].ExpectedOutput) {
				resp.TestResults[ind].Verdict = structs.VerdictOK
			} else {
				resp.TestResults[ind].Verdict = structs.VerdictWrong
			}
		}

	}

	respData, err := json.Marshal(resp)
	if err != nil {
		errorMessage := "error on json marshalling response"
		pkg.Log.Error(errorMessage)
		respData = []byte(errorMessage)
	}
	err = msg.Respond(respData)
	if err != nil {
		pkg.Log.Error("error on respond to judge task", err)
	}

}

func (r RunnerConsumerImp) checkOutput(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	return actual == expected

}
