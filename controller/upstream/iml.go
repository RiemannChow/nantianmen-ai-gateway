package upstream

import (
	"github.com/RiemannChow/nantianmen-ai-gateway/module/upstream"
	upstream_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/upstream/dto"
	"github.com/gin-gonic/gin"
)

var (
	_ IUpstreamController = (*imlUpstreamController)(nil)
)

type imlUpstreamController struct {
	upstreamModule upstream.IUpstreamModule `autowired:""`
}

func (i *imlUpstreamController) Get(ctx *gin.Context, serviceId string) (upstream_dto.UpstreamConfig, error) {
	return i.upstreamModule.Get(ctx, serviceId)
}

func (i *imlUpstreamController) Save(ctx *gin.Context, serviceId string, upstream *upstream_dto.Upstream) (upstream_dto.UpstreamConfig, error) {
	return i.upstreamModule.Save(ctx, serviceId, upstream)
}
