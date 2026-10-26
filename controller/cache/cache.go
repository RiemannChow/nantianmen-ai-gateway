package cache

import (
	"reflect"

	cache_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/cache/dto"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type ICacheController interface {
	GetConfig(ctx *gin.Context) (*cache_dto.Config, error)
	SetConfig(ctx *gin.Context, input *cache_dto.Config) error
	Stats(ctx *gin.Context) (*cache_dto.Stats, error)
	Clear(ctx *gin.Context) error
	List(ctx *gin.Context) ([]*cache_dto.Entry, error)
	Query(ctx *gin.Context, input *cache_dto.Query) (*cache_dto.Entry, error)
	Store(ctx *gin.Context, input *cache_dto.Store) error
}

func init() {
	autowire.Auto[ICacheController](func() reflect.Value {
		return reflect.ValueOf(new(imlCacheController))
	})
}
