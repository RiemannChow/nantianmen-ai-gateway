package prompt

import (
	prompt "github.com/RiemannChow/nantianmen-ai-gateway/module/prompt"
	prompt_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/prompt/dto"
	"github.com/gin-gonic/gin"
)

var _ IPromptController = (*imlPromptController)(nil)

type imlPromptController struct {
	module prompt.IPromptModule `autowired:""`
}

func (i *imlPromptController) List(ctx *gin.Context, keyword string) ([]*prompt_dto.Item, error) {
	return i.module.List(ctx, keyword)
}

func (i *imlPromptController) Save(ctx *gin.Context, input *prompt_dto.Save) (*prompt_dto.Item, error) {
	return i.module.Save(ctx, input)
}

func (i *imlPromptController) Delete(ctx *gin.Context, id string) error {
	return i.module.Delete(ctx, id)
}
