package submissions

import (
	"context"
	"net/http"
	"ocontest/internal/db"
	"ocontest/internal/minio"
	"ocontest/pkg"
	"ocontest/pkg/structs"

	"github.com/sirupsen/logrus"
)

type Handler interface {
	Submit(ctx context.Context, request structs.RequestSubmit) (submissionID int64, status int)
	Get(ctx context.Context, userID, submissionID int64) (structs.ResponseGetSubmission, string, int)
}

type SubmissionsHandlerImp struct {
	Handler
	submissionMetadataRepo db.SubmissionMetadataRepo
	minioHandler           minio.MinioHandler
}

func NewSubmissionsHandler(submissionRepo db.SubmissionMetadataRepo, minioHandler minio.MinioHandler) Handler {
	return &SubmissionsHandlerImp{
		submissionMetadataRepo: submissionRepo,
		minioHandler:           minioHandler,
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

	objectName := getObjectName(request.UserID, request.ProblemID, submissionID)
	err = s.minioHandler.UploadFile(ctx, request.Code, objectName, request.ContentType)
	if err != nil {
		logger.Error("error on uploading file from minio: ", err)
		return submissionID, http.StatusInternalServerError
	}

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
