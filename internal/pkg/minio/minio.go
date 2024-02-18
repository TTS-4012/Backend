package minio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/configs"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type MinioHandler interface {
	UploadFile(ctx context.Context, file []byte, objectName, contentType string) error
	DownloadFile(ctx context.Context, objectName string) ([]byte, string, error)
	GenCodeObjectname(userID, problemID, submissionID int64) string
}

type MinioHandlerImp struct {
	minioClient *minio.Client
	bucket      string
}

func NewMinioHandler(ctx context.Context, conf configs.SectionMinIO) (MinioHandler, error) {
	endpoint := conf.Endpoint
	accessKeyID := conf.AccessKey
	secretAccessKey := conf.SecretKey
	useTLS := conf.Secure

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useTLS,
	})
	if err != nil {
		return nil, err
	}

	err = createNewBucket(ctx, conf, minioClient)
	if err != nil {
		return nil, err
	}

	return MinioHandlerImp{
		minioClient: minioClient,
		bucket:      conf.Bucket,
	}, nil
}

func createNewBucket(ctx context.Context, conf configs.SectionMinIO, minioClient *minio.Client) error {
	logger := pkg.Log.WithFields(logrus.Fields{
		"module": "minio",
		"method": "new bucket",
	})

	bucketName := conf.Bucket
	location := conf.Region

	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			logger.Info("We already own the bucket ", bucketName)
		} else {
			return err
		}
	} else {
		logger.Info("Successfully created bucket ", bucketName)
	}

	return nil
}

func (f MinioHandlerImp) UploadFile(ctx context.Context, file []byte, objectName, contentType string) error {
	logger := pkg.Log.WithFields(logrus.Fields{
		"module": "minio",
		"method": "upload",
	})

	fileSize := int64(len(file))
	fileReader := bytes.NewReader(file)
	_, err := f.minioClient.PutObject(ctx, f.bucket, objectName, fileReader, fileSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		logger.Error("Failed to store object, error: ", err)
	}
	return err

}

func (f MinioHandlerImp) DownloadFile(ctx context.Context, objectName string) ([]byte, string, error) {
	logger := pkg.Log.WithFields(logrus.Fields{
		"module": "minio",
		"method": "download",
	})

	file, err := f.minioClient.GetObject(ctx, f.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		logger.Error("Failed to get object, error: ", err)
		return nil, "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		logger.Error("error on stat file, error: ", err)
		return nil, "", err
	}

	bytefile, err := io.ReadAll(file)
	if err != nil {
		logger.Error("error on decoding byte file, error: ", err)
		return nil, "", err
	}

	return bytefile, info.ContentType, nil
}

func (f MinioHandlerImp) GenCodeObjectname(userID, problemID, submissionID int64) string {
	return fmt.Sprintf("%d/%d/%d", problemID, userID, submissionID)
}

