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
	UpdateProblem(ctx context.Context, id int64, title string) error
	DeleteProblem(ctx context.Context, id int64) (string, error)
}

type ProblemDescriptionsRepo interface {
	Save(description string) (string, error)
	Get(id string) (string, error)
	Update(id string, description string) error
	Delete(id string) error
}
type SubmissionMetadataRepo interface {
	Insert(ctx context.Context, submission structs.SubmissionMetadata) (int64, error)
	Get(ctx context.Context, id int64) (structs.SubmissionMetadata, error)
}
