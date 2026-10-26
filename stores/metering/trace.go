package metering

import (
	"context"
	"reflect"
	"time"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/store"
)

var _ ITraceStore = (*imlTraceStore)(nil)

// ITraceStore persists per-LLM-call traces (prompt/response, token detail,
// latency breakdown). It is the local-first store that powers the tracing /
// LLM observability module and removes the dependency on external InfluxDB.
type ITraceStore interface {
	Insert(ctx context.Context, t ...*TraceEvent) error
	CountInWindow(ctx context.Context, start, end int64) (int64, error)
	SummaryInWindow(ctx context.Context, start, end int64) (calls, inTokens, outTokens int64, err error)
	ListInWindow(ctx context.Context, start, end int64, limit, offset int) ([]*TraceEvent, error)
	ByProviderInWindow(ctx context.Context, start, end int64, limit int) ([]*ProviderStat, error)
	DeleteAll(ctx context.Context) error
}

type imlTraceStore struct {
	store.Store[TraceEvent]
}

func init() {
	autowire.Auto[ITraceStore](func() reflect.Value {
		return reflect.ValueOf(new(imlTraceStore))
	})
}

func (s *imlTraceStore) window() string {
	return "created_at >= FROM_UNIXTIME(?) AND created_at <= FROM_UNIXTIME(?)"
}

func (s *imlTraceStore) Insert(ctx context.Context, t ...*TraceEvent) error {
	if len(t) == 0 {
		return nil
	}
	return s.DB(ctx).Create(t).Error
}

func (s *imlTraceStore) CountInWindow(ctx context.Context, start, end int64) (int64, error) {
	var n int64
	err := s.DB(ctx).Model(&TraceEvent{}).Where(s.window(), start, end).Count(&n).Error
	return n, err
}

func (s *imlTraceStore) SummaryInWindow(ctx context.Context, start, end int64) (int64, int64, int64, error) {
	type agg struct {
		Calls  int64 `gorm:"column:calls"`
		Input  int64 `gorm:"column:input"`
		Output int64 `gorm:"column:output"`
	}
	var a agg
	err := s.DB(ctx).Model(&TraceEvent{}).
		Select("COALESCE(COUNT(*),0) AS calls, COALESCE(SUM(prompt_tokens),0) AS input, COALESCE(SUM(completion_tokens),0) AS output").
		Where(s.window(), start, end).
		Scan(&a).Error
	return a.Calls, a.Input, a.Output, err
}

func (s *imlTraceStore) ListInWindow(ctx context.Context, start, end int64, limit, offset int) ([]*TraceEvent, error) {
	db := s.DB(ctx).Model(&TraceEvent{}).Where(s.window(), start, end).Order("created_at DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}
	var rows []*TraceEvent
	err := db.Find(&rows).Error
	return rows, err
}

func (s *imlTraceStore) ByProviderInWindow(ctx context.Context, start, end int64, limit int) ([]*ProviderStat, error) {
	type row struct {
		Provider string `gorm:"column:provider"`
		Calls    int64  `gorm:"column:calls"`
	}
	db := s.DB(ctx).Model(&TraceEvent{}).
		Select("provider, COALESCE(COUNT(*),0) AS calls").
		Where(s.window(), start, end).
		Group("provider").Order("calls DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	rows, err := db.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*ProviderStat, 0)
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Provider, &r.Calls); err != nil {
			return nil, err
		}
		out = append(out, &ProviderStat{Provider: r.Provider, Calls: r.Calls})
	}
	return out, nil
}

func (s *imlTraceStore) DeleteAll(ctx context.Context) error {
	return s.DB(ctx).Where("1=1").Delete(&TraceEvent{}).Error
}

// TraceEvent is one LLM call record captured by the tracing pipeline. It carries
// the full prompt/completion, model, token detail and a latency breakdown so the
// control plane can troubleshoot and attribute cost per call.
type TraceEvent struct {
	Id               int64     `json:"id" gorm:"column:id;type:BIGINT(20);AUTO_INCREMENT;NOT NULL;primary_key;comment:主键ID"`
	ClusterId        string    `json:"cluster_id" gorm:"size:64;not null;default:'';column:cluster_id;index:idx_trace_cluster;comment:集群ID"`
	ServiceId        string    `json:"service_id" gorm:"size:128;not null;default:'';column:service_id;index:idx_trace_service;comment:服务ID"`
	Provider         string    `json:"provider" gorm:"size:128;not null;default:'';column:provider;index:idx_trace_provider;comment:模型供应商"`
	Subscriber       string    `json:"subscriber" gorm:"size:128;not null;default:'';column:subscriber;comment:调用方/消费者"`
	Model            string    `json:"model" gorm:"size:128;not null;default:'';column:model;comment:模型"`
	Prompt           string    `json:"prompt" gorm:"type:text;column:prompt;comment:输入提示词"`
	Completion       string    `json:"completion" gorm:"type:text;column:completion;comment:模型输出"`
	FinishReason     string    `json:"finish_reason" gorm:"size:32;not null;default:'';column:finish_reason;comment:结束原因"`
	PromptTokens     int64     `json:"prompt_tokens" gorm:"not null;default:0;column:prompt_tokens;comment:输入Token"`
	CompletionTokens int64     `json:"completion_tokens" gorm:"not null;default:0;column:completion_tokens;comment:输出Token"`
	TotalTokens      int64     `json:"total_tokens" gorm:"not null;default:0;column:total_tokens;comment:总Token"`
	StatusCode       int       `json:"status_code" gorm:"not null;default:200;column:status_code;comment:HTTP状态码"`
	LatencyTotalMs   int64     `json:"latency_total_ms" gorm:"not null;default:0;column:latency_total_ms;comment:总延迟(ms)"`
	LatencyTTFTMs    int64     `json:"latency_ttft_ms" gorm:"not null;default:0;column:latency_ttft_ms;comment:首字延迟TTFT(ms)"`
	LatencyGenMs     int64     `json:"latency_gen_ms" gorm:"not null;default:0;column:latency_gen_ms;comment:生成延迟(ms)"`
	CreatedAt        time.Time `json:"created_at" gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;column:created_at;index:idx_trace_created_at;comment:调用时间"`
}

func (t *TraceEvent) IdValue() int64 {
	return t.Id
}

func (t *TraceEvent) TableName() string {
	return "apinto_trace_event"
}
