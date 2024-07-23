package mocks

import (
	"github.com/satimoto/go-datastore/pkg/db/mocks"
	cdr "github.com/satimoto/go-lnm/internal/cdr/mocks"
	"github.com/satimoto/go-lnm/internal/monitor/startup"
	"github.com/satimoto/go-lnm/internal/service"
	session "github.com/satimoto/go-lnm/internal/session/mocks"
)

func NewService(repositoryService *mocks.MockRepositoryService, services *service.ServiceResolver) *startup.StartupService {
	return &startup.StartupService{
		CdrResolver:     cdr.NewResolver(repositoryService, services),
		SessionResolver: session.NewResolver(repositoryService, services),
	}
}
