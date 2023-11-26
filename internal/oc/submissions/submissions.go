package submissions

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
	"ocontest/internal/db"
	"ocontest/internal/minio"
	"ocontest/pkg"
	"ocontest/pkg/structs"
)

type Handler interface {
	Submit(ctx context.Context, userID, problemID int64, code []byte, contentType string) (submissionID int64, status int)
	Get(ctx context.Context, userID, submissionID int64) (structs.ResponseGetSubmission, string, int)
}

type SubmissionHandlerImp struct {
	Handler
	submissionMetadataRepo db.SubmissionMetadataRepo
	minioHandler           minio.MinioHandler
}

func NewSubmissionHandler(submissionRepo db.SubmissionMetadataRepo, minioHandler minio.MinioHandler) Handler {
	return &SubmissionHandlerImp{
		submissionMetadataRepo: submissionRepo,
		minioHandler:           minioHandler,
	}
}

func (s *SubmissionHandlerImp) Submit(ctx context.Context, userID, problemID int64, code []byte, contentType string) (submissionID int64, status int) {
	logger := pkg.Log.WithField("handler", "Submit")

	status = http.StatusInternalServerError

	submission := structs.SubmissionMetadata{
		UserID:    userID,
		ProblemID: problemID,
	}
	submissionID, err := s.submissionMetadataRepo.Insert(ctx, submission)
	if err != nil {
		logger.Error("error on insert to submission repo: ", err)
		return
	}

	objectName := getObjectName(userID, problemID, submissionID)
	err = s.minioHandler.UploadFile(ctx, code, objectName, contentType)
	if err != nil {
		return submissionID, http.StatusInternalServerError
	}
	return submissionID, http.StatusOK
}

func (s *SubmissionHandlerImp) Get(ctx context.Context, userID, submissionID int64) (ans structs.ResponseGetSubmission, contentType string, status int) {
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
