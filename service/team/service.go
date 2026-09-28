package team

import (
	"context"
	"reflect"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/universally"
	"github.com/eolinker/go-common/autowire"
)

type ITeamService interface {
	universally.IServiceGet[Team]
	universally.IServiceDelete
	universally.IServiceCreate[CreateTeam]
	universally.IServiceEdit[EditTeam]
	DefaultTeam(ctx context.Context) (*Team, error)
}

func init() {
	autowire.Auto[ITeamService](func() reflect.Value {
		return reflect.ValueOf(new(imlTeamService))
	})
}
