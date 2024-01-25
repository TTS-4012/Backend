package submissions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ocontest/backend/internal/db"
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
	ListSubmission(ctx context.Context, req structs.RequestListSubmissions, getAll bool) (structs.ResponseListSubmissions, int)
}

type SubmissionsHandlerImp struct {
	submissionMetadataRepo db.SubmissionMetadataRepo
	minioHandler           minio.MinioHandler
	judge                  judge.Judge
}

func NewSubmissionsHandler(submissionRepo db.SubmissionMetadataRepo, minioHandler minio.MinioHandler, judgeHandler judge.Judge) Handler {
	return &SubmissionsHandlerImp{
		submissionMetadataRepo: submissionRepo,
		minioHandler:           minioHandler,
		judge:                  judgeHandler,
	}
}

func (s *SubmissionsHandlerImp) Submit(ctx context.Context, request structs.RequestSubmit) (submissionID int64, status int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "Submit",
		"module": "Submission",
	})

	status = http.StatusInternalServerError

	submission := structs.SubmissionMetadata{
		UserID:    request.UserID,
		ProblemID: request.ProblemID,
		FileName:  request.FileName,
		Language:  request.Language,
	}

	submissionID, err := s.submissionMetadataRepo.Insert(ctx, submission)
	if err != nil {
		logger.Error("error on insert to submission repo: ", err)
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

func (s *SubmissionsHandlerImp) ListSubmission(ctx context.Context, req structs.RequestListSubmissions, getAll bool) (structs.ResponseListSubmissions, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "ListSubmission",
		"module": "submissions",
	})

	submissions, totalCount, err := s.submissionMetadataRepo.ListSubmissions(ctx, req.ProblemID, req.UserID, req.Descending, req.Limit, req.Offset, req.GetCount)
	if err != nil {
		logger.Error("error on listing problems: ", err)
		return structs.ResponseListSubmissions{}, http.StatusInternalServerError
	}

	ans := structs.ResponseListSubmissions{
		Submissions: make([]structs.ResponseListSubmissionsItem, 0),
	}

	for _, sub := range submissions {

		results, status := s.GetResults(ctx, sub.ID)
		if err != nil {
			logger.WithError(err).Error("coudn't get submission result")
			return structs.ResponseListSubmissions{}, status
		}

		meta := structs.SubmissionListMetadata{
			ID:        sub.ID,
			UserID:    sub.UserID,
			Language:  sub.Language,
			CreatedAt: sub.CreatedAT,
			FileName:  sub.FileName,
		}

		ans.Submissions = append(ans.Submissions, structs.ResponseListSubmissionsItem{
			Metadata: meta,
			Results:  results,
		})
	}

	ans.TotalCount = totalCount
	return ans, totalCount
}
