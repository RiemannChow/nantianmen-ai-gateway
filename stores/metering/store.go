package metering

import (
	"context"
	"reflect"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/store"
)

var _ IUsageStore = (*imlUsageStore)(nil)

// IUsageStore is the persistence + aggregation interface for usage events.
type IUsageStore interface {
	Insert(ctx context.Context, t ...*UsageEvent) error
	// CountInWindow returns the number of usage events in [start, end] (unix sec).
	CountInWindow(ctx context.Context, start, end int64) (int64, error)
	// SummaryInWindow returns total calls, input tokens, output tokens in window.
	SummaryInWindow(ctx context.Context, start, end int64) (calls, input, output int64, err error)
	// StatusBreakdownInWindow returns 2xx/4xx/5xx call counts in window.
	StatusBreakdownInWindow(ctx context.Context, start, end int64) (c2xx, c4xx, c5xx int64, err error)
	// ByProviderInWindow returns per-provider call counts (desc), capped at limit.
	ByProviderInWindow(ctx context.Context, start, end int64, limit int) ([]*ProviderStat, error)
	// SummaryBySubscriberInWindow returns total calls/input/output tokens for a
	// single subscriber (== metering consumer key) in window.
	SummaryBySubscriberInWindow(ctx context.Context, subscriber string, start, end int64) (calls, input, output int64, err error)
	// DeleteAll removes every usage event (used by re-seed).
	DeleteAll(ctx context.Context) error
}

type imlUsageStore struct {
	store.Store[UsageEvent]
}

func init() {
	autowire.Auto[IUsageStore](func() reflect.Value {
		return reflect.ValueOf(new(imlUsageStore))
	})
}

func (s *imlUsageStore) window() string {
	return "created_at >= FROM_UNIXTIME(?) AND created_at <= FROM_UNIXTIME(?)"
}

func (s *imlUsageStore) CountInWindow(ctx context.Context, start, end int64) (int64, error) {
	var n int64
	err := s.DB(ctx).Model(&UsageEvent{}).Where(s.window(), start, end).Count(&n).Error
	return n, err
}

func (s *imlUsageStore) SummaryInWindow(ctx context.Context, start, end int64) (int64, int64, int64, error) {
	type agg struct {
		Calls  int64 `gorm:"column:calls"`
		Input  int64 `gorm:"column:input"`
		Output int64 `gorm:"column:output"`
	}
	var a agg
	err := s.DB(ctx).Model(&UsageEvent{}).
		Select("COALESCE(SUM(calls),0) AS calls, COALESCE(SUM(input_tokens),0) AS input, COALESCE(SUM(output_tokens),0) AS output").
		Where(s.window(), start, end).
		Scan(&a).Error
	return a.Calls, a.Input, a.Output, err
}

func (s *imlUsageStore) StatusBreakdownInWindow(ctx context.Context, start, end int64) (int64, int64, int64, error) {
	type row struct {
		Code int   `gorm:"column:status_code"`
		Cnt  int64 `gorm:"column:cnt"`
	}
	rows, err := s.DB(ctx).Model(&UsageEvent{}).
		Select("status_code, COALESCE(SUM(calls),0) AS cnt").
		Where(s.window(), start, end).
		Group("status_code").Rows()
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()
	var c2xx, c4xx, c5xx int64
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Code, &r.Cnt); err != nil {
			return 0, 0, 0, err
		}
		switch {
		case r.Code >= 200 && r.Code < 300:
			c2xx += r.Cnt
		case r.Code >= 400 && r.Code < 500:
			c4xx += r.Cnt
		case r.Code >= 500:
			c5xx += r.Cnt
		}
	}
	return c2xx, c4xx, c5xx, nil
}

func (s *imlUsageStore) ByProviderInWindow(ctx context.Context, start, end int64, limit int) ([]*ProviderStat, error) {
	type row struct {
		Provider string `gorm:"column:provider"`
		Calls    int64  `gorm:"column:calls"`
	}
	db := s.DB(ctx).Model(&UsageEvent{}).
		Select("provider, COALESCE(SUM(calls),0) AS calls").
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

func (s *imlUsageStore) SummaryBySubscriberInWindow(ctx context.Context, subscriber string, start, end int64) (int64, int64, int64, error) {
	type agg struct {
		Calls  int64 `gorm:"column:calls"`
		Input  int64 `gorm:"column:input"`
		Output int64 `gorm:"column:output"`
	}
	var a agg
	err := s.DB(ctx).Model(&UsageEvent{}).
		Select("COALESCE(SUM(calls),0) AS calls, COALESCE(SUM(input_tokens),0) AS input, COALESCE(SUM(output_tokens),0) AS output").
		Where(s.window(), start, end).
		Where("subscriber = ?", subscriber).
		Scan(&a).Error
	return a.Calls, a.Input, a.Output, err
}

func (s *imlUsageStore) DeleteAll(ctx context.Context) error {
	return s.DB(ctx).Where("1=1").Delete(&UsageEvent{}).Error
}
