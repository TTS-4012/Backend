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
	AddProblemContest(ctx *gin.Context, req structs.RequestAddProblemContest) int
}

type ContestsHandlerImp struct {
	ContestsRepo       db.ContestsMetadataRepo
	ContestProblemRepo db.ContestsProblemsRepo
}

func NewContestsHandler(contestsRepo db.ContestsMetadataRepo, contestProblemRepo db.ContestsProblemsRepo) ContestsHandler {
	return &ContestsHandlerImp{
		ContestsRepo:       contestsRepo,
		ContestProblemRepo: contestProblemRepo,
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
	res.ContestID, err = c.ContestsRepo.InsertContest(ctx, contest)
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

	contest, err := c.ContestsRepo.GetContest(ctx, contestID)
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
	contests, err := c.ContestsRepo.ListContests(ctx, req.Descending, req.Limit, req.Offset)
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

func (c ContestsHandlerImp) AddProblemContest(ctx *gin.Context, req structs.RequestAddProblemContest) int {
	logger := pkg.Log.WithField("method", "add_problem_contest")

	err := c.ContestProblemRepo.AddProblem(ctx, req.ContestID, req.ProblemID)
	if err != nil {
		logger.Error("error on adding problem to contest: ", err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (c ContestsHandlerImp) GetContestScoreboard(ctx context.Context, contestID int64) (ans structs.ResponseGetContestScoreboard, status int) {
	// logger := pkg.Log.WithField("method", "get_contest_scoreboard")
	status = http.StatusNotImplemented
	return
}
