package mocks

import (
	"errors"
)

type MockBackupService struct {
	backupChannelsMockData [][]byte
}

func NewService() *MockBackupService {
	return &MockBackupService{}
}

func (s *MockBackupService) BackupChannels(data []byte) {
	s.backupChannelsMockData = append(s.backupChannelsMockData, data)
}

func (s *MockBackupService) BackupChannelsWithRetry(data []byte, retries int) {
	s.backupChannelsMockData = append(s.backupChannelsMockData, data)
}

func (s *MockBackupService) GetBackupMockData() ([]byte, error) {
	if len(s.backupChannelsMockData) == 0 {
		return nil, errors.New("NotFound")
	}

	response := s.backupChannelsMockData[0]
	s.backupChannelsMockData = s.backupChannelsMockData[1:]
	return response, nil}
