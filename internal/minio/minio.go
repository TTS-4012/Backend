package minio

import (
	"context"
	"mime/multipart"
	"net/http"
	"ocontest/pkg"
	"ocontest/pkg/configs"
	"ocontest/pkg/structs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FilesHandler interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader) (structs.ResponseUploadFile, int)
	DownloadFile(ctx context.Context, objectName string) (structs.ResponseDownloadFile, int)
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
		logger.Info("Successfully created bucket ", bucketName)
	}

	return nil
}

func (f FilesHandlerImp) UploadFile(ctx context.Context, file *multipart.FileHeader) (structs.ResponseUploadFile, int) {
	logger := pkg.Log.WithField("method", "UploadFile")

	buffer, err := file.Open()
	if err != nil {
		logger.Error("Failed to open file", err)
		return structs.ResponseUploadFile{}, http.StatusInternalServerError
	}
	defer buffer.Close()

	objectName := file.Filename
	fileBuffer := buffer
	contentType := file.Header["Content-Type"][0]
	fileSize := file.Size
	info, err := f.minioClient.PutObject(ctx, f.bucket, objectName, fileBuffer, fileSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		logger.Error("Failed to store object", err)
		return structs.ResponseUploadFile{}, http.StatusInternalServerError
	}

	logger.Info("Successfully uploaded ", objectName, " of size ", info.Size)
	return structs.ResponseUploadFile{
		FileName: objectName,
	}, http.StatusOK
}

func (f FilesHandlerImp) DownloadFile(ctx context.Context, objectName string) (structs.ResponseDownloadFile, int) {
	//TODO

	return structs.ResponseDownloadFile{}, http.StatusNotImplemented
}
