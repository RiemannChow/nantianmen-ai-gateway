package prompt

import (
	"context"
	"testing"

	prompt_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/prompt/dto"
)

// memSetting is an in-memory ISettingService for tests.
type memSetting struct {
	data map[string]string
}

func (m *memSetting) Get(ctx context.Context, name string) (string, bool) {
	v, ok := m.data[name]
	return v, ok
}
func (m *memSetting) Set(ctx context.Context, name, value, operator string) error {
	m.data[name] = value
	return nil
}
func (m *memSetting) All(ctx context.Context) (map[string]string, error) { return m.data, nil }

func newTestModule() (*imlPromptModule, *memSetting) {
	s := &memSetting{data: map[string]string{}}
	return &imlPromptModule{settingService: s}, s
}

func TestPromptSaveAndList(t *testing.T) {
	m, _ := newTestModule()
	ctx := context.Background()
	item, err := m.Save(ctx, &prompt_dto.Save{Name: "greeting", Tags: []string{"a"}, Content: "hi"})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if item.Id == "" {
		t.Fatal("expected generated id")
	}
	list, err := m.List(ctx, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Name != "greeting" {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func TestPromptUpdateKeepsId(t *testing.T) {
	m, _ := newTestModule()
	ctx := context.Background()
	item, _ := m.Save(ctx, &prompt_dto.Save{Name: "v1", Content: "x"})
	updated, err := m.Save(ctx, &prompt_dto.Save{Id: item.Id, Name: "v2", Content: "y"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Id != item.Id {
		t.Fatalf("id changed: %s -> %s", item.Id, updated.Id)
	}
	list, _ := m.List(ctx, "")
	if len(list) != 1 {
		t.Fatalf("expected one template after update, got %d", len(list))
	}
}

func TestPromptDelete(t *testing.T) {
	m, _ := newTestModule()
	ctx := context.Background()
	item, _ := m.Save(ctx, &prompt_dto.Save{Name: "tmp"})
	if err := m.Delete(ctx, item.Id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, _ := m.List(ctx, "")
	if len(list) != 0 {
		t.Fatalf("expected empty after delete, got %d", len(list))
	}
	if err := m.Delete(ctx, "missing"); err == nil {
		t.Fatal("expected error deleting missing template")
	}
}

func TestPromptKeywordFilter(t *testing.T) {
	m, _ := newTestModule()
	ctx := context.Background()
	m.Save(ctx, &prompt_dto.Save{Name: "alpha", Content: "A"})
	m.Save(ctx, &prompt_dto.Save{Name: "beta", Content: "B"})
	list, _ := m.List(ctx, "alpha")
	if len(list) != 1 || list[0].Name != "alpha" {
		t.Fatalf("keyword filter failed: %+v", list)
	}
}

func TestPromptNameRequired(t *testing.T) {
	m, _ := newTestModule()
	if _, err := m.Save(context.Background(), &prompt_dto.Save{}); err == nil {
		t.Fatal("expected error when name is empty")
	}
}
