package cluster_dto

import (
	"github.com/eolinker/go-common/auto"
)

type Item struct {
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ClusterNum  int            `json:"cluster_num"`
	CreateTime  auto.TimeLabel `json:"create_time" `
	UpdateTime  auto.TimeLabel `json:"update_time"`
	Updater     auto.Label     `json:"updater" aolabel:"user"`
	Creator     auto.Label     `json:"creator" aolabel:"user"`
}
type Simple struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
type Cluster struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
type SimpleWithCluster struct {
	Id       string     `json:"id,omitempty"`
	Name     string     `json:"name,omitempty"`
	Clusters []*Cluster `json:"clusters,omitempty"`
}

type Detail struct {
	Updater     auto.Label     `json:"updater"`
	Creator     auto.Label     `json:"creator"`
	Id          string         `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Prefix      string         `json:"prefix,omitempty"`
	CreateTime  auto.TimeLabel `json:"create_time,omitempty"`
	UpdateTime  auto.TimeLabel `json:"update_time,omitempty"`
	CanDelete   bool           `json:"can_delete"`
}

type Node struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Admins   []string `json:"manager_address"`
	Peers    []string `json:"peer_address"`
	Gateways []string `json:"service_address"`
	Status   int      `json:"status"`
}

// ClusterView is the API projection of a managed cluster, including DR attributes.
type ClusterView struct {
	Uuid       string         `json:"uuid"`
	Name       string         `json:"name"`
	Resume     string         `json:"resume"`
	Role       string         `json:"role"`
	Priority   int            `json:"priority"`
	Enabled    bool           `json:"enabled"`
	Healthy    int            `json:"healthy"`
	NodeCount  int            `json:"node_count"`
	CreateTime auto.TimeLabel `json:"create_time"`
	UpdateTime auto.TimeLabel `json:"update_time"`
}

// DRView is the disaster-recovery overview across all clusters.
type DRView struct {
	Total    int            `json:"total"`
	Enabled  int            `json:"enabled"`
	Healthy  int            `json:"healthy"`
	Primary  int            `json:"primary"`
	Backup   int            `json:"backup"`
	Clusters []*ClusterView `json:"clusters"`
}
