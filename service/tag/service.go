package tag

import (
	"context"
	"reflect"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/universally"
	"github.com/eolinker/go-common/autowire"
)

type ITagService interface {
	universally.IServiceGet[Tag]
	universally.IServiceDelete
	universally.IServiceCreate[CreateTag]
	Map(ctx context.Context, ids ...string) (map[string]*Tag, error)
}

func init() {
	autowire.Auto[ITagService](func() reflect.Value {
		return reflect.ValueOf(new(imlTagService))
	})
}
