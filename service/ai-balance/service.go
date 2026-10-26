package ai_balance

import (
	"context"
	"reflect"

	"github.com/eolinker/go-common/autowire"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/universally"
)

type IBalanceService interface {
	universally.IServiceGet[Balance]
	universally.IServiceCreate[Create]
	universally.IServiceEdit[Edit]
	universally.IServiceDelete
	MaxPriority(ctx context.Context) (int, error)
	SortBefore(ctx context.Context, originID string, targetID string) ([]*Balance, error)
	SortAfter(ctx context.Context, originID string, targetID string) ([]*Balance, error)
	Exist(ctx context.Context, provider string, model string) (bool, error)
	UpdateHealth(ctx context.Context, id string, state int, lastCheckAt int64, lastSuccessAt int64, failCount int64) error
}

func init() {
	autowire.Auto[IBalanceService](func() reflect.Value {
		return reflect.ValueOf(new(imlBalanceService))
	})
}
