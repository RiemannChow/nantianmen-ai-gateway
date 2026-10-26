package strategy_filter

import (
	"testing"

	strategy_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/strategy/dto"
)

func globalScope() strategy_dto.Scope {
	return strategy_dto.ToScope(strategy_dto.ScopeGlobal)
}

func TestCheckFilters_Valid(t *testing.T) {
	filters := []*strategy_dto.Filter{
		{Name: "ip", Values: []string{"192.168.1.0/24"}},
		{Name: "method", Values: []string{"GET", "POST"}},
		{Name: "path", Values: []string{"/api/*"}},
	}
	if err := CheckFilters("strategy", globalScope(), filters); err != nil {
		t.Errorf("CheckFilters with valid filters: %v", err)
	}
}

func TestCheckFilters_UnknownFilterName(t *testing.T) {
	filters := []*strategy_dto.Filter{
		{Name: "no-such-filter", Values: []string{"x"}},
	}
	err := CheckFilters("strategy", globalScope(), filters)
	if err == nil {
		t.Error("CheckFilters with unknown filter name should fail")
	}
}

func TestCheckFilters_PatternMismatch(t *testing.T) {
	filters := []*strategy_dto.Filter{
		// ip filter 只接受 IPv4/CIDR
		{Name: "ip", Values: []string{"not-an-ip"}},
	}
	err := CheckFilters("strategy", globalScope(), filters)
	if err == nil {
		t.Error("CheckFilters with pattern-mismatched value should fail")
	}
}

func TestCheckFilters_EmptyValues(t *testing.T) {
	filters := []*strategy_dto.Filter{
		{Name: "ip", Values: []string{}},
	}
	err := CheckFilters("strategy", globalScope(), filters)
	if err == nil {
		t.Error("CheckFilters with empty values should fail")
	}
}

func TestCheckFilters_DuplicateName(t *testing.T) {
	filters := []*strategy_dto.Filter{
		{Name: "ip", Values: []string{"10.0.0.1"}},
		{Name: "ip", Values: []string{"10.0.0.2"}},
	}
	err := CheckFilters("strategy", globalScope(), filters)
	if err == nil {
		t.Error("CheckFilters with duplicate filter names should fail")
	}
}

func TestFilterLabel(t *testing.T) {
	info, err := FilterLabel("method", []string{"GET"})
	if err != nil {
		t.Fatalf("FilterLabel(method): %v", err)
	}
	if info.Title != "api method" {
		t.Errorf("FilterLabel title = %q, want %q", info.Title, "api method")
	}
	if len(info.Label) == 0 {
		t.Error("FilterLabel returned empty label")
	}

	if _, err := FilterLabel("no-such-filter", nil); err == nil {
		t.Error("FilterLabel with unknown name should fail")
	}
}

func TestMethodFilterLabels(t *testing.T) {
	m := newMethodFilter()

	// ALL 应展示为 "all method"
	if got := m.Labels(ValuesALL); len(got) != 1 || got[0] != "all method" {
		t.Errorf("Labels(ALL) = %v, want [all method]", got)
	}
	// 空值展示占位符
	if got := m.Labels(); len(got) != 1 || got[0] != "-" {
		t.Errorf("Labels() = %v, want [-]", got)
	}
	// 具体方法原样返回
	if got := m.Labels("GET", "POST"); len(got) != 2 || got[0] != "GET" || got[1] != "POST" {
		t.Errorf("Labels(GET, POST) = %v, want [GET POST]", got)
	}
}

func TestOptionsRegistered(t *testing.T) {
	// ip/method/path 三个内置 filter 通过 init 注册，应在 global scope 下可见
	opts, ok := Options(ScopeGlobal)
	if !ok {
		t.Fatal("global scope options not registered")
	}
	for _, name := range []string{"ip", "method", "path"} {
		if _, has := opts[name]; !has {
			t.Errorf("filter %q not registered in global scope", name)
		}
	}
}
