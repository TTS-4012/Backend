package contests

import (
	"context"
	"errors"
	"github.com/ocontest/backend/internal/judge"
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
	GetContestScoreboard(ctx context.Context, req structs.RequestGetScoreboard) (structs.ResponseGetContestScoreboard, int)
	UpdateContest()
	DeleteContest()
	AddProblemToContest(ctx context.Context, contestID, problemID int64) (status int)
	GetContestProblems(ctx *gin.Context, contestID int64) ([]int64, int)
	RemoveProblemFromContest(ctx context.Context, contestID, problemID int64) (status int)
}

type ContestsHandlerImp struct {
	usersRepo          db.AuthRepo
	problemsRepo       db.ProblemsMetadataRepo
	submissionsRepo    db.SubmissionMetadataRepo
	contestsRepo       db.ContestsMetadataRepo
	contestProblemRepo db.ContestsProblemsRepo

	judge judge.Judge
}

func NewContestsHandler(
	contestsRepo db.ContestsMetadataRepo, contestProblemRepo db.ContestsProblemsRepo,
	problemsRepo db.ProblemsMetadataRepo, submissionsRepo db.SubmissionMetadataRepo,
	authRepo db.AuthRepo,
	judge judge.Judge,
) ContestsHandler {
	return &ContestsHandlerImp{
		problemsRepo:       problemsRepo,
		submissionsRepo:    submissionsRepo,
		contestsRepo:       contestsRepo,
		contestProblemRepo: contestProblemRepo,
		usersRepo:          authRepo,
		judge:              judge,
	}
}

func (c ContestsHandlerImp) CreateContest(ctx context.Context, req structs.RequestCreateContest) (res structs.ResponseCreateContest, status int) {
	logger := pkg.Log.WithField("method", "create_contest")
	contest := structs.Contest{
		CreatedBy: ctx.Value("user_id").(int64),
		Title:     req.Title,
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

	problemIDs, err := c.contestProblemRepo.GetContestProblems(ctx, contestID)
	if err != nil {
		logger.Error("error on getting contest problems from repo: ", err)
		status := http.StatusInternalServerError
		if errors.Is(err, pkg.ErrNotFound) {
			status = http.StatusNotFound
		}
		return structs.ResponseGetContest{}, status
	}

	problems := make([]structs.ContestProblem, len(problemIDs))
	for i := 0; i < len(problemIDs); i++ {
		title, err := c.problemsRepo.GetProblemTitle(ctx, problemIDs[i])
		if err != nil {
			logger.Error("error on get problem title: ", err)
			status := http.StatusInternalServerError
			if errors.Is(err, pkg.ErrNotFound) {
				status = http.StatusNotFound
			}
			return structs.ResponseGetContest{}, status
		}
		problems[i].ID = problemIDs[i]
		problems[i].Title = title
	}

	return structs.ResponseGetContest{
		ContestID: contestID,
		Title:     contest.Title,
		Problems:  problems,
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

	err := c.contestProblemRepo.AddProblemToContest(ctx, contestID, problemID)

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

	problems, err := c.contestProblemRepo.GetContestProblems(ctx, contestID)
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

	err := c.contestProblemRepo.RemoveProblemFromContest(ctx, contestID, problemID)
	if err != nil {
		logger.Error("error on removing problem from contest: ", err)
		status = http.StatusInternalServerError
		return
	}

	status = http.StatusOK
	return
}

// GetContestScoreboard TODO: when I have time I should do it in a way like I care a shit about pagination performance
func (c ContestsHandlerImp) GetContestScoreboard(ctx context.Context, req structs.RequestGetScoreboard) (ans structs.ResponseGetContestScoreboard, status int) {
	logger := pkg.Log.WithField("method", "get_contest_scoreboard")
	problems, err := c.contestProblemRepo.GetContestProblems(ctx, req.ContestID)
	if err != nil {
		logger.Error("error on get problems from ContestProblem repo: ", err)
		status = http.StatusInternalServerError
		return
	}

	userStandings := make([]structs.ScoreboardUserStanding, 0)
	userScores := make([]int, 0)
	userIDtoStandingIndex := make(map[int64]int)
	for _, p := range problems {
		submissions, err := c.submissionsRepo.GetByProblem(ctx, p)
		if err != nil {
			logger.Error("error on get submissions: ", err)
		}
		pkg.Log.Debug(submissions)
		for _, s := range submissions {
			if _, exists := userIDtoStandingIndex[s.UserID]; !exists {
				userIDtoStandingIndex[s.UserID] = len(userStandings)
				username, err := c.usersRepo.GetUsername(ctx, s.UserID)
				if err != nil {
					logger.Error("error on get username: ", username)
				}

				userStandings = append(userStandings, structs.ScoreboardUserStanding{
					UserID:      s.UserID,
					Username:    username,
					Submissions: make([]structs.ScoreboardCell, 0),
				})
			}

			res, err := c.judge.GetResults(ctx, s.JudgeResultID)
			if err != nil {
				logger.Error("error on get judge result: ", err)
			}
			score := calcScore(res.TestResults)
			var cell structs.ScoreboardCell
			cell.ProblemID = p
			cell.Score = score

			userStandings[userIDtoStandingIndex[s.UserID]].Submissions = append(userStandings[userIDtoStandingIndex[s.UserID]].Submissions, cell)
		}
	}

	for i := range userStandings {
		sumScore := 0
		for _, s := range userStandings[i].Submissions {
			sumScore += s.Score
		}
		userScores[i] = sumScore
	}

	userStandings = assoaciateSort(userScores, userStandings)

	if req.GetCount {
		ans.Count = len(userStandings)
	}
	if req.Limit == -1 {
		req.Limit = len(userStandings)
	}
	if req.Offset == -1 {
		req.Offset = 0
	}
	ans.Users = userStandings[req.Offset:req.Limit]
	status = http.StatusNotImplemented
	return
}
