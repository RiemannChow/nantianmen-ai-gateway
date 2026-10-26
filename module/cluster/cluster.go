package cluster

import (
	"context"
	"reflect"

	cluster_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/cluster/dto"
	"github.com/eolinker/go-common/autowire"
)

type IClusterModule interface {
	CheckCluster(ctx context.Context, address ...string) ([]*cluster_dto.Node, error)
	ResetCluster(ctx context.Context, clusterId string, address string) ([]*cluster_dto.Node, error)
	ClusterNodes(ctx context.Context, clusterId string) ([]*cluster_dto.Node, error)
	// DR / multi-cluster management (gated by the multi_cluster license module).
	ListClusters(ctx context.Context) ([]*cluster_dto.ClusterView, error)
	CreateCluster(ctx context.Context, input *cluster_dto.CreateCluster) (*cluster_dto.ClusterView, error)
	UpdateCluster(ctx context.Context, id string, input *cluster_dto.UpdateCluster) (*cluster_dto.ClusterView, error)
	DeleteCluster(ctx context.Context, id string) error
	DRView(ctx context.Context) (*cluster_dto.DRView, error)
	// ProbeCluster re-checks connectivity of a managed cluster by its stored admin
	// address and refreshes its health state.
	ProbeCluster(ctx context.Context, id string) ([]*cluster_dto.Node, error)
}

func init() {
	autowire.Auto[IClusterModule](func() reflect.Value {
		return reflect.ValueOf(new(imlClusterModule))
	})
}
