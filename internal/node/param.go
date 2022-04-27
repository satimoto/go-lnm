package node

import "github.com/satimoto/go-datastore/db"

func NewUpdateNodeParams(node db.Node) db.UpdateNodeParams {
	return db.UpdateNodeParams{
		ID:         node.ID,
		NodeAddr:   node.NodeAddr,
		LspAddr:    node.LspAddr,
		Alias:      node.Alias,
		Color:      node.Color,
		CommitHash: node.CommitHash,
		Version:    node.Version,
		Channels:   node.Channels,
		Peers:      node.Peers,
	}
}
