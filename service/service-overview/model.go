package service_overview

import (
	"github.com/RiemannChow/nantianmen-ai-gateway/stores/service"
)

type Overview struct {
	Service         string
	ApiCount        int64
	ReleaseApiCount int64
	IsReleased      bool
}

func FromEntity(e *service.Overview) *Overview {
	return &Overview{
		Service:         e.Service,
		ApiCount:        e.ApiCount,
		ReleaseApiCount: e.ReleaseApiCount,
		IsReleased:      e.IsReleased,
	}
}

type Update struct {
	ApiCount        *int64
	ReleaseApiCount *int64
	IsReleased      *bool
}
