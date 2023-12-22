package db

import (
	"context"

	"github.com/ocontest/backend/pkg/structs"
)

type AuthRepo interface {
	InsertUser(ctx context.Context, user structs.User) (int64, error)
	VerifyUser(ctx context.Context, userID int64) error
	GetByUsername(ctx context.Context, username string) (structs.User, error)
	GetByID(ctx context.Context, userID int64) (structs.User, error)
	GetByEmail(ctx context.Context, email string) (structs.User, error)
	UpdateUser(ctx context.Context, user structs.User) error
}

type ProblemsMetadataRepo interface {
	InsertProblem(ctx context.Context, problem structs.Problem) (int64, error)
	GetProblem(ctx context.Context, id int64) (structs.Problem, error)
	ListProblems(ctx context.Context, searchCol string, descending bool, limit, offset int, getCount bool) ([]structs.Problem, int, error)
	UpdateProblem(ctx context.Context, id int64, title string) error
	DeleteProblem(ctx context.Context, id int64) (string, error)
}

type ContestsMetadataRepo interface {
	InsertContest(ctx context.Context, contest structs.Contest) (int64, error)
	GetContest(ctx context.Context, id int64) (structs.Contest, error)
	ListContests(ctx context.Context, descending bool, limit, offset int) ([]structs.Contest, error)
	DeleteContest(ctx context.Context, id int64) error
}

type ContestsProblemsRepo interface {
	GetContestProblems(ctx context.Context, contestID int64) ([]int64, error)
	AddProblem(ctx context.Context, contestID int64, problemID int64) error
}

type ProblemDescriptionsRepo interface {
	Insert(description string, testCases []string) (string, error)
	Get(id string) (structs.ProblemDescription, error)
	Update(id string, description string) error
	Delete(id string) error
}

type TestCaseRepo interface {
	Insert(ctx context.Context, testCase structs.Testcase) (int64, error)
	GetByID(ctx context.Context, id int64) (structs.Testcase, error)
	GetAllTestsOfProblem(ctx context.Context, problemID int64) ([]structs.Testcase, error)
}

type SubmissionMetadataRepo interface {
	Insert(ctx context.Context, submission structs.SubmissionMetadata) (int64, error)
	Get(ctx context.Context, id int64) (structs.SubmissionMetadata, error)
	AddJudgeResult(ctx context.Context, submissionID int64, judgeResultID string) error
	ListSubmissions(ctx context.Context, problemID, userID int64, descending bool, limit, offset int, getCount bool) ([]structs.SubmissionMetadata, int, error)
}

type JudgeRepo interface {
	Insert(ctx context.Context, response structs.JudgeResponse) (string, error)
	GetResults(ctx context.Context, id string) (structs.JudgeResponse, error)
}
