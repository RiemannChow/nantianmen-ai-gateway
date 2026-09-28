package team_member

import (
	"reflect"
	
	"github.com/eolinker/go-common/autowire"
	
	"github.com/RiemannChow/nantianmen-ai-gateway/stores/team"
	"github.com/eolinker/ap-account/service/member"
)

type ITeamMemberService member.IMemberService

func init() {
	autowire.Auto[ITeamMemberService](func() reflect.Value {
		return reflect.ValueOf(new(member.Service[team.ITeamMemberStore]))
	})
}
