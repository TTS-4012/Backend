package contests

import (
	"context"
	"net/http"
	"ocontest/internal/db"
	"ocontest/pkg"
	"ocontest/pkg/structs"
)

type ContestsHandler interface {
	CreateContest(ctx context.Context, req structs.RequestCreateContest) (res structs.ResponseCreateContest, status int)
	GetContest()
	ListContests()
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
	problem := structs.Contest{
		CreatedBy: ctx.Value("user_id").(int64),
		Title:     req.Title,
		Problems:  nil,
	}
	var err error
	res.ContestID, err = c.ContestsRepo.InsertContest(ctx, problem)
	if err != nil {
		logger.Error("error on inserting contest: ", err)
		status = http.StatusInternalServerError
		return
	}
	status = http.StatusOK
	return
}

func (c ContestsHandlerImp) GetContest()    {}
func (c ContestsHandlerImp) ListContests()  {}
func (c ContestsHandlerImp) UpdateContest() {}
func (c ContestsHandlerImp) DeleteContest() {}
