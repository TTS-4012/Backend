package contests

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
	"github.com/sirupsen/logrus"
)

type ContestsHandler interface {
	CreateContest(ctx context.Context, req structs.RequestCreateContest) (res structs.ResponseCreateContest, status int)
	GetContest(ctx *gin.Context, contestID int64) (structs.ResponseGetContest, int)
	ListContests(ctx context.Context, req structs.RequestListContests) ([]structs.ResponseListContestsItem, int)
	GetContestScoreboard(ctx context.Context, contestID int64) (structs.ResponseGetContestScoreboard, int)
	UpdateContest()
	DeleteContest()
	AddProblemToContest(ctx context.Context, contestID, problemID int64) (status int)
	GetContestProblems(ctx *gin.Context, contestID int64) ([]int64, int)
	RemoveProblemFromContest(ctx context.Context, contestID, problemID int64) (status int)
}

type ContestsHandlerImp struct {
	contestsRepo         db.ContestsMetadataRepo
	contestsProblemsRepo db.ContestsProblemsRepo
}

func NewContestsHandler(contestsRepo db.ContestsMetadataRepo, contestProblemsRepo db.ContestsProblemsRepo) ContestsHandler {
	return &ContestsHandlerImp{
		contestsRepo:         contestsRepo,
		contestsProblemsRepo: contestProblemsRepo,
	}
}

func (c ContestsHandlerImp) CreateContest(ctx context.Context, req structs.RequestCreateContest) (res structs.ResponseCreateContest, status int) {
	logger := pkg.Log.WithField("method", "create_contest")
	contest := structs.Contest{
		CreatedBy: ctx.Value("user_id").(int64),
		Title:     req.Title,
		Problems:  nil,
		StartTime: req.StartTime,
		Duration:  req.Duration,
	}
	var err error
	res.ContestID, err = c.contestsRepo.InsertContest(ctx, contest)
	if err != nil {
		logger.Error("error on inserting contest: ", err)
		status = http.StatusInternalServerError
		return
	}
	status = http.StatusOK
	return
}

func (c ContestsHandlerImp) GetContest(ctx *gin.Context, contestID int64) (structs.ResponseGetContest, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "GetContest",
		"module": "Contests",
	})

	contest, err := c.contestsRepo.GetContest(ctx, contestID)
	if err != nil {
		logger.Error("error on getting contest from repo: ", err)
		status := http.StatusInternalServerError
		if errors.Is(err, pkg.ErrNotFound) {
			status = http.StatusNotFound
		}
		return structs.ResponseGetContest{}, status
	}

	return structs.ResponseGetContest{
		ContestID: contestID,
		Title:     contest.Title,
		Problems:  contest.Problems,
		StartTime: contest.StartTime,
		Duration:  contest.Duration,
	}, http.StatusOK
}

func (c ContestsHandlerImp) ListContests(ctx context.Context, req structs.RequestListContests) ([]structs.ResponseListContestsItem, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "ListContests",
		"module": "Contests",
	})
	contests, err := c.contestsRepo.ListContests(ctx, req.Descending, req.Limit, req.Offset, req.Started)
	if err != nil {
		logger.Error("error on listing contests: ", err)
		return nil, http.StatusInternalServerError
	}

	res := make([]structs.ResponseListContestsItem, 0)
	for _, contest := range contests {
		res = append(res, structs.ResponseListContestsItem{
			ContestID: contest.ID,
			Title:     contest.Title,
		})
	}
	return res, http.StatusOK
}

func (c ContestsHandlerImp) UpdateContest() {}
func (c ContestsHandlerImp) DeleteContest() {}

func (c ContestsHandlerImp) AddProblemToContest(ctx context.Context, contestID, problemID int64) (status int) {
	logger := pkg.Log.WithField("method", "add_problem_to_contest")

	err := c.contestsProblemsRepo.AddProblemToContest(ctx, contestID, problemID)
	if err != nil {
		logger.Error("error on adding problem to contest: ", err)
		status = http.StatusInternalServerError
		return
	}

	status = http.StatusOK
	return
}

func (c ContestsHandlerImp) GetContestProblems(ctx *gin.Context, contestID int64) ([]int64, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "GetContestsProblem",
		"module": "ContestsProblems",
	})

	problems, err := c.contestsProblemsRepo.GetContestProblems(ctx, contestID)
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

func (c ContestsHandlerImp) RemoveProblemFromContest(ctx context.Context, contestID, problemID int64) (status int) {
	logger := pkg.Log.WithField("method", "remove_problem_from_contest")

	err := c.contestsProblemsRepo.RemoveProblemFromContest(ctx, contestID, problemID)
	if err != nil {
		logger.Error("error on removing problem from contest: ", err)
		status = http.StatusInternalServerError
		return
	}

	status = http.StatusOK
	return
}

func (c ContestsHandlerImp) GetContestScoreboard(ctx context.Context, contestID int64) (ans structs.ResponseGetContestScoreboard, status int) {
	// logger := pkg.Log.WithField("method", "get_contest_scoreboard")
	status = http.StatusNotImplemented
	return
}
