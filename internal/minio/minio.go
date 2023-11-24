package minio

import (
	"context"
	"fmt"
	"io"
	"log"
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

func NewFilesHandler(ctx context.Context, conf configs.SectionMinIO, minioClient *minio.Client) FilesHandler {
	err := createNewBucket(ctx, conf, minioClient)
	if err != nil {
		log.Fatal("error on creating new minio bucket", err)
	}

	return FilesHandlerImp{
		minioClient: minioClient,
		bucket:      conf.Bucket,
	}
}

func GetNewClient(ctx context.Context, conf configs.SectionMinIO) (*minio.Client, error) {
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

	return minioClient, nil
}

func createNewBucket(ctx context.Context, conf configs.SectionMinIO, minioClient *minio.Client) error {
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
	logger := pkg.Log.WithField("method", "DownloadFile")

	file, err := f.minioClient.GetObject(ctx, f.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		logger.Error("Failed to get object", err)
		return structs.ResponseDownloadFile{}, http.StatusInternalServerError
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		logger.Error("Failed to get object info", err)
		return structs.ResponseDownloadFile{}, http.StatusInternalServerError
	}

	bytefile := make([]byte, info.Size)
	_, err = file.Read(bytefile)
	if err != io.EOF {
		logger.Error("Failed to read file", err)
		return structs.ResponseDownloadFile{}, http.StatusInternalServerError
	}

	logger.Info("Successfully downloaded ", objectName, " of size ", info.Size)
	return structs.ResponseDownloadFile{
		ContentDisposition: fmt.Sprintf("attachment; filename=%s", objectName),
		ContentType:        info.ContentType,
		Data:               bytefile,
	}, http.StatusOK
}
