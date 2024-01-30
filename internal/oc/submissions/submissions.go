package submissions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ocontest/backend/internal/db/repos"
	"github.com/ocontest/backend/internal/judge"
	"github.com/ocontest/backend/internal/minio"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"

	"github.com/sirupsen/logrus"
)

type Handler interface {
	Submit(ctx context.Context, request structs.RequestSubmit) (submissionID int64, status int)
	Get(ctx context.Context, userID, submissionID int64) (structs.ResponseGetSubmission, string, int)
	GetResults(ctx context.Context, submissionID int64) (structs.ResponseGetSubmissionResults, int)
	ListSubmission(ctx context.Context, req structs.RequestListSubmissions) (structs.ResponseListSubmissions, int)
}

type SubmissionsHandlerImp struct {
	submissionMetadataRepo repos.SubmissionMetadataRepo
	minioHandler           minio.MinioHandler
	judge                  judge.Judge
	contestsUsersRepo      repos.ContestsUsersRepo
	contestsMetadataRepo   repos.ContestsMetadataRepo
	contestsProblemsRepo   repos.ContestsProblemsRepo
}

func NewSubmissionsHandler(submissionRepo repos.SubmissionMetadataRepo, contestRepo repos.ContestsMetadataRepo, contestsProblemsRepo repos.ContestsProblemsRepo, contestsUsersRepo repos.ContestsUsersRepo, minioHandler minio.MinioHandler, judgeHandler judge.Judge) Handler {
	return &SubmissionsHandlerImp{
		submissionMetadataRepo: submissionRepo,
		minioHandler:           minioHandler,
		judge:                  judgeHandler,
		contestsUsersRepo:      contestsUsersRepo,
		contestsMetadataRepo:   contestRepo,
		contestsProblemsRepo:   contestsProblemsRepo,
	}
}

func (s *SubmissionsHandlerImp) Submit(ctx context.Context, request structs.RequestSubmit) (submissionID int64, status int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "Submit",
		"module": "Submission",
	})

	status = http.StatusInternalServerError

	if request.ContestID != 0 {
		isReg, err := s.contestsUsersRepo.IsRegistered(ctx, request.ContestID, request.UserID)
		if err != nil {
			logger.Error("error on check to contests users repo: ", err)
			return
		}
		if !isReg {
			logger.Warningf("forbidden contest submit, user id: %v, contest id: %v", request.UserID, request.ContestID)
			status = http.StatusForbidden
			return
		}

		started, err := s.contestsMetadataRepo.HasStarted(ctx, request.ContestID)
		if err != nil {
			logger.Error("error on check to contests metadata repo: ", err)
			return
		}
		if !started {
			logger.Warningf("early contest submit, user id: %v, contest id: %v", request.UserID, request.ContestID)
			status = http.StatusForbidden
			return
		}

		validProblem, err := s.contestsProblemsRepo.HasProblem(ctx, request.ContestID, request.ProblemID)
		if err != nil {
			logger.Error("error on check to contests problems repo: ", err)
			return
		}
		if !validProblem {
			logger.Warningf("not valid contest problem submit, user id: %v, problem id: %v, contest id: %v", request.UserID, request.ProblemID, request.ContestID)
			status = http.StatusNotFound
			return
		}
	}

	submission := structs.SubmissionMetadata{
		UserID:    request.UserID,
		ProblemID: request.ProblemID,
		FileName:  request.FileName,
		Language:  request.Language,
		ContestID: request.ContestID,
	}

	submissionID, err := s.submissionMetadataRepo.Insert(ctx, submission)
	if err != nil {
		logger.Error("error on insert to submission repos: ", err)
		return
	}

	objectName := s.minioHandler.GenCodeObjectname(request.UserID, request.ProblemID, submissionID)
	err = s.minioHandler.UploadFile(ctx, request.Code, objectName, request.ContentType)
	if err != nil {
		logger.Error("error on uploading file from minio: ", err)
		return submissionID, http.StatusInternalServerError
	}

	go func() {
		err = s.judge.Dispatch(context.Background(), submissionID, request.ContestID) // ctx must not be passed to judge, because deadlines are different
		if err != nil {
			logger.Error("error on dispatching judge: ", err)
			return
		}
	}()

	return submissionID, http.StatusOK
}

