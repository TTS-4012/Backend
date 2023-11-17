package problems

import (
	"context"
	"ocontest/pkg/structs"
)

type ProblemsHandler interface {
	CreateProblem(ctx context.Context, req structs.RequestCreateProblem) (structs.ResponseCreateProblem, int)
	GetProblem(ctx context.Context, problemID int) (structs.ResponseGetProblem, int)
	ListProblem(ctx context.Context, req structs.RequestListProblems) ([]structs.ResponseListProblemsItem, int)
}

type ProblemsHandlerImp struct {
}

func (p ProblemsHandlerImp) ListProblem(ctx context.Context, req structs.RequestListProblems) ([]structs.ResponseListProblemsItem, int) {
	//TODO implement me
	panic("implement me")
}

func NewProblemsHandler() ProblemsHandler {
	return &ProblemsHandlerImp{}
}

func (p ProblemsHandlerImp) CreateProblem(ctx context.Context, req structs.RequestCreateProblem) (structs.ResponseCreateProblem, int) {
	//TODO implement me
	panic("implement me")
}

func (p ProblemsHandlerImp) GetProblem(ctx context.Context, problemID int) (structs.ResponseGetProblem, int) {
	//TODO implement me
	panic("implement me")
}
