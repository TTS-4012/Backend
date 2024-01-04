package submissions

import (
	"context"
	"fmt"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/internal/judge"
	"github.com/ocontest/backend/internal/minio"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
	"net/http"

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
		err = s.judge.Dispatch(context.Background(), submissionID) // ctx must not be passed to judge, because deadlines are different
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
		ans.Message = "not judged yet!"
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
			Verdicts: nil,
			Message:  "Something Went Wrong!, please try again later...",
		}, http.StatusInternalServerError
	}
	message := `All tests ran successfully`
	verdicts := make([]structs.Verdict, 0)
	for _, t := range judgeResult.TestResults {
		if t.RunnerError != "" {
			message = fmt.Sprintf("Error: %v", t.RunnerError)
		}
		if t.Verdict != structs.VerdictOK {
			message = "Failed, verdict: " + t.Verdict.String()
		}
		verdicts = append(verdicts, t.Verdict)
	}
	if len(verdicts) == 0 {
		message = "There wasn't any test!"
	}

	return structs.ResponseGetSubmissionResults{
		Verdicts: verdicts,
		Message:  message,
	}, http.StatusOK

}

func (s *SubmissionsHandlerImp) ListSubmission(ctx context.Context, req structs.RequestListSubmissions, getAll bool) (structs.ResponseListSubmissions, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "ListSubmission",
		"module": "submissions",
	})

	submissions, total_count, err := s.submissionMetadataRepo.ListSubmissions(ctx, req.ProblemID, req.UserID, req.Descending, req.Limit, req.Offset, req.GetCount)
	if err != nil {
		logger.Error("error on listing problems: ", err)
		return structs.ResponseListSubmissions{}, http.StatusInternalServerError
	}

	ans := make([]structs.ResponseListSubmissionsItem, 0)
	for _, sub := range submissions {
		var uID int64 = 0
		if getAll {
			uID = sub.UserID
		}

		metadata := structs.SubmissionListMetadata{
			ID:        sub.ID,
			UserID:    uID,
			Language:  sub.Language,
			CreatedAt: sub.CreatedAT,
			FileName:  sub.FileName,
		}
		results := structs.ResponseGetSubmissionResults{}

		testResultID := sub.JudgeResultID
		judgeResult, err := s.judge.GetTestResults(ctx, testResultID)
		if err != nil {
			logger.Error("error on getting test results from judge: ", err)
		}

		if judgeResult.ServerError != "" && err != nil {
			results = structs.ResponseGetSubmissionResults{
				Verdicts: nil,
				Message:  "Something Went Wrong!, please try again later...",
			}
		} else {
			message := `All tests ran successfully`
			verdicts := make([]structs.Verdict, 0)
			for _, t := range judgeResult.TestResults {
				if t.RunnerError != "" {
					message = fmt.Sprintf("Error: %v", t.RunnerError)
				}
				if t.Verdict != structs.VerdictOK {
					message = "Failed"
				}
				verdicts = append(verdicts, t.Verdict)
			}

			results = structs.ResponseGetSubmissionResults{
				Verdicts: verdicts,
				Message:  message,
			}
		}

		ans = append(ans, structs.ResponseListSubmissionsItem{
			Metadata: metadata,
			Results:  results,
		})
	}

	return structs.ResponseListSubmissions{
		TotalCount:  total_count,
		Submissions: ans,
	}, http.StatusOK
}
