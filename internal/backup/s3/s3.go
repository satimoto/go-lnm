package s3

import (
	"bytes"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/satimoto/go-datastore/pkg/util"
)

type S3Backup interface {
	BackupChannels(name string, data []byte) error
	BackupChannelsWithRetry(name string, data []byte, retries int)  error
}

type S3BackupHandler struct {
	session    *session.Session
	bucketName string
}

func NewHandler(region, accessKeyID, secretAccessKey, bucketName string) S3Backup {
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})

	util.PanicOnError("LSP064", "Invalid AWS session", err)

	return &S3BackupHandler{
		session: session,
		bucketName: bucketName,
	}
}

func (h *S3BackupHandler) BackupChannels(name string, data []byte) error {
	return h.BackupChannelsWithRetry(name, data, 0)
}

func (h *S3BackupHandler) BackupChannelsWithRetry(name string, data []byte, retries int) error {
	for i := 0; i <= retries; i++ {
		err := h.processBackupChannels(name, data)

		if err == nil {
			break
		}

		if i == retries {
			util.LogOnError("LSP069", "Error in S3 backup", err)
			log.Printf("LSP069: Name=%v, Retries=%v", name, retries)
			return err
		}

		time.Sleep((time.Duration(i) + 1) * time.Second)
	}

	return nil
}

func (h *S3BackupHandler) processBackupChannels(name string, data []byte) error {
	uploader := s3manager.NewUploader(h.session)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(h.bucketName),
		Key: aws.String(name),
		Body: bytes.NewBuffer(data),
	})

	if err != nil {
		util.LogOnError("LSP070", "Error uploading to S3", err)
		log.Printf("LSP070: Name=%v", name)
		return errors.New("error uploading to S3")
	}

	return nil
}
