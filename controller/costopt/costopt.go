package costopt

import (
	"reflect"
	"strconv"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/costopt"
	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type ProvidersInput struct {
	Providers map[string]costopt.ProviderCost `json:"providers"`
}

type SubstitutesInput struct {
	Substitutes map[string]string `json:"substitutes"`
}

type ICostOptController interface {
	GetBudget(ctx *gin.Context) (*costopt.BudgetConfig, error)
	SaveBudget(ctx *gin.Context, cfg *costopt.BudgetConfig) (*costopt.BudgetConfig, error)
	BudgetStatus(ctx *gin.Context) (*costopt.BudgetStatus, error)
	GetProviders(ctx *gin.Context) (*ProvidersInput, error)
	SaveProviders(ctx *gin.Context, input *ProvidersInput) (*ProvidersInput, error)
	GetSubstitutes(ctx *gin.Context) (*SubstitutesInput, error)
	SaveSubstitutes(ctx *gin.Context, input *SubstitutesInput) (*SubstitutesInput, error)
	GetJudge(ctx *gin.Context) (*costopt.JudgeConfig, error)
	SaveJudge(ctx *gin.Context, cfg *costopt.JudgeConfig) (*costopt.JudgeConfig, error)
	Recommend(ctx *gin.Context, start string, end string) (*costopt.RecommendationResult, error)
	Analytics(ctx *gin.Context, start string, end string) (*costopt.AnalyticsResult, error)
	Anomalies(ctx *gin.Context, start string, end string) (*costopt.AnomalyResult, error)
	Evaluate(ctx *gin.Context, input *costopt.EvaluateInput) (*costopt.EvaluateResult, error)
	GetNotifyConfig(ctx *gin.Context) (*costopt.NotifyConfig, error)
	SaveNotifyConfig(ctx *gin.Context, cfg *costopt.NotifyConfig) (*costopt.NotifyConfig, error)
	SendTestNotification(ctx *gin.Context, channel string) error
	GetConsumerBudgets(ctx *gin.Context) (*costopt.ConsumerBudgetConfig, error)
	SaveConsumerBudgets(ctx *gin.Context, cfg *costopt.ConsumerBudgetConfig) (*costopt.ConsumerBudgetConfig, error)
	ConsumerBudgetStatus(ctx *gin.Context, consumer string) (*costopt.ConsumerBudgetStatus, error)
	PublishConsumerQuota(ctx *gin.Context, consumer string) error
}

func init() {
	autowire.Auto[ICostOptController](func() reflect.Value {
		return reflect.ValueOf(new(imlCostOptController))
	})
}

var _ ICostOptController = (*imlCostOptController)(nil)

type imlCostOptController struct {
	service costopt.ICostOptService `autowired:""`
}

func gate(ctx *gin.Context) error {
	return license.Check(ctx, "cost_optimization")
}

func (i *imlCostOptController) GetBudget(ctx *gin.Context) (*costopt.BudgetConfig, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return i.service.GetBudget(ctx), nil
}

func (i *imlCostOptController) SaveBudget(ctx *gin.Context, cfg *costopt.BudgetConfig) (*costopt.BudgetConfig, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &costopt.BudgetConfig{}
	}
	return i.service.SaveBudget(ctx, cfg)
}

func (i *imlCostOptController) BudgetStatus(ctx *gin.Context) (*costopt.BudgetStatus, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return i.service.BudgetStatus(ctx)
}

func (i *imlCostOptController) GetProviders(ctx *gin.Context) (*ProvidersInput, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return &ProvidersInput{Providers: i.service.GetProviders(ctx)}, nil
}

func (i *imlCostOptController) SaveProviders(ctx *gin.Context, input *ProvidersInput) (*ProvidersInput, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	if input == nil {
		input = &ProvidersInput{Providers: map[string]costopt.ProviderCost{}}
	}
	p, err := i.service.SaveProviders(ctx, input.Providers)
	if err != nil {
		return nil, err
	}
	return &ProvidersInput{Providers: p}, nil
}

