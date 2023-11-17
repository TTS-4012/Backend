package problems

import (
	"context"
	"net/http"
	"ocontest/internal/db"
	"ocontest/pkg"
	"ocontest/pkg/structs"
)

type ProblemsHandler interface {
	CreateProblem(ctx context.Context, req structs.RequestCreateProblem) (structs.ResponseCreateProblem, int)
	GetProblem(ctx context.Context, problemID int) (structs.ResponseGetProblem, int)
	ListProblem(ctx context.Context, req structs.RequestListProblems) ([]structs.ResponseListProblemsItem, int)
}

type ProblemsHandlerImp struct {
	problemMetadataRepo     db.ProblemsMetadataRepo
	problemsDescriptionRepo db.ProblemDescriptionsRepo
}

func NewProblemsHandler(problemsRepo db.ProblemsMetadataRepo, problemsDescriptionRepo db.ProblemDescriptionsRepo) ProblemsHandler {
	return &ProblemsHandlerImp{
		problemMetadataRepo:     problemsRepo,
		problemsDescriptionRepo: problemsDescriptionRepo,
	}
}

func (p ProblemsHandlerImp) CreateProblem(ctx context.Context, req structs.RequestCreateProblem) (ans structs.ResponseCreateProblem, status int) {
	logger := pkg.Log.WithField("method", "create_problem")
	docID, err := p.problemsDescriptionRepo.Save(req.Description)
	if err != nil {
		logger.Error("error on inserting problem description: ", err)
		status = http.StatusInternalServerError
		return
	}
	problem := structs.Problem{
		Title:      req.Title,
		DocumentID: docID,
		CreatedBy:  ctx.Value("user_id").(int64),
	}
	ans.ProblemID, err = p.problemMetadataRepo.InsertProblem(ctx, problem)
	if err != nil {
		logger.Error("error on inserting problem metadata: ", err)
		status = http.StatusInternalServerError
		return
	}
	status = http.StatusOK
	return
}

func (p ProblemsHandlerImp) GetProblem(ctx context.Context, problemID int) (structs.ResponseGetProblem, int) {
	//TODO implement me
	panic("implement me")
}

func (p ProblemsHandlerImp) ListProblem(ctx context.Context, req structs.RequestListProblems) ([]structs.ResponseListProblemsItem, int) {
	//TODO implement me
	panic("implement me")
}
