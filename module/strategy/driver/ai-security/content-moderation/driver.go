package contentmoderation

import (
	"encoding/json"

	"github.com/RiemannChow/nantianmen-ai-gateway/gateway"
	"github.com/eolinker/eosc"

	strategy_driver "github.com/RiemannChow/nantianmen-ai-gateway/module/strategy/driver"

	strategy_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/strategy/dto"
)

// Strategy driver that applies content-moderation / guardrail rules to LLM
// inputs and outputs (toxicity, policy words, etc.). Rules are opaque JSON
// enforced by the data-plane worker registered under "content_moderation".
func init() {
	strategy_driver.Register(&strategyDriver{confName: "content_moderation"})
}

type strategyDriver struct {
	confName string
}

func (d *strategyDriver) Driver() string {
	return "content-moderation"
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
		Driver:   "content_moderation",
		Priority: initStep + s.Priority,
		Filters:  filters,
		IsDelete: s.IsDelete,
	}
	cfg := make(map[string]interface{})
	cfg[d.confName] = s.Config
	base.Append = cfg
	return base
}

func (d *strategyDriver) Check(config interface{}) error {
	if config == nil {
		return nil
	}
	_, err := json.Marshal(config)
	return err
}
