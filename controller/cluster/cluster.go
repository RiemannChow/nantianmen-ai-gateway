package cluster

import (
	"reflect"

	cluster_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/cluster/dto"

	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type IClusterController interface {
	Nodes(ctx *gin.Context, clusterId string) ([]*cluster_dto.Node, error)
	ResetCluster(ctx *gin.Context, clusterId string, input *cluster_dto.ResetCluster) ([]*cluster_dto.Node, error)
	Check(ctx *gin.Context, input *cluster_dto.CheckCluster) ([]*cluster_dto.Node, error)
	// DR / multi-cluster management (gated by the multi_cluster license module).
	ListClusters(ctx *gin.Context) ([]*cluster_dto.ClusterView, error)
	CreateCluster(ctx *gin.Context, input *cluster_dto.CreateCluster) (*cluster_dto.ClusterView, error)
	UpdateCluster(ctx *gin.Context, id string, input *cluster_dto.UpdateCluster) (*cluster_dto.ClusterView, error)
	DeleteCluster(ctx *gin.Context, id string) (string, error)
	DRView(ctx *gin.Context) (*cluster_dto.DRView, error)
	Probe(ctx *gin.Context, id string) ([]*cluster_dto.Node, error)
}

func init() {
	autowire.Auto[IClusterController](func() reflect.Value {
		return reflect.ValueOf(new(imlCluster))
	})
}
