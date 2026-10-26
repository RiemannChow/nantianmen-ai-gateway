package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/cache/dto"
	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"

	gocache "github.com/eolinker/go-common/cache"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"
)

const (
	configKey  = "semantic_cache.config"
	indexKey   = "semantic_cache:index"
	hitsKey    = "semantic_cache:stats:hits"
	missesKey  = "semantic_cache:stats:misses"
	keyPrefix  = "semantic_cache:"
	defaultTTL = int64(7 * 24 * 60 * 60) // 7 days
)

// ICacheModule manages the semantic cache configuration, storage and stats.
// Storage uses the shared Redis ICommonCache; the control-plane exposes the
// Query/Store seam that the data-plane (or any caller) uses to short-circuit
// repeated/similar prompts.
type ICacheModule interface {
	GetConfig(ctx context.Context) (*dto.Config, error)
	SetConfig(ctx context.Context, input *dto.Config) error
	Stats(ctx context.Context) (*dto.Stats, error)
	Clear(ctx context.Context) error
	List(ctx context.Context) ([]*dto.Entry, error)
	Query(ctx context.Context, input *dto.Query) (*dto.Entry, error)
	Store(ctx context.Context, input *dto.Store) error
}

type imlCacheModule struct {
	settingService setting.ISettingService `autowired:""`
	commonCache    gocache.ICommonCache    `autowired:""`
}

func init() {
	autowire.Auto[ICacheModule](func() reflect.Value {
		return reflect.ValueOf(new(imlCacheModule))
	})
}

var _ ICacheModule = (*imlCacheModule)(nil)

func (i *imlCacheModule) GetConfig(ctx context.Context) (*dto.Config, error) {
	v, has := i.settingService.Get(ctx, configKey)
	if !has || v == "" {
		return &dto.Config{Enabled: false, TTLSeconds: defaultTTL}, nil
	}
	cfg := &dto.Config{}
	if err := json.Unmarshal([]byte(v), cfg); err != nil {
		return &dto.Config{Enabled: false, TTLSeconds: defaultTTL}, nil
	}
	if cfg.TTLSeconds <= 0 {
		cfg.TTLSeconds = defaultTTL
	}
	return cfg, nil
}

func (i *imlCacheModule) SetConfig(ctx context.Context, input *dto.Config) error {
	if input.TTLSeconds <= 0 {
		input.TTLSeconds = defaultTTL
	}
	bytes, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return i.settingService.Set(ctx, configKey, string(bytes), utils.UserId(ctx))
}

func (i *imlCacheModule) Stats(ctx context.Context) (*dto.Stats, error) {
	hits, err := i.commonCache.GetInt(ctx, hitsKey)
	if err != nil {
		hits = 0
	}
	misses, err := i.commonCache.GetInt(ctx, missesKey)
	if err != nil {
		misses = 0
	}
	total := hits + misses
	rate := 0.0
	if total > 0 {
		rate = float64(hits) / float64(total)
	}
	entries := int64(0)
	index, err := i.commonCache.HGetAll(ctx, indexKey)
	if err == nil {
		entries = int64(len(index))
	}
	return &dto.Stats{
		Hits:    hits,
		Misses:  misses,
		HitRate: rate,
		Entries: entries,
	}, nil
}

func (i *imlCacheModule) Clear(ctx context.Context) error {
	index, err := i.commonCache.HGetAll(ctx, indexKey)
	if err == nil && len(index) > 0 {
		keys := make([]string, 0, len(index))
		for k := range index {
			keys = append(keys, k)
		}
		if derr := i.commonCache.Del(ctx, keys...); derr != nil {
			return derr
		}
	}
	if derr := i.commonCache.Del(ctx, indexKey, hitsKey, missesKey); derr != nil {
		return derr
	}
	return nil
}

func (i *imlCacheModule) List(ctx context.Context) ([]*dto.Entry, error) {
	index, err := i.commonCache.HGetAll(ctx, indexKey)
	if err != nil {
		return []*dto.Entry{}, nil
	}
	entries := make([]*dto.Entry, 0, len(index))
	for key := range index {
		bytes, gerr := i.commonCache.Get(ctx, key)
		if gerr != nil || len(bytes) == 0 {
			continue
		}
		entry := &dto.Entry{}
		if jerr := json.Unmarshal(bytes, entry); jerr == nil {
			entries = append(entries, entry)
		}
	}
	sort.Slice(entries, func(a, b int) bool {
		return entries[a].CreatedAt > entries[b].CreatedAt
	})
	return entries, nil
}

func (i *imlCacheModule) Query(ctx context.Context, input *dto.Query) (*dto.Entry, error) {
	cfg, err := i.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return nil, nil
	}
	if input.Prompt == "" || input.Model == "" {
		return nil, errors.New("prompt and model are required")
	}
	key := buildKey(input.Tenant, input.Model, normalize(input.Prompt))
	entry, hit, err := i.get(ctx, key)
	if err != nil {
		return nil, err
	}
	if hit {
		_ = i.commonCache.Incr(ctx, hitsKey, 0)
		return entry, nil
	}
	_ = i.commonCache.Incr(ctx, missesKey, 0)
	return nil, nil
}

func (i *imlCacheModule) Store(ctx context.Context, input *dto.Store) error {
	cfg, err := i.GetConfig(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return errors.New("semantic cache is disabled")
	}
	if input.Prompt == "" || input.Model == "" {
		return errors.New("prompt and model are required")
	}
	key := buildKey(input.Tenant, input.Model, normalize(input.Prompt))
	entry := &dto.Entry{
		Key:        key,
		Model:      input.Model,
		Tenant:     input.Tenant,
		PromptHash: hashPrompt(normalize(input.Prompt)),
		Response:   input.Response,
		TokenCount: input.TokenCount,
		CreatedAt:  time.Now().Unix(),
	}
	bytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	ttl := time.Duration(cfg.TTLSeconds) * time.Second
	if err := i.commonCache.Set(ctx, key, bytes, ttl); err != nil {
		return err
	}
	return i.commonCache.HMSet(ctx, indexKey, map[string][]byte{key: []byte(entry.Model)}, 0)
}

func (i *imlCacheModule) get(ctx context.Context, key string) (*dto.Entry, bool, error) {
	bytes, err := i.commonCache.Get(ctx, key)
	if err != nil || len(bytes) == 0 {
		return nil, false, nil
	}
	entry := &dto.Entry{}
	if jerr := json.Unmarshal(bytes, entry); jerr != nil {
		return nil, false, nil
	}
	return entry, true, nil
}

func buildKey(tenant, model, normalized string) string {
	return keyPrefix + tenant + ":" + model + ":" + hashPrompt(normalized)
}

func hashPrompt(normalized string) string {
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

// normalize trims whitespace and lowercases so trivial edits still collide.
func normalize(prompt string) string {
	out := []rune(prompt)
	start, end := 0, len(out)
	for start < end && isSpace(out[start]) {
		start++
	}
	for end > start && isSpace(out[end-1]) {
		end--
	}
	return lower(string(out[start:end]))
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
