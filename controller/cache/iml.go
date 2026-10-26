package cache

import (
	cache_module "github.com/RiemannChow/nantianmen-ai-gateway/module/cache"
	cache_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/cache/dto"
	"github.com/gin-gonic/gin"
)

var _ ICacheController = (*imlCacheController)(nil)

type imlCacheController struct {
	module cache_module.ICacheModule `autowired:""`
}

func (i *imlCacheController) GetConfig(ctx *gin.Context) (*cache_dto.Config, error) {
	return i.module.GetConfig(ctx)
}

func (i *imlCacheController) SetConfig(ctx *gin.Context, input *cache_dto.Config) error {
	return i.module.SetConfig(ctx, input)
}

func (i *imlCacheController) Stats(ctx *gin.Context) (*cache_dto.Stats, error) {
	return i.module.Stats(ctx)
}

func (i *imlCacheController) Clear(ctx *gin.Context) error {
	return i.module.Clear(ctx)
}

func (i *imlCacheController) List(ctx *gin.Context) ([]*cache_dto.Entry, error) {
	return i.module.List(ctx)
}

func (i *imlCacheController) Query(ctx *gin.Context, input *cache_dto.Query) (*cache_dto.Entry, error) {
	return i.module.Query(ctx, input)
}

func (i *imlCacheController) Store(ctx *gin.Context, input *cache_dto.Store) error {
	return i.module.Store(ctx, input)
}
