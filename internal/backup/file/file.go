package file

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	metrics "github.com/satimoto/go-lsp/internal/metric"
)

type FileBackup interface {
	BackupChannels(name string, data []byte) error
	BackupChannelsWithRetry(name string, data []byte, retries int) error
}

type FileBackupHandler struct {
	filePath string
}

func NewHandler(filePath string) FileBackup {
	return &FileBackupHandler{
		filePath: filePath,
	}
}

func (h *FileBackupHandler) BackupChannels(name string, data []byte) error {
	return h.BackupChannelsWithRetry(name, data, 0)
}

func (h *FileBackupHandler) BackupChannelsWithRetry(name string, data []byte, retries int) error {
	fileName := fmt.Sprintf("%s/%s", h.filePath, name)

	for i := 0; i <= retries; i++ {
		err := h.processBackupChannels(fileName, data)

		if err == nil {
			break
		}

		if i == retries {
			metrics.RecordError("LSP065", "Error in file backup", err)
			log.Printf("LSP065: FileName=%v, Retries=%v", fileName, retries)
			return err
		}

		time.Sleep((time.Duration(i) + 1) * time.Second)
	}

	return nil
}

func (h *FileBackupHandler) processBackupChannels(fileName string, data []byte) error {
	file, err := os.Create(fileName)

	if err != nil {
		metrics.RecordError("LSP066", "Error creating file", err)
		log.Printf("LSP066: FileName=%v", fileName)
		return errors.New("error creating file")
	}

	writer := bufio.NewWriter(file)
	_, err = writer.Write(data)

	if err != nil {
		metrics.RecordError("LSP067", "Error writing to file", err)
		return errors.New("error writing to file")
	}

	err = writer.Flush()

	if err != nil {
		metrics.RecordError("LSP068", "Error flushing writer", err)
		return errors.New("error flushing writer")
	}

	return nil
}
