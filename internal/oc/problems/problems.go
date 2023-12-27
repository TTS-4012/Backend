package problems

import (
	"context"
	"errors"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
	"net/http"

	"github.com/sirupsen/logrus"
)

type ProblemsHandler interface {
	CreateProblem(ctx context.Context, req structs.RequestCreateProblem) (structs.ResponseCreateProblem, int)
	GetProblem(ctx context.Context, problemID int64) (structs.ResponseGetProblem, int)
	ListProblem(ctx context.Context, req structs.RequestListProblems) (structs.ResponseListProblems, int)
	DeleteProblem(ctx context.Context, problemId int64) int
	UpdateProblem(ctx context.Context, req structs.RequestUpdateProblem) int
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
	docID, err := p.problemsDescriptionRepo.Insert(req.Description, nil)
	if err != nil {
		logger.Error("error on inserting problem description: ", err)
		status = http.StatusInternalServerError
		return
	}
	problem := structs.Problem{
		Title:      req.Title,
		DocumentID: docID,
		CreatedBy:  ctx.Value("user_id").(int64),
		IsPrivate:  req.IsPrivate,
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
		Description: doc.Description,
		IsOwned:     problem.CreatedBy == ctx.Value("user_id").(int64),
	}, http.StatusOK
}

func (p ProblemsHandlerImp) ListProblem(ctx context.Context, req structs.RequestListProblems) (structs.ResponseListProblems, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "ListProblem",
		"module": "Problems",
	})
	problems, total_count, err := p.problemMetadataRepo.ListProblems(ctx, req.OrderedBy, req.Descending, req.Limit, req.Offset, req.GetCount)
	if err != nil {
		logger.Error("error on listing problems: ", err)
		return structs.ResponseListProblems{}, http.StatusInternalServerError
	}

	ans := make([]structs.ResponseListProblemsItem, 0)
	for _, p := range problems {
		ans = append(ans, structs.ResponseListProblemsItem{
			ProblemID:  p.ID,
			Title:      p.Title,
			SolveCount: p.SolvedCount,
			Hardness:   p.Hardness,
		})
	}
	return structs.ResponseListProblems{
		TotalCount: total_count,
		Problems:   ans,
	}, http.StatusOK
}

func (p ProblemsHandlerImp) UpdateProblem(ctx context.Context, req structs.RequestUpdateProblem) int {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "DeleteProblem",
		"module": "Problems",
	})

	if req.Title != "" {
		err := p.problemMetadataRepo.UpdateProblem(ctx, req.Id, req.Title)
		if err != nil {
			logger.Error("error on updating problem on problem metadata repo: ", err)
			status := http.StatusInternalServerError
			if errors.Is(err, pkg.ErrNotFound) {
				status = http.StatusNotFound
			}
			return status
		}
	}

	if req.Description != "" {
		problem, err := p.problemMetadataRepo.GetProblem(ctx, req.Id)
		if err != nil {
			logger.Error("error on getting problem from problem metadata repo: ", err)
			status := http.StatusInternalServerError
			if errors.Is(err, pkg.ErrNotFound) {
				status = http.StatusNotFound
			}
			return status
		}
		err = p.problemsDescriptionRepo.Update(problem.DocumentID, req.Description)
		if err != nil {
			logger.Error("error on updating problem description: ", err)
			status := http.StatusInternalServerError
			return status
		}
	}

	return http.StatusAccepted
}

func (p ProblemsHandlerImp) DeleteProblem(ctx context.Context, problemID int64) int {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "DeleteProblem",
		"module": "Problems",
	})

	documentID, err := p.problemMetadataRepo.DeleteProblem(ctx, problemID)
	if err != nil {
		logger.Error("error on getting problem from problem metadata repo: ", err)
		status := http.StatusInternalServerError
		if errors.Is(err, pkg.ErrNotFound) {
			status = http.StatusNotFound
		}
		return status
	}

	err = p.problemsDescriptionRepo.Delete(documentID)
	if err != nil {
		logger.Error("error on getting problem from problem decription repo: ", err)
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}
