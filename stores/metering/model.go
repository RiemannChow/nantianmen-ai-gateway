package metering

import "time"

// UsageEvent is one normalized AI/API call record collected by the gateway's
// local metering pipeline. It is the control plane's own copy of usage data,
// which removes the hard dependency on an external InfluxDB for the FinOps suite
// (billing, analytics, cost optimization). Apinto (or any collector) pushes
// access-log events here through the metering ingest API.
type UsageEvent struct {
	Id           int64     `json:"id" gorm:"column:id;type:BIGINT(20);AUTO_INCREMENT;NOT NULL;primary_key;comment:主键ID"`
	ClusterId    string    `json:"cluster_id" gorm:"size:64;not null;default:'';column:cluster_id;index:idx_usage_cluster;comment:集群ID"`
	ServiceId    string    `json:"service_id" gorm:"size:128;not null;default:'';column:service_id;index:idx_usage_service;comment:服务ID"`
	Provider     string    `json:"provider" gorm:"size:128;not null;default:'';column:provider;index:idx_usage_provider;comment:模型供应商/上游名称"`
	Subscriber   string    `json:"subscriber" gorm:"size:128;not null;default:'';column:subscriber;comment:调用方/消费者"`
	Calls        int64     `json:"calls" gorm:"not null;default:1;column:calls;comment:调用次数"`
	InputTokens  int64     `json:"input_tokens" gorm:"not null;default:0;column:input_tokens;comment:输入Token"`
	OutputTokens int64     `json:"output_tokens" gorm:"not null;default:0;column:output_tokens;comment:输出Token"`
	StatusCode   int       `json:"status_code" gorm:"not null;default:200;column:status_code;comment:HTTP状态码"`
	LatencyMs    int64     `json:"latency_ms" gorm:"not null;default:0;column:latency_ms;comment:延迟(ms)"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;column:created_at;index:idx_usage_created_at;comment:调用时间"`
}

func (u *UsageEvent) IdValue() int64 {
	return u.Id
}

func (u *UsageEvent) TableName() string {
	return "apinto_usage_event"
}

// ProviderStat is the per-provider aggregated call count for a time window.
type ProviderStat struct {
	Provider string
	Calls    int64
}
