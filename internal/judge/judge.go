package judge

import (
	"context"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/internal/minio"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/configs"
	"github.com/ocontest/backend/pkg/structs"

	"github.com/pkg/errors"
)

type Judge interface {
	Dispatch(ctx context.Context, submissionID int64) (err error)
	GetTestresults(ctx context.Context, id string) (structs.JudgeResponse, error)
	GetScore(ctx context.Context, id string) (int, error)
}

type JudgeImp struct {
	queue                  JudgeQueue
	submissionMetadataRepo db.SubmissionMetadataRepo
	minioHandler           minio.MinioHandler
	testcaseRepo           db.TestCaseRepo
	judgeRepo              db.JudgeRepo
}

func NewJudge(c configs.SectionJudge, submissionMetadataRepo db.SubmissionMetadataRepo,
	minioHandler minio.MinioHandler, testcaseRepo db.TestCaseRepo, judgeRepo db.JudgeRepo) (Judge, error) {
	queue, err := NewJudgeQueue(c.Nats)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create judge queue for judge")
	}

	return JudgeImp{
		queue:                  queue,
		submissionMetadataRepo: submissionMetadataRepo,
		minioHandler:           minioHandler,
		judgeRepo:              judgeRepo,
		testcaseRepo:           testcaseRepo,
	}, nil
}

func (j JudgeImp) Dispatch(ctx context.Context, submissionID int64) (err error) {
	submission, err := j.submissionMetadataRepo.Get(ctx, submissionID)
	if err != nil {
		err = errors.Wrap(err, "couldn't get submission from db")
		return
	}
	codeObjectName := j.minioHandler.GenCodeObjectname(submission.UserID, submission.ProblemID, submission.ID)
	code, _, err := j.minioHandler.DownloadFile(ctx, codeObjectName)
	if err != nil {
		err = errors.Wrap(err, "couldn't get code from minio to judge")
		return
	}
	testCases, err := j.testcaseRepo.GetAllTestsOfProblem(ctx, submission.ProblemID)
	if err != nil {
		err = errors.Wrap(err, "couldn't get test cases from db")
		return
	}
	req := structs.JudgeRequest{
		SubmissionID: submissionID,
		Code:         string(code),
		Testcases:    testCases,
	}

	resp, err := j.queue.Send(req)
	if err == nil && resp.ServerError != "" {
		err = errors.New(resp.ServerError)
	}

	if err != nil {
		err = errors.Wrap(err, "error on send to queue")
		return
	}

	docID, err := j.judgeRepo.Insert(ctx, resp)
	if err != nil {
		return errors.Wrap(err, "couldn't insert judge result to judge repo")
	}

	//lastScore :=
	err = j.submissionMetadataRepo.UpdateJudgeResults(ctx, submission.ProblemID, submission.UserID, submissionID, docID)
	if err != nil {
		return errors.Wrap(err, "couldn't update judge result in submission metadata repo")
	}

	return nil
}

func (j JudgeImp) GetTestresults(ctx context.Context, id string) (structs.JudgeResponse, error) {
	return j.judgeRepo.GetResults(ctx, id)
}

func (j JudgeImp) GetScore(ctx context.Context, id string) (int, error) {
	results, err := j.judgeRepo.GetResults(ctx, id)
	if err != nil {
		pkg.Log.Error("error on get score: ", err)
		return 0, err
	}

	total := len(results.TestResults)
	if total == 0 {
		return 0, nil
	}

	correct := 0
	for _, r := range results.TestResults {
		if r.Verdict == structs.VerdictOK {
			correct++
		}
	}

	return 100 * correct / 100, nil
}
