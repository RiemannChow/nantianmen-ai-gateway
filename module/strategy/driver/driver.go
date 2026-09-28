package strategy_driver

import (
	"github.com/RiemannChow/nantianmen-ai-gateway/gateway"
	strategy_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/strategy/dto"
	"github.com/eolinker/eosc"
)

type IStrategyDriver interface {
	Driver() string
	ToRelease(s *strategy_dto.Strategy, labels map[string][]string, initStep int) *eosc.Base[gateway.StrategyRelease]
	Check(config interface{}) error
}
