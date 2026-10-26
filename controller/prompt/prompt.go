package prompt

import (
	"reflect"

	prompt_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/prompt/dto"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type IPromptController interface {
	List(ctx *gin.Context, keyword string) ([]*prompt_dto.Item, error)
	Save(ctx *gin.Context, input *prompt_dto.Save) (*prompt_dto.Item, error)
	Delete(ctx *gin.Context, id string) error
}

func init() {
	autowire.Auto[IPromptController](func() reflect.Value {
		return reflect.ValueOf(new(imlPromptController))
	})
}
