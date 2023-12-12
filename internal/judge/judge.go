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
	queue                   JudgeQueue
	submissionMetadataRepo  db.SubmissionMetadataRepo
	minioHandler            minio.MinioHandler
	problemsDescriptionRepo db.ProblemDescriptionsRepo
	problemsMetadataRepo    db.ProblemsMetadataRepo
	judgeRepo               db.JudgeRepo
}

func NewJudge(c configs.SectionJudge, submissionMetadataRepo db.SubmissionMetadataRepo,
	minioHandler minio.MinioHandler, problemDescriptionRepo db.ProblemDescriptionsRepo,
	problemMetadataRepo db.ProblemsMetadataRepo, judgeRepo db.JudgeRepo) (Judge, error) {
	queue, err := NewJudgeQueue(c.Nats)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create judge queue for judge")
	}

	return JudgeImp{
		queue:                   queue,
		submissionMetadataRepo:  submissionMetadataRepo,
		minioHandler:            minioHandler,
		problemsDescriptionRepo: problemDescriptionRepo,
		problemsMetadataRepo:    problemMetadataRepo,
		judgeRepo:               judgeRepo,
	}, nil
}

func (j JudgeImp) Dispatch(ctx context.Context, submissionID int64) (err error) {
	submission, err := j.submissionMetadataRepo.Get(ctx, submissionID)
	if err != nil {
		return
	}
	codeObjectName := j.minioHandler.GenCodeObjectname(submission.UserID, submission.ProblemID, submission.ID)
	code, _, err := j.minioHandler.DownloadFile(ctx, codeObjectName)
	if err != nil {
		err = errors.Wrap(err, "couldn't get code from minio to judge")
		return
	}
	problemMeta, err := j.problemsMetadataRepo.GetProblem(ctx, submission.ProblemID)
	if err != nil {
		err = errors.Wrap(err, "couldn't get problem from db")
		return
	}

	problemDescription, err := j.problemsDescriptionRepo.Get(problemMeta.DocumentID)
	if err != nil {
		err = errors.Wrap(err, "couldn't get problem desc from db")
	}

	req := structs.JudgeRequest{
		Code:      code,
		Testcases: problemDescription.Testcases,
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
