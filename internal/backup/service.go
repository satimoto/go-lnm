package backup

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/satimoto/go-lnm/internal/backup/s3"
)

type Backup interface {
	BackupChannels(data []byte)
	BackupChannelsWithRetry(data []byte, retries int)
}

type BackupService struct {
	S3Backup   s3.S3Backup
}

func NewService() Backup {
	backupAwsRegion := os.Getenv("BACKUP_AWS_REGION")
	backupAwsAccessKeyID := os.Getenv("BACKUP_AWS_ACCESS_KEY_ID")
	backupAwsSecretAccessKey := os.Getenv("BACKUP_AWS_SECRET_ACCESS_KEY")
	backupS3Bucket := os.Getenv("BACKUP_S3_BUCKET")
	service := &BackupService{}

	if len(backupS3Bucket) > 0 {
		service.S3Backup = s3.NewHandler(backupAwsRegion, backupAwsAccessKeyID, backupAwsSecretAccessKey, backupS3Bucket)
	}

	return service
}

func (s *BackupService) BackupChannels(data []byte) {
	s.BackupChannelsWithRetry(data, 0)
}

func (s *BackupService) BackupChannelsWithRetry(data []byte, retries int) {
	name := fmt.Sprintf("%s.backup", strconv.FormatInt(time.Now().Unix(), 10))

	if s.S3Backup != nil {
		s.S3Backup.BackupChannelsWithRetry(name, data, retries)
	}
}
