package log

import (
	"github.com/RiemannChow/nantianmen-ai-gateway/module/log"
	log_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/log/dto"
	"github.com/gin-gonic/gin"
)

type imlLogController struct {
	module log.ILogModule `autowired:""`
}

func (c *imlLogController) Save(ctx *gin.Context, driver string, input *log_dto.Save) error {
	return c.module.Save(ctx, driver, input)
}

func (c *imlLogController) Get(ctx *gin.Context, driver string) (*log_dto.LogSource, error) {
	return c.module.Get(ctx, driver)
}
