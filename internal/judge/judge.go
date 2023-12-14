package judge

import (
	"context"
	"ocontest/internal/db"
	"ocontest/internal/minio"
	"ocontest/pkg/configs"
	"ocontest/pkg/structs"

	"github.com/pkg/errors"
)

type Judge interface {
	Dispatch(ctx context.Context, submissionID int64) (err error)
	GetResults(ctx context.Context, id string) (structs.JudgeResponse, error)
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

	err = j.submissionMetadataRepo.AddJudgeResult(ctx, submissionID, docID)
	if err != nil {
		return errors.Wrap(err, "couldn't update judge result in submission metadata repo")
	}

	return nil
}

func (j JudgeImp) GetResults(ctx context.Context, id string) (structs.JudgeResponse, error) {
	return j.judgeRepo.GetResults(ctx, id)
}
