package prompt

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"time"

	prompt_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/prompt/dto"
	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"

	"github.com/google/uuid"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"
)

const templatesKey = "prompt.templates"

// IPromptModule manages saved prompt templates backed by the KV setting store.
type IPromptModule interface {
	List(ctx context.Context, keyword string) ([]*prompt_dto.Item, error)
	Save(ctx context.Context, input *prompt_dto.Save) (*prompt_dto.Item, error)
	Delete(ctx context.Context, id string) error
}

func init() {
	autowire.Auto[IPromptModule](func() reflect.Value {
		return reflect.ValueOf(new(imlPromptModule))
	})
}

var _ IPromptModule = (*imlPromptModule)(nil)

type imlPromptModule struct {
	settingService setting.ISettingService `autowired:""`
}

// load reads the full template list from KV, tolerating an empty/malformed value.
func (i *imlPromptModule) load(ctx context.Context) ([]*prompt_dto.Item, error) {
	v, has := i.settingService.Get(ctx, templatesKey)
	if !has || v == "" {
		return []*prompt_dto.Item{}, nil
	}
	list := make([]*prompt_dto.Item, 0)
	if err := json.Unmarshal([]byte(v), &list); err != nil {
		// a corrupted blob must not wedge the feature; start clean
		return []*prompt_dto.Item{}, nil
	}
	return list, nil
}

// store writes the template list back to KV (best-effort operator attribution).
func (i *imlPromptModule) store(ctx context.Context, list []*prompt_dto.Item) error {
	bytes, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return i.settingService.Set(ctx, templatesKey, string(bytes), utils.UserId(ctx))
}

func (i *imlPromptModule) List(ctx context.Context, keyword string) ([]*prompt_dto.Item, error) {
	list, err := i.load(ctx)
	if err != nil {
		return nil, err
	}
	if keyword != "" {
		kw := keyword
		filtered := make([]*prompt_dto.Item, 0, len(list))
		for _, item := range list {
			if contains(item.Name, kw) || contains(item.Content, kw) {
				filtered = append(filtered, item)
			}
		}
		list = filtered
	}
	sort.Slice(list, func(a, b int) bool {
		return list[a].UpdatedAt > list[b].UpdatedAt
	})
	return list, nil
}

func (i *imlPromptModule) Save(ctx context.Context, input *prompt_dto.Save) (*prompt_dto.Item, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	list, err := i.load(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	item := &prompt_dto.Item{
		Name:      input.Name,
		Tags:      input.Tags,
		Content:   input.Content,
		Variables: input.Variables,
		UpdatedAt: now,
	}
	if input.Id == "" {
		item.Id = uuid.NewString()
		list = append(list, item)
	} else {
		found := false
		for idx, existing := range list {
			if existing.Id == input.Id {
				item.Id = existing.Id
				list[idx] = item
				found = true
				break
			}
		}
		if !found {
			item.Id = input.Id
			list = append(list, item)
		}
	}
	if err := i.store(ctx, list); err != nil {
		return nil, err
	}
	return item, nil
}

func (i *imlPromptModule) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	list, err := i.load(ctx)
	if err != nil {
		return err
	}
	next := make([]*prompt_dto.Item, 0, len(list))
	found := false
	for _, item := range list {
		if item.Id == id {
			found = true
			continue
		}
		next = append(next, item)
	}
	if !found {
		return errors.New("template not found")
	}
	return i.store(ctx, next)
}

func contains(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
