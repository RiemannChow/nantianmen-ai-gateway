package permit_team

import (
	"github.com/RiemannChow/nantianmen-ai-gateway/module/permit/team"
	"github.com/gin-gonic/gin"
)

var (
	_ ITeamPermitController = (*imlTeamPermitController)(nil)
)

type imlTeamPermitController struct {
	teamPermitModule team.ITeamPermitModule `autowired:""`
}

func (c *imlTeamPermitController) Permissions(ctx *gin.Context, team string) ([]string, error) {
	return c.teamPermitModule.Permissions(ctx, team)
}
