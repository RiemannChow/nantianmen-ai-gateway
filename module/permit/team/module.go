package team

import (
	"context"
	"reflect"

	"github.com/eolinker/go-common/autowire"
)

type ITeamPermitModule interface {
	Permissions(ctx context.Context, teamId string) ([]string, error)
}

func init() {
	var m *imlTeamPermitModule

	autowire.Auto[ITeamPermitModule](func() reflect.Value {
		if m == nil {
			m = new(imlTeamPermitModule)
		}
		return reflect.ValueOf(m)
	})

}
