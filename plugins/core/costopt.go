package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"
	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) costOptApis() []pm3.Api {
	return []pm3.Api{
		// budget
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/budget", []string{"context"}, []string{"budget"}, p.costOptController.GetBudget, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/costopt/budget", []string{"context", "body"}, []string{"budget"}, p.costOptController.SaveBudget, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/budget/status", []string{"context"}, []string{"status"}, p.costOptController.BudgetStatus, access.SystemSettingsApiGatewayView),
		// model cost profiles
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/providers", []string{"context"}, []string{"providers"}, p.costOptController.GetProviders, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/costopt/providers", []string{"context", "body"}, []string{"providers"}, p.costOptController.SaveProviders, access.SystemSettingsApiGatewayManager),
		// substitution map
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/substitutes", []string{"context"}, []string{"substitutes"}, p.costOptController.GetSubstitutes, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/costopt/substitutes", []string{"context", "body"}, []string{"substitutes"}, p.costOptController.SaveSubstitutes, access.SystemSettingsApiGatewayManager),
		// judge config
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/judge", []string{"context"}, []string{"judge"}, p.costOptController.GetJudge, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/costopt/judge", []string{"context", "body"}, []string{"judge"}, p.costOptController.SaveJudge, access.SystemSettingsApiGatewayManager),
		// analytics / recommendations / anomalies
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/analytics", []string{"context", "query:start", "query:end"}, []string{"analytics"}, p.costOptController.Analytics, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/recommend", []string{"context", "query:start", "query:end"}, []string{"recommend"}, p.costOptController.Recommend, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/anomalies", []string{"context", "query:start", "query:end"}, []string{"anomalies"}, p.costOptController.Anomalies, access.SystemSettingsApiGatewayView),
		// LLM-as-judge evaluation
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/costopt/evaluate", []string{"context", "body"}, []string{"evaluation"}, p.costOptController.Evaluate, access.SystemSettingsApiGatewayManager),
		// notification delivery
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/notify", []string{"context"}, []string{"notify"}, p.costOptController.GetNotifyConfig, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/costopt/notify", []string{"context", "body"}, []string{"notify"}, p.costOptController.SaveNotifyConfig, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/costopt/notify/test", []string{"context", "query:channel"}, []string{}, p.costOptController.SendTestNotification, access.SystemSettingsApiGatewayManager),
		// per-consumer / per-key budget & quota
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/consumer_budgets", []string{"context"}, []string{"consumer_budgets"}, p.costOptController.GetConsumerBudgets, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/costopt/consumer_budgets", []string{"context", "body"}, []string{"consumer_budgets"}, p.costOptController.SaveConsumerBudgets, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/costopt/consumer/:consumer/status", []string{"context", "path:consumer"}, []string{"status"}, p.costOptController.ConsumerBudgetStatus, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/costopt/consumer/:consumer/quota", []string{"context", "path:consumer"}, []string{}, p.costOptController.PublishConsumerQuota, access.SystemSettingsApiGatewayManager),
	}
}
