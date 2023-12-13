package contests

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"ocontest/internal/db"
	"ocontest/pkg"
	"ocontest/pkg/structs"
)

type ContestsHandler interface {
	CreateContest(ctx context.Context, req structs.RequestCreateContest) (res structs.ResponseCreateContest, status int)
	GetContest(ctx *gin.Context, contestID int64) (structs.ResponseGetContest, int)
	ListContests(ctx context.Context, req structs.RequestListContests) ([]structs.ResponseListContestsItem, int)
	UpdateContest()
	DeleteContest()
}

type ContestsHandlerImp struct {
	ContestsRepo db.ContestsMetadataRepo
}

func NewContestsHandler(contestsRepo db.ContestsMetadataRepo) ContestsHandler {
	return &ContestsHandlerImp{
		ContestsRepo: contestsRepo,
	}
}

func (c ContestsHandlerImp) CreateContest(ctx context.Context, req structs.RequestCreateContest) (res structs.ResponseCreateContest, status int) {
	logger := pkg.Log.WithField("method", "create_contest")
	contest := structs.Contest{
		//CreatedBy: ctx.Value("user_id").(int64),
		CreatedBy: (int64)(1),
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
