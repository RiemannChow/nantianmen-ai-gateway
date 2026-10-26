package limiting

import (
	"testing"

	strategy_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/strategy/dto"
)

func TestStrategyDriverToRelease(t *testing.T) {
	d := &strategyDriver{confName: "limiting"}
	if d.Driver() != "limiting" {
		t.Errorf("Driver() = %q, want %q", d.Driver(), "limiting")
	}

	s := &strategy_dto.Strategy{
		Id:       "strategy-1",
		Name:     "限流策略",
		Priority: 5,
		Filters: []*strategy_dto.Filter{
			{Name: "ip", Values: []string{"10.0.0.0/8"}},
		},
		Config: map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{"second": 10, "max": 100},
			},
		},
	}
	labels := map[string][]string{"service": {"svc-1"}}

	base := d.ToRelease(s, labels, 100)
	if base.Config == nil {
		t.Fatal("ToRelease returned nil Config")
	}
	if base.Config.Driver != "limiting" {
		t.Errorf("Driver = %q, want %q", base.Config.Driver, "limiting")
	}
	if base.Config.Priority != 105 {
		t.Errorf("Priority = %d, want %d (initStep + s.Priority)", base.Config.Priority, 105)
	}
	if base.Config.Name != "strategy-1" {
		t.Errorf("Name = %q, want %q", base.Config.Name, "strategy-1")
	}
	// strategy 自身 filters 与 labels 都应进入 filters
	if got := base.Config.Filters["ip"]; len(got) != 1 || got[0] != "10.0.0.0/8" {
		t.Errorf("Filters[ip] = %v, want [10.0.0.0/8]", got)
	}
	if got := base.Config.Filters["service"]; len(got) != 1 || got[0] != "svc-1" {
		t.Errorf("Filters[service] = %v, want [svc-1]", got)
	}
	// Append 下应以驱动名（limiting）为 key 携带原始 config
	appendCfg, ok := base.Append["limiting"]
	if !ok {
		t.Fatalf("Append missing %q key, got %v", "limiting", base.Append)
	}
	m, ok := appendCfg.(map[string]interface{})
	if !ok {
		t.Fatalf("Append[limiting] type = %T, want map[string]interface{}", appendCfg)
	}
	if _, has := m["rules"]; !has {
		t.Error("Append[limiting] missing rules key")
	}
}

func TestStrategyDriverCheck(t *testing.T) {
	d := &strategyDriver{confName: "limiting"}
	// nil 与非 nil 配置当前均通过（详细 schema 由 apinto 数据面校验）
	if err := d.Check(nil); err != nil {
		t.Errorf("Check(nil) = %v, want nil", err)
	}
	if err := d.Check(map[string]interface{}{"rules": []interface{}{}}); err != nil {
		t.Errorf("Check(config) = %v, want nil", err)
	}
}
