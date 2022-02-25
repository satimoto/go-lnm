package node

import (
	"context"

	"github.com/satimoto/go-datastore/db"
)

type NodeRepository interface {
	CreateNode(ctx context.Context, arg db.CreateNodeParams) (db.Node, error)
	GetNode(ctx context.Context, id int64) (db.Node, error)
	GetNodeByPubkey(ctx context.Context, pubkey string) (db.Node, error)
	UpdateNode(ctx context.Context, arg db.UpdateNodeParams) (db.Node, error)
}

type NodeResolver struct {
	Repository NodeRepository
}

func NewResolver(repositoryService *db.RepositoryService) *NodeResolver {
	repo := NodeRepository(repositoryService)

	return &NodeResolver{
		Repository: repo,
	}
}

func NewUpdateNodeParams(node db.Node) db.UpdateNodeParams {
	return db.UpdateNodeParams{
		ID:         node.ID,
		Alias:      node.Alias,
		Color:      node.Color,
		CommitHash: node.CommitHash,
		Version:    node.Version,
		Channels:   node.Channels,
		Peers:      node.Peers,
	}
}
