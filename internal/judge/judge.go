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
	Do(ctx context.Context,  submissionID int64) (err error)
}

type JudgeImp struct {
	queue                   JudgeQueue
	submissionMetadataRepo  db.SubmissionMetadataRepo
	minioHandler            minio.MinioHandler
	problemsDescriptionRepo db.ProblemDescriptionsRepo
	problemsMetadataRepo    db.ProblemsMetadataRepo
}

func NewJudge(c configs.SectionJudge) (Judge, error) {
	queue, err := NewJudgeQueue(c.Nats)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create judge queue for judge")
	}

	return JudgeImp{
		queue: queue,
	}, nil
}

func (j JudgeImp) processTestResponse (submission structs.SubmissionMetadata, resp structs.JudgeResponse) (structs.SubmissionMetadata){
	panic("not implemented yet")
}

func (j JudgeImp) Do(ctx context.Context, submissionID int64) (err error) {
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
		Code: code,
		Testcases: problemDescription.Testcases,
	}

	resp, err := j.queue.Send(req)
	if err == nil && resp.ServerError != nil{
		err = resp.ServerError
	}

	if err != nil{
		err = errors.Wrap(err, "error on send to queue")
		return
	}
	submission = j.processTestResponse(submission, resp)

	panic("not implemented yet!")
}
