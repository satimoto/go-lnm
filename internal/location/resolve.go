package location

import (
	"context"

	"github.com/satimoto/go-datastore/db"
)

type LocationRepository interface {
	GetConnector(ctx context.Context, id int64) (db.Connector, error)
	GetEvse(ctx context.Context, id int64) (db.Evse, error)
	GetLocation(ctx context.Context, id int64) (db.Location, error)
}

type LocationResolver struct {
	Repository LocationRepository
}

func NewResolver(repositoryService *db.RepositoryService) *LocationResolver {
	repo := LocationRepository(repositoryService)
	return &LocationResolver{repo}
}
