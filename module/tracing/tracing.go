package tracing

import (
	"context"
	"math/rand"
	"reflect"
	"sort"
	"time"

	metering_store "github.com/RiemannChow/nantianmen-ai-gateway/stores/metering"
	"github.com/eolinker/go-common/autowire"
)

// ITracingService owns the local LLM-call trace pipeline: ingestion, a demo
// seeder, and a local-first aggregation used by the tracing observability page.
type ITracingService interface {
	Ingest(ctx context.Context, events []*metering_store.TraceEvent) error
	Seed(ctx context.Context, days int) (int, error)
	Overview(ctx context.Context, start, end int64) (*TraceOverview, error)
	List(ctx context.Context, start, end int64, page, pageSize int) (*TraceList, error)
}

func init() {
	autowire.Auto[ITracingService](func() reflect.Value {
		return reflect.ValueOf(new(imlTracingService))
	})
}

type imlTracingService struct {
	traceStore metering_store.ITraceStore `autowired:""`
}

func (i *imlTracingService) Ingest(ctx context.Context, events []*metering_store.TraceEvent) error {
	return i.traceStore.Insert(ctx, events...)
}

// Seed generates realistic sample LLM traces so the tracing page shows
// meaningful data on a fresh install without real Apinto traffic.
func (i *imlTracingService) Seed(ctx context.Context, days int) (int, error) {
	if days <= 0 {
		days = 30
	}
	if err := i.traceStore.DeleteAll(ctx); err != nil {
		return 0, err
	}
	providers := []string{"openai", "azure", "通义千问", "本地模型"}
	models := map[string]string{
		"openai": "gpt-4o",
		"azure":  "gpt-4o-mini",
		"通义千问":   "qwen-max",
		"本地模型":   "qwen2.5-7b",
	}
	subs := []string{"team-alpha", "team-beta", "internal-app"}
	svcs := []string{"chat-service", "summary-service", "code-assistant"}
	prompts := []string{
		"请把下面这段中文翻译成英文：",
		"总结一下这段会议纪要的要点：",
		"用 Python 写一个快速排序。",
		"给我写一封礼貌的催款邮件。",
		"解释一下什么是向量数据库。",
	}
	completions := []string{
		"Here is the translation: ...",
		"要点如下：1) 项目进度符合预期；2) 需关注交付风险；3) ...",
		"def quicksort(arr): ...",
		"尊敬的客户，您好，关于款项事宜...",
		"向量数据库是一种专门存储和检索向量嵌入的数据库...",
	}
	now := time.Now()
	r := rand.New(rand.NewSource(now.UnixNano()))
	events := make([]*metering_store.TraceEvent, 0, days*6)
	for d := days; d >= 0; d-- {
		day := now.AddDate(0, 0, -d)
		for h := 0; h < 6; h++ {
			ts := day.Add(time.Duration(r.Intn(24)) * time.Hour)
			p := providers[r.Intn(len(providers))]
			prompt := prompts[r.Intn(len(prompts))]
			completion := completions[r.Intn(len(completions))]
			inTok := int64(r.Intn(1500) + 200)
			outTok := int64(r.Intn(800) + 100)
			ttft := int64(r.Intn(800) + 50)
			gen := int64(r.Intn(4000) + 200)
			code := 200
			if roll := r.Intn(100); roll > 95 {
				code = 500
			}
			events = append(events, &metering_store.TraceEvent{
				ClusterId:        "default",
				ServiceId:        svcs[r.Intn(len(svcs))],
				Provider:         p,
				Subscriber:       subs[r.Intn(len(subs))],
				Model:            models[p],
				Prompt:           prompt,
				Completion:       completion,
				FinishReason:     "stop",
				PromptTokens:     inTok,
				CompletionTokens: outTok,
				TotalTokens:      inTok + outTok,
				StatusCode:       code,
				LatencyTotalMs:   ttft + gen,
				LatencyTTFTMs:    ttft,
				LatencyGenMs:     gen,
				CreatedAt:        ts,
			})
		}
	}
	if err := i.traceStore.Insert(ctx, events...); err != nil {
		return 0, err
	}
	return len(events), nil
}

// Overview aggregates the local trace table for the dashboard. It is local-first:
// if no local traces exist in the window it falls back to an empty projection
// (the Apinto InfluxDB trace path is out of scope for the control-plane suite).
func (i *imlTracingService) Overview(ctx context.Context, start, end int64) (*TraceOverview, error) {
	out := &TraceOverview{Window: Window{Start: start, End: end}}
	n, err := i.traceStore.CountInWindow(ctx, start, end)
	if err != nil {
		return nil, err
	}
	out.TotalCalls = n
	if n == 0 {
		return out, nil
	}
	calls, in, outTok, err := i.traceStore.SummaryInWindow(ctx, start, end)
	if err != nil {
		return nil, err
	}
	out.TotalCalls = calls
	out.InputTokens = in
	out.OutputTokens = outTok
	out.TotalTokens = in + outTok

	// provider distribution
	prov, err := i.traceStore.ByProviderInWindow(ctx, start, end, 200)
	if err != nil {
		return nil, err
	}
	rows := make([]*ProviderCall, 0, len(prov))
	for _, p := range prov {
		rows = append(rows, &ProviderCall{Provider: p.Provider, Calls: p.Calls})
	}
	sort.SliceStable(rows, func(a, b int) bool { return rows[a].Calls > rows[b].Calls })
	out.Providers = rows
	return out, nil
}

// List returns paginated trace records for the detail table.
func (i *imlTracingService) List(ctx context.Context, start, end int64, page, pageSize int) (*TraceList, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	total, err := i.traceStore.CountInWindow(ctx, start, end)
	if err != nil {
		return nil, err
	}
	rows, err := i.traceStore.ListInWindow(ctx, start, end, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	return &TraceList{Total: total, Page: page, PageSize: pageSize, Items: rows}, nil
}

// ---- DTOs ----

type Window struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type ProviderCall struct {
	Provider string `json:"provider"`
	Calls    int64  `json:"calls"`
}

type TraceOverview struct {
	Window       Window          `json:"window"`
	TotalCalls   int64           `json:"total_calls"`
	InputTokens  int64           `json:"input_tokens"`
	OutputTokens int64           `json:"output_tokens"`
	TotalTokens  int64           `json:"total_tokens"`
	Providers    []*ProviderCall `json:"providers"`
}

type TraceList struct {
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
	Items    []*metering_store.TraceEvent `json:"items"`
}
