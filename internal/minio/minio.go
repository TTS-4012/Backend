package minio

import (
	"context"
	"ocontest/pkg"
	"ocontest/pkg/configs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FilesHandler interface {
	UploadFile()
	DownloadFile()
}

type FilesHandlerImp struct {
	minioClient *minio.Client
	bucket      string
}

func NewFilesHandler(minioClient *minio.Client, bucket string) FilesHandler {
	return FilesHandlerImp{
		minioClient: minioClient,
		bucket:      bucket,
	}
}

func GetNewClient(ctx context.Context, conf configs.SectionMinIO) (*minio.Client, error) {
	endpoint := conf.Endpoint
	accessKeyID := conf.AccessKey
	secretAccessKey := conf.SecretKey
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}

func CreateNewBucket(ctx context.Context, conf configs.SectionMinIO, minioClient *minio.Client) error {
	logger := pkg.Log.WithField("method", "CreateNewBucket")
	bucketName := conf.Bucket
	location := conf.Region

	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			logger.Warn("We already own ", bucketName)
		} else {
			return err
		}
	} else {
		logger.Warn("Successfully created ", bucketName)
	}

	return nil
}

func (f FilesHandlerImp) UploadFile() {
	//TODO
}

func (f FilesHandlerImp) DownloadFile() {
	//TODO
}
