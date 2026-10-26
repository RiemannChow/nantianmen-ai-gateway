package cluster

import (
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/stores/cluster"
)

type Cluster struct {
	Uuid       string
	Name       string
	Resume     string
	Cluster    string
	Role       string
	Priority   int
	Enabled    bool
	Healthy    int
	Creator    string
	Updater    string
	Status     int
	CreateTime time.Time
	UpdateTime time.Time
}

func FromEntity(entity *cluster.Cluster) *Cluster {
	return &Cluster{
		Uuid:       entity.UUID,
		Name:       entity.Name,
		Resume:     entity.Resume,
		Cluster:    entity.Cluster,
		Role:       entity.Role,
		Priority:   entity.Priority,
		Enabled:    entity.Enabled,
		Healthy:    entity.Healthy,
		Creator:    entity.Creator,
		Updater:    entity.Updater,
		CreateTime: entity.CreateAt,
		UpdateTime: entity.UpdateAt,
	}
}

type Node struct {
	Uuid       string
	Name       string
	Cluster    string
	Peer       []string
	Admin      []string
	Server     []string
	CreateTime time.Time
}
