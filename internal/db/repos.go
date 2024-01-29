package db

import (
	"context"

	"github.com/ocontest/backend/pkg/structs"
)

type UsersRepo interface {
	InsertUser(ctx context.Context, user structs.User) (int64, error)
	VerifyUser(ctx context.Context, userID int64) error
	GetByUsername(ctx context.Context, username string) (structs.User, error)
	GetByID(ctx context.Context, userID int64) (structs.User, error)
	GetUsername(ctx context.Context, userID int64) (string, error)
	GetByEmail(ctx context.Context, email string) (structs.User, error)
	UpdateUser(ctx context.Context, user structs.User) error
}

type ProblemsMetadataRepo interface {
	InsertProblem(ctx context.Context, problem structs.Problem) (int64, error)
	GetProblem(ctx context.Context, id int64) (structs.Problem, error)
	GetProblemTitle(ctx context.Context, id int64) (string, error)
	ListProblems(ctx context.Context, searchCol string, descending bool, limit, offset int, getCount bool) ([]structs.Problem, int, error)
	UpdateProblem(ctx context.Context, id int64, title string, hardness int64) error
	DeleteProblem(ctx context.Context, id int64) (string, error)
}

type ContestsMetadataRepo interface {
	InsertContest(ctx context.Context, contest structs.Contest) (int64, error)
	GetContest(ctx context.Context, id int64) (structs.Contest, error)
	ListContests(ctx context.Context, descending bool, limit, offset int, started bool, userID int64, owned, getCount bool) ([]structs.Contest, int, error)
	ListMyContests(ctx context.Context, descending bool, limit, offset int, started bool, userID int64, getCount bool) ([]structs.Contest, int, error)
	UpdateContests(ctx context.Context, id int64, newContest structs.RequestUpdateContest) error
	DeleteContest(ctx context.Context, id int64) error
}

type ProblemDescriptionsRepo interface {
	Insert(description string, testCases []string) (string, error)
	Get(id string) (structs.ProblemDescription, error)
	Update(id string, description string) error
	Delete(id string) error
}

type ContestsProblemsRepo interface {
	AddProblemToContest(ctx context.Context, contestID, problemID int64) error
	GetContestProblems(ctx context.Context, id int64) ([]int64, error)
	RemoveProblemFromContest(ctx context.Context, contestID, problemID int64) error
}

type TestCaseRepo interface {
	Insert(ctx context.Context, testCase structs.Testcase) (int64, error)
	GetByID(ctx context.Context, id int64) (structs.Testcase, error)
	GetAllTestsOfProblem(ctx context.Context, problemID int64) ([]structs.Testcase, error)
}

type SubmissionMetadataRepo interface {
	Insert(ctx context.Context, submission structs.SubmissionMetadata) (int64, error)
	Get(ctx context.Context, id int64) (structs.SubmissionMetadata, error)
	GetByProblem(ctx context.Context, problemID int64) ([]structs.SubmissionMetadata, error)
	GetFinalSubmission(ctx context.Context, userID, problemID int64) (structs.SubmissionMetadata, error)
	UpdateJudgeResults(ctx context.Context, problemID, userID, submissionID int64, judgeResultID string, score int, isFinal bool) error
	ListSubmissions(ctx context.Context, problemID, userID, contestID int64, descending bool, limit, offset int, getCount bool) ([]structs.SubmissionMetadata, int, error)
}

type JudgeRepo interface {
	Insert(ctx context.Context, response structs.JudgeResponse) (string, error)
	GetResults(ctx context.Context, id string) (structs.JudgeResponse, error)
}

type ContestsUsersRepo interface {
	Add(ctx context.Context, contestID, userID int64) error
	Delete(ctx context.Context, contestID, userID int64) error
	IsRegistered(ctx context.Context, contestID, userID int64) (bool, error)
	ListUsersByScore(ctx context.Context, contestID int64, limit, offset int) ([]structs.User, error)
	GetContestUsersCount(ctx context.Context, contestID int64) (int, error)
	AddUserScore(ctx context.Context, userID, contestID int64, delta int) error
}
