package cluster

import (
	"github.com/RiemannChow/nantianmen-ai-gateway/module/cluster"
	cluster_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/cluster/dto"
	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
	"github.com/gin-gonic/gin"
)

var (
	_ IClusterController = (*imlCluster)(nil)
)

type imlCluster struct {
	module cluster.IClusterModule `autowired:""`
}

func (p *imlCluster) Nodes(ctx *gin.Context, clusterId string) ([]*cluster_dto.Node, error) {
	if clusterId == "" {
		clusterId = "default"
	}
	return p.module.ClusterNodes(ctx, clusterId)
}

func (p *imlCluster) ResetCluster(ctx *gin.Context, clusterId string, input *cluster_dto.ResetCluster) ([]*cluster_dto.Node, error) {
	if clusterId == "" {
		clusterId = "default"
	}
	return p.module.ResetCluster(ctx, clusterId, input.ManagerAddress)
}

func (p *imlCluster) Check(ctx *gin.Context, input *cluster_dto.CheckCluster) ([]*cluster_dto.Node, error) {
	return p.module.CheckCluster(ctx, input.Address)
}

func (p *imlCluster) ListClusters(ctx *gin.Context) ([]*cluster_dto.ClusterView, error) {
	if err := license.Check(ctx, "multi_cluster"); err != nil {
		return nil, err
	}
	return p.module.ListClusters(ctx)
}

func (p *imlCluster) CreateCluster(ctx *gin.Context, input *cluster_dto.CreateCluster) (*cluster_dto.ClusterView, error) {
	if err := license.Check(ctx, "multi_cluster"); err != nil {
		return nil, err
	}
	return p.module.CreateCluster(ctx, input)
}

func (p *imlCluster) UpdateCluster(ctx *gin.Context, id string, input *cluster_dto.UpdateCluster) (*cluster_dto.ClusterView, error) {
	if err := license.Check(ctx, "multi_cluster"); err != nil {
		return nil, err
	}
	return p.module.UpdateCluster(ctx, id, input)
}

func (p *imlCluster) DeleteCluster(ctx *gin.Context, id string) (string, error) {
	if err := license.Check(ctx, "multi_cluster"); err != nil {
		return "", err
	}
	if err := p.module.DeleteCluster(ctx, id); err != nil {
		return "", err
	}
	return id, nil
}

func (p *imlCluster) DRView(ctx *gin.Context) (*cluster_dto.DRView, error) {
	if err := license.Check(ctx, "multi_cluster"); err != nil {
		return nil, err
	}
	return p.module.DRView(ctx)
}

func (p *imlCluster) Probe(ctx *gin.Context, id string) ([]*cluster_dto.Node, error) {
	if err := license.Check(ctx, "multi_cluster"); err != nil {
		return nil, err
	}
	return p.module.ProbeCluster(ctx, id)
}

//
//func (p *imlCluster) SimpleWithCluster(ctx *gin.Context) ([]*parition_dto.SimpleWithCluster, error) {
//	return p.module.SimpleWithCluster(ctx)
//}
//
//func (p *imlCluster) Delete(ctx *gin.Context, id string) (string, error) {
//	err := p.module.Delete(ctx, id)
//	if err != nil {
//		return "", err
//	}
//	return id, nil
//}
//
//func (p *imlCluster) SearchByDriver(ctx *gin.Context, keyword string) ([]*parition_dto.Item, error) {
//	return p.module.SearchByDriver(ctx, keyword)
//}
//
//func (p *imlCluster) Simple(ctx *gin.Context) ([]*parition_dto.Simple, error) {
//	return p.module.Simple(ctx)
//}
//
//func (p *imlCluster) Info(ctx *gin.Context, id string) (*parition_dto.Detail, error) {
//	if id == "" {
//		return nil, errors.New("id is empty")
//	}
//	return p.module.Get(ctx, id)
//}
//
//func (p *imlCluster) Update(ctx *gin.Context, id string, input *parition_dto.Edit) (*parition_dto.Detail, error) {
//	return p.module.Update(ctx, id, input)
//}
//
//func (p *imlCluster) Create(ctx *gin.Context, input *parition_dto.Create) (*parition_dto.Detail, string, auto.TimeLabel, error) {
//	detail, err := p.module.CreatePartition(ctx, input)
//	if err != nil {
//		return nil, "", auto.TimeLabel{}, err
//	}
//	return detail, detail.Id, detail.UpdateTime, nil
//}
