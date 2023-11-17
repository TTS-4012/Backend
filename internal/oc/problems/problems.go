package problems

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"ocontest/internal/db"
	"ocontest/pkg"
	"ocontest/pkg/structs"
)

type ProblemsHandler interface {
	CreateProblem(ctx context.Context, req structs.RequestCreateProblem) (structs.ResponseCreateProblem, int)
	GetProblem(ctx context.Context, problemID int64) (structs.ResponseGetProblem, int)
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

func (p ProblemsHandlerImp) GetProblem(ctx context.Context, problemID int64) (structs.ResponseGetProblem, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "GetProblem",
		"module": "Problems",
	})

	problem, err := p.problemMetadataRepo.GetProblem(ctx, problemID)
	if err != nil {
		logger.Error("error on getting problem from problem metadata repo: ", err)
		status := http.StatusInternalServerError
		if errors.Is(err, pkg.ErrNotFound) {
			status = http.StatusNotFound
		}
		return structs.ResponseGetProblem{}, status
	}

	doc, err := p.problemsDescriptionRepo.Get(problem.DocumentID)
	if err != nil {
		logger.Error("error on getting problem from problem decription repo: ", err)
		return structs.ResponseGetProblem{}, http.StatusInternalServerError
	}

	return structs.ResponseGetProblem{
		ProblemID:   problemID,
		Title:       problem.Title,
		SolveCount:  problem.SolvedCount,
		Hardness:    problem.Hardness,
		Description: doc,
	}, http.StatusOK
}

func (p ProblemsHandlerImp) ListProblem(ctx context.Context, req structs.RequestListProblems) ([]structs.ResponseListProblemsItem, int) {
	//TODO implement me
	panic("implement me")
}
