package contestsProblemspackage

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/pkg"
	"github.com/sirupsen/logrus"
	"net/http"
)

type ContestsProblemsHandler interface {
	AddProblemToContest(ctx context.Context, contestID, problemID int64) (status int)
	GetContestProblems(ctx *gin.Context, contestID int64) ([]int64, int)
}

type ContestsProblemsHandlerImp struct {
	ContestsProblemsRepo db.ContestsProblemsRepo
}

func NewContestsProblemsHandler(contestsProblemsRepo db.ContestsProblemsRepo) ContestsProblemsHandler {
	return &ContestsProblemsHandlerImp{
		ContestsProblemsRepo: contestsProblemsRepo,
	}
}

func (c ContestsProblemsHandlerImp) AddProblemToContest(ctx context.Context, contestID, problemID int64) (status int) {
	logger := pkg.Log.WithField("method", "add_problem_to_contest")

	err := c.ContestsProblemsRepo.AddProblemToContest(ctx, contestID, problemID)
	if err != nil {
		logger.Error("error on adding problem to contest: ", err)
		status = http.StatusInternalServerError
		return
	}

	status = http.StatusOK
	return
}

func (c ContestsProblemsHandlerImp) GetContestProblems(ctx *gin.Context, contestID int64) ([]int64, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "GetContestsProblem",
		"module": "ContestsProblems",
	})

	problems, err := c.ContestsProblemsRepo.GetContestProblems(ctx, contestID)
	if err != nil {
		logger.Error("error on getting contest problems from repo: ", err)
		status := http.StatusInternalServerError
		if errors.Is(err, pkg.ErrNotFound) {
			status = http.StatusNotFound
		}
		return make([]int64, 0), status
	}

	return problems, http.StatusOK
}