func (s *SubmissionsHandlerImp) Get(ctx context.Context, userID, submissionID int64) (ans structs.ResponseGetSubmission, contentType string, status int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "GetByID",
		"module": "Submission",
	})

	status = http.StatusInternalServerError
	submissionMetadata, err := s.submissionMetadataRepo.Get(ctx, submissionID)
	if err != nil {
		if err == pkg.ErrNotFound {
			status = http.StatusNotFound
		}
		logger.Error("error on get submission, error: ", err)
		return
	}
	if !submissionMetadata.Public && submissionMetadata.UserID != userID {
		logger.Warningf("forbidden submission download, user id: %v, owner id: %v", userID, submissionMetadata.UserID)
		status = http.StatusForbidden
		return
	}

	objectName := getObjectName(submissionMetadata.UserID, submissionMetadata.ProblemID, submissionMetadata.ID)
	object, contentType, err := s.minioHandler.DownloadFile(ctx, objectName)
	if err != nil {
		logger.Error("error on get file from minio: ", err)
		return
	}

	status = http.StatusOK
	ans = structs.ResponseGetSubmission{
		Metadata: submissionMetadata,
		RawCode:  object,
	}

	return
}

func (s *SubmissionsHandlerImp) GetResults(ctx context.Context, submissionID int64) (ans structs.ResponseGetSubmissionResults, status int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "GetResult",
		"module": "Submissions",
	})

	submission, err := s.submissionMetadataRepo.Get(ctx, submissionID)
	if err != nil {
		logger.Error("error on getting submission from db:", err)
		status = http.StatusInternalServerError
		return
	}

	if submission.Status != "processed" {
		ans.Verdicts = nil
		ans.ServiceMessage = "not judged yet!"
		status = http.StatusTooEarly
		return
	}

	testResultID := submission.JudgeResultID
	judgeResult, err := s.judge.GetTestResults(ctx, testResultID)
	if err != nil {
		logger.Error("error on getting test results from judge: ", err)
		status = http.StatusInternalServerError
		return
	}

	if judgeResult.ServerError != "" {
		return structs.ResponseGetSubmissionResults{
			Verdicts:       nil,
			ServiceMessage: "Something Went Wrong!, please try again later...",
		}, http.StatusInternalServerError
	}
	ans = structs.ResponseGetSubmissionResults{
		Verdicts:       make([]structs.Verdict, 0),
		ServiceMessage: `All tests ran successfully`,
	}
	isFailed := false

	for _, t := range judgeResult.TestResults {
		if !isFailed && t.Verdict != structs.VerdictOK {
			ans.ServiceMessage = fmt.Sprintf("Failed on testcase with id %d, verdict status: %v", t.TestcaseID, t.Verdict.String())
			ans.ErrorMessage = t.RunnerError
			isFailed = true
		}
		ans.Verdicts = append(ans.Verdicts, t.Verdict)
	}
	if len(ans.Verdicts) == 0 {
		ans.ServiceMessage = "There wasn't any test!"
	}

	status = http.StatusOK
	return
}

func (s *SubmissionsHandlerImp) ListSubmission(ctx context.Context, req structs.RequestListSubmissions) (structs.ResponseListSubmissions, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "ListSubmission",
		"module": "submissions",
	})

	submissions, totalCount, err := s.submissionMetadataRepo.ListSubmissions(ctx, req.ProblemID, req.UserID, req.ContestID, req.Descending, req.Limit, req.Offset, req.GetCount)
	if err != nil {
		logger.Error("error on listing submissions: ", err)
		return structs.ResponseListSubmissions{}, http.StatusInternalServerError
	}

	ans := structs.ResponseListSubmissions{
		Submissions: make([]structs.ResponseListSubmissionsItem, 0),
	}

	for _, sub := range submissions {

		results, _ := s.GetResults(ctx, sub.ID)

		meta := structs.SubmissionListMetadata{
			ID:           sub.ID,
			UserID:       sub.UserID,
			Language:     sub.Language,
			CreatedAt:    sub.CreatedAT,
			FileName:     sub.FileName,
			Score:        sub.Score,
			ProblemTitle: sub.ProblemTitle,
			ProblemID:    sub.ProblemID,
		}

		ans.Submissions = append(ans.Submissions, structs.ResponseListSubmissionsItem{
			Metadata: meta,
			Results:  results,
		})
	}

	ans.TotalCount = totalCount
	return ans, http.StatusOK
}
