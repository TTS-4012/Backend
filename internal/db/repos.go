package db

import (
	"context"
	"ocontest/pkg/structs"
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
	ListProblems(ctx context.Context, searchCol string, descending bool, limit, offset int) ([]structs.Problem, error)
}

type ContestsMetadataRepo interface {
	InsertContest(ctx context.Context, contest structs.Contest) (int64, error)
	GetContest(ctx context.Context, id int64) (structs.Contest, error)
	ListContests(ctx context.Context, descending bool, limit, offset int) ([]structs.Contest, error)
	DeleteContest(ctx context.Context, id int64) error
	AddProblem(ctx context.Context, contestID int64, problemID int64) error
}

type ProblemDescriptionsRepo interface {
	Save(description string) (string, error)
	Get(id string) (string, error)
}
type SubmissionMetadataRepo interface {
	Insert(ctx context.Context, submission structs.SubmissionMetadata) (int64, error)
	Get(ctx context.Context, id int64) (structs.SubmissionMetadata, error)
}
