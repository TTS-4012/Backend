package judge

import (
	"context"

	"github.com/ocontest/backend/internal/db/repos"
	"github.com/ocontest/backend/internal/minio"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/configs"
	"github.com/ocontest/backend/pkg/structs"

	"github.com/pkg/errors"
)

type Judge interface {
	Dispatch(ctx context.Context, submissionID, contestID int64) (err error)
	GetTestResults(ctx context.Context, id string) (structs.JudgeResponse, error)
	GetScore(ctx context.Context, id string) (int, error)
}

type JudgeImp struct {
	queue                  JudgeQueue
	contestUsersRepo       repos.ContestsUsersRepo
	problemsRepo           repos.ProblemsMetadataRepo
	submissionMetadataRepo repos.SubmissionMetadataRepo
	minioHandler           minio.MinioHandler
	testcaseRepo           repos.TestCaseRepo
	judgeRepo              repos.JudgeRepo
}

func NewJudge(c configs.SectionJudge, submissionMetadataRepo repos.SubmissionMetadataRepo,
	minioHandler minio.MinioHandler, testcaseRepo repos.TestCaseRepo, contestUsersRepo repos.ContestsUsersRepo, judgeRepo repos.JudgeRepo, problemsRepo repos.ProblemsMetadataRepo) (Judge, error) {
	queue, err := NewJudgeQueue(c.Nats)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create judge queue for judge")
	}

	return JudgeImp{
		queue:                  queue,
		problemsRepo:           problemsRepo,
		submissionMetadataRepo: submissionMetadataRepo,
		minioHandler:           minioHandler,
		judgeRepo:              judgeRepo,
		testcaseRepo:           testcaseRepo,
		contestUsersRepo:       contestUsersRepo,
	}, nil
}

func (j JudgeImp) Dispatch(ctx context.Context, submissionID, contestID int64) (err error) {
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
		return errors.Wrap(err, "couldn't insert judge result to judge repos")
	}

	currentScore := j.CalcScore(resp.TestResults)
	lastSub, err := j.submissionMetadataRepo.GetFinalSubmission(ctx, submission.ProblemID, submission.UserID, contestID)
	if err != nil && !errors.Is(err, pkg.ErrNotFound) {
		return errors.Wrap(err, "coudn't get last submission")
	}

	isFinal := true
	if lastSub.Score > currentScore {
		isFinal = false
	}

	err = j.submissionMetadataRepo.UpdateJudgeResults(ctx, submission.ProblemID, submission.UserID, submission.ContestID, submissionID, docID, currentScore, isFinal)
	if err != nil {
		return errors.Wrap(err, "couldn't update judge result in submission metadata repos")
	}
	err = j.problemsRepo.AddSolve(ctx, submission.ProblemID)
	if err != nil{
		return errors.Wrap(err, "coudn't update solve count")
	}

	if isFinal && contestID != 0 {
		err = j.contestUsersRepo.AddUserScore(ctx, submission.UserID, contestID, currentScore-lastSub.Score)
		if err != nil {
			return errors.Wrap(err, "couldn't update contest score")
		}
	}

	return nil
}

func (j JudgeImp) GetTestResults(ctx context.Context, id string) (structs.JudgeResponse, error) {
	return j.judgeRepo.GetResults(ctx, id)
}

func (j JudgeImp) CalcScore(t []structs.TestResult) int {
	total := len(t)
	if total == 0 {
		return 0
	}

	correct := 0
	for _, r := range t {
		if r.Verdict == structs.VerdictOK {
			correct++
		}
	}

	return 100 * correct / total
}

func (j JudgeImp) GetScore(ctx context.Context, id string) (int, error) {
	results, err := j.judgeRepo.GetResults(ctx, id)
	if err != nil {
		pkg.Log.Error("error on get score: ", err)
		return 0, err
	}
	return j.CalcScore(results.TestResults), nil

}
