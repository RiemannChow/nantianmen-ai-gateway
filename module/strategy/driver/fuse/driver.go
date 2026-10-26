package fuse

import (
	"github.com/RiemannChow/nantianmen-ai-gateway/gateway"
	"github.com/eolinker/eosc"

	strategy_driver "github.com/RiemannChow/nantianmen-ai-gateway/module/strategy/driver"
	strategy_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/strategy/dto"
)

// init registers the "fuse" (熔断) strategy driver. It maps to the Apinto
// strategy-plugin "fuse" so circuit breaking can be published to the data plane.
// This is an enterprise (commercial) capability.
func init() {
	strategy_driver.Register(&strategyDriver{confName: "fuse"})
}

type strategyDriver struct {
	confName string
}

func (d *strategyDriver) Driver() string {
	return "fuse"
}

func (d *strategyDriver) ToRelease(s *strategy_dto.Strategy, labels map[string][]string, initStep int) *eosc.Base[gateway.StrategyRelease] {
	filters := make(map[string][]string)
	for _, f := range s.Filters {
		filters[f.Name] = f.Values
	}
	for key, value := range labels {
		filters[key] = value
	}

	base := eosc.NewBase[gateway.StrategyRelease]()
	base.Config = &gateway.StrategyRelease{
		Name:     s.Id,
		Desc:     s.Name,
		Driver:   "fuse",
		Priority: initStep + s.Priority,
		Filters:  filters,
		IsDelete: s.IsDelete,
	}
	cfg := make(map[string]interface{})
	cfg[d.confName] = s.Config
	base.Append = cfg
	return base
}

// Check is intentionally permissive: the detailed schema is validated by the
// Apinto data plane on publish.
func (d *strategyDriver) Check(config interface{}) error {
	if config == nil {
		return nil
	}
	return nil
}
