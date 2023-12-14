package submissions

import (
	"context"
	"net/http"
	"ocontest/internal/db"
	"ocontest/internal/judge"
	"ocontest/internal/minio"
	"ocontest/pkg"
	"ocontest/pkg/structs"

	"github.com/sirupsen/logrus"
)

type Handler interface {
	Submit(ctx context.Context, request structs.RequestSubmit) (submissionID int64, status int)
	Get(ctx context.Context, userID, submissionID int64) (structs.ResponseGetSubmission, string, int)
	GetResults(ctx context.Context, submissionID int64) (structs.ResponseGetSubmissionResults, int)
	ListSubmission(ctx context.Context, req structs.RequestListSubmissions) (structs.ResponseListSubmissions, int)
	ListAllSubmission(ctx context.Context, req structs.RequestListSubmissions) (structs.ResponseListSubmissions, int)
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

	go s.judge.Dispatch(ctx, submissionID)
	return submissionID, http.StatusOK
}

func (s *SubmissionsHandlerImp) Get(ctx context.Context, userID, submissionID int64) (ans structs.ResponseGetSubmission, contentType string, status int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "Get",
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

	testResultID := submission.JudgeResultID
	testResults, err := s.judge.GetResults(ctx, testResultID)
	if err != nil {
		logger.Error("error on getting test results from judge: ", err)
		status = http.StatusInternalServerError
		return
	}

	if testResults.ServerError != "" {
		return structs.ResponseGetSubmissionResults{
			TestStates: nil,
			Message:    "Something Went Wrong!, please try again later...",
			Score:      0,
		}, http.StatusInternalServerError
	}
	return structs.ResponseGetSubmissionResults{
		TestStates: testResults.TestStates,
		Message:    testResults.UserError,
		Score:      calcScore(testResults.TestStates, testResults.UserError),
	}, http.StatusOK

}

func (s *SubmissionsHandlerImp) ListSubmission(ctx context.Context, req structs.RequestListSubmissions) (structs.ResponseListSubmissions, int) {
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
		metadata := structs.SubmissionListMetadata{
			ID:        sub.ID,
			UserID:    0,
			Language:  sub.Language,
			CreatedAt: sub.CreatedAT,
			FileName:  sub.FileName,
		}
		results := structs.ResponseGetSubmissionResults{}

		testResults, err := s.judge.GetResults(ctx, sub.JudgeResultID)
		if err != nil {
			logger.Error("error on getting test results from judge: ", err)
		}

		if testResults.ServerError != "" && err != nil {
			results = structs.ResponseGetSubmissionResults{
				TestStates: nil,
				Message:    "Something Went Wrong!, please try again later...",
				Score:      0,
			}
		} else {
			results = structs.ResponseGetSubmissionResults{
				TestStates: testResults.TestStates,
				Message:    testResults.UserError,
				Score:      calcScore(testResults.TestStates, testResults.UserError),
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

func (s *SubmissionsHandlerImp) ListAllSubmission(ctx context.Context, req structs.RequestListSubmissions) (structs.ResponseListSubmissions, int) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"method": "ListAllSubmission",
		"module": "submissions",
	})

	submissions, total_count, err := s.submissionMetadataRepo.ListSubmissions(ctx, req.ProblemID, req.UserID, req.Descending, req.Limit, req.Offset, req.GetCount)
	if err != nil {
		logger.Error("error on listing problems: ", err)
		return structs.ResponseListSubmissions{}, http.StatusInternalServerError
	}

	ans := make([]structs.ResponseListSubmissionsItem, 0)
	for _, sub := range submissions {
		metadata := structs.SubmissionListMetadata{
			ID:        sub.ID,
			UserID:    sub.UserID,
			Language:  sub.Language,
			CreatedAt: sub.CreatedAT,
			FileName:  sub.FileName,
		}
		results := structs.ResponseGetSubmissionResults{}

		testResults, err := s.judge.GetResults(ctx, sub.JudgeResultID)
		if err != nil {
			logger.Error("error on getting test results from judge: ", err)
		}

		if testResults.ServerError != "" && err != nil {
			results = structs.ResponseGetSubmissionResults{
				TestStates: nil,
				Message:    "Something Went Wrong!, please try again later...",
				Score:      0,
			}
		} else {
			results = structs.ResponseGetSubmissionResults{
				TestStates: testResults.TestStates,
				Message:    testResults.UserError,
				Score:      calcScore(testResults.TestStates, testResults.UserError),
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