func (i *imlCostOptController) GetSubstitutes(ctx *gin.Context) (*SubstitutesInput, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return &SubstitutesInput{Substitutes: i.service.GetSubstitutes(ctx)}, nil
}

func (i *imlCostOptController) SaveSubstitutes(ctx *gin.Context, input *SubstitutesInput) (*SubstitutesInput, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	if input == nil {
		input = &SubstitutesInput{Substitutes: map[string]string{}}
	}
	s, err := i.service.SaveSubstitutes(ctx, input.Substitutes)
	if err != nil {
		return nil, err
	}
	return &SubstitutesInput{Substitutes: s}, nil
}

func (i *imlCostOptController) GetJudge(ctx *gin.Context) (*costopt.JudgeConfig, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return i.service.GetJudge(ctx), nil
}

func (i *imlCostOptController) SaveJudge(ctx *gin.Context, cfg *costopt.JudgeConfig) (*costopt.JudgeConfig, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &costopt.JudgeConfig{}
	}
	return i.service.SaveJudge(ctx, cfg)
}

func (i *imlCostOptController) Recommend(ctx *gin.Context, start string, end string) (*costopt.RecommendationResult, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	s, e := defaultWindow(start, end)
	return i.service.Recommend(ctx, s, e, 0.3)
}

func (i *imlCostOptController) Analytics(ctx *gin.Context, start string, end string) (*costopt.AnalyticsResult, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	s, e := defaultWindow(start, end)
	return i.service.Analytics(ctx, s, e)
}

func (i *imlCostOptController) Anomalies(ctx *gin.Context, start string, end string) (*costopt.AnomalyResult, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	s, e := defaultWindow(start, end)
	return i.service.Anomalies(ctx, s, e)
}

func (i *imlCostOptController) Evaluate(ctx *gin.Context, input *costopt.EvaluateInput) (*costopt.EvaluateResult, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return i.service.Evaluate(ctx, input)
}

func (i *imlCostOptController) GetNotifyConfig(ctx *gin.Context) (*costopt.NotifyConfig, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return i.service.GetNotifyConfig(ctx), nil
}

func (i *imlCostOptController) SaveNotifyConfig(ctx *gin.Context, cfg *costopt.NotifyConfig) (*costopt.NotifyConfig, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &costopt.NotifyConfig{}
	}
	return i.service.SaveNotifyConfig(ctx, cfg)
}

func (i *imlCostOptController) SendTestNotification(ctx *gin.Context, channel string) error {
	if err := gate(ctx); err != nil {
		return err
	}
	return i.service.SendTestNotification(ctx, channel)
}

func (i *imlCostOptController) GetConsumerBudgets(ctx *gin.Context) (*costopt.ConsumerBudgetConfig, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return i.service.GetConsumerBudgets(ctx)
}

func (i *imlCostOptController) SaveConsumerBudgets(ctx *gin.Context, cfg *costopt.ConsumerBudgetConfig) (*costopt.ConsumerBudgetConfig, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &costopt.ConsumerBudgetConfig{Items: map[string]costopt.ConsumerBudgetItem{}}
	}
	return i.service.SaveConsumerBudgets(ctx, cfg)
}

func (i *imlCostOptController) ConsumerBudgetStatus(ctx *gin.Context, consumer string) (*costopt.ConsumerBudgetStatus, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return i.service.ConsumerBudgetStatus(ctx, consumer)
}

func (i *imlCostOptController) PublishConsumerQuota(ctx *gin.Context, consumer string) error {
	if err := gate(ctx); err != nil {
		return err
	}
	return i.service.PublishConsumerQuota(ctx, consumer)
}

func defaultWindow(start string, end string) (int64, int64) {
	now := time.Now().Unix()
	s, e := now-(30*24*3600), now
	if v, err := strconv.ParseInt(start, 10, 64); err == nil {
		s = v
	}
	if v, err := strconv.ParseInt(end, 10, 64); err == nil {
		e = v
	}
	if e < s {
		e = s
	}
	return s, e
}
