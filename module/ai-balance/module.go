package ai_balance

import (
	"context"
	"reflect"

	"github.com/RiemannChow/nantianmen-ai-gateway/gateway"

	"github.com/eolinker/go-common/autowire"

	ai_balance_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/ai-balance/dto"
)

type IBalanceModule interface {
	Create(ctx context.Context, input *ai_balance_dto.Create) error
	Sort(ctx context.Context, input *ai_balance_dto.Sort) error
	List(ctx context.Context, keyword string) ([]*ai_balance_dto.Item, error)
	Delete(ctx context.Context, id string) error
	SyncLocalBalances(ctx context.Context, address string) error
	ProbeModel(ctx context.Context, id string) error
	BatchProbe(ctx context.Context) error
	UpdateState(ctx context.Context, id string, state int) error
	GetAutoFailover(ctx context.Context) (bool, error)
	SetAutoFailover(ctx context.Context, enabled bool) error
}

func init() {
	balanceModule := new(imlBalanceModule)
	autowire.Auto[IBalanceModule](func() reflect.Value {
		gateway.RegisterInitHandleFunc(balanceModule.initGateway)
		return reflect.ValueOf(balanceModule)
	})
}
