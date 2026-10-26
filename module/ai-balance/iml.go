package ai_balance

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"

	ai_provider_local "github.com/RiemannChow/nantianmen-ai-gateway/ai-provider/local"

	model_runtime "github.com/RiemannChow/nantianmen-ai-gateway/ai-provider/model-runtime"

	"gorm.io/gorm"

	ai_key "github.com/RiemannChow/nantianmen-ai-gateway/service/ai-key"

	ai_api "github.com/RiemannChow/nantianmen-ai-gateway/service/ai-api"

	"github.com/google/uuid"

	ai_balance "github.com/RiemannChow/nantianmen-ai-gateway/service/ai-balance"

	"github.com/eolinker/go-common/store"

	"github.com/RiemannChow/nantianmen-ai-gateway/gateway"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/cluster"

	ai_balance_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/ai-balance/dto"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/go-common/register"
	"github.com/eolinker/go-common/server"
	"github.com/eolinker/go-common/utils"
)

var _ IBalanceModule = (*imlBalanceModule)(nil)

type imlBalanceModule struct {
	clusterService cluster.IClusterService    `autowired:""`
	aiAPIService   ai_api.IAPIService         `autowired:""`
	aiKeyService   ai_key.IKeyService         `autowired:""`
	balanceService ai_balance.IBalanceService `autowired:""`
	settingService setting.ISettingService    `autowired:""`
	transaction    store.ITransaction         `autowired:""`
	probeCancel    context.CancelFunc
}

func (i *imlBalanceModule) SyncLocalBalances(ctx context.Context, address string) error {
	releases, err := i.getLocalBalances(ctx, address)
	if err != nil {
		return err
	}
	return i.syncGateway(ctx, cluster.DefaultClusterID, releases, true)
}

func (i *imlBalanceModule) Create(ctx context.Context, input *ai_balance_dto.Create) error {
	has, err := i.balanceService.Exist(ctx, input.Provider, input.Model)
	if err != nil {
		return err
	}
	if has {
		return fmt.Errorf("model already exists")
	}
	priority, err := i.balanceService.MaxPriority(ctx)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		priority = 0
	}
	if input.Id == "" {
		input.Id = uuid.NewString()
	}
	providerName := ""
	modelName := ""
	base := ""
	switch input.Type {
	case ai_balance_dto.ModelTypeOnline:
		p, has := model_runtime.GetProvider(input.Provider)
		if !has {
			return fmt.Errorf("provider not found")
		}
		providerName = p.Name()
		modelName = input.Model
		base = fmt.Sprintf("%s://%s", p.URI().Scheme(), p.URI().Host())
	case ai_balance_dto.ModelTypeLocal:
		input.Provider = ai_provider_local.ProviderLocal
		providerName = ai_provider_local.ProviderLocal
		modelName = input.Model
		v, has := i.settingService.Get(ctx, "system.ai_model.ollama_address")
		if !has {
			return fmt.Errorf("ollama address not found")
		}
		base = v
	}

	return i.transaction.Transaction(ctx, func(ctx context.Context) error {
		err = i.balanceService.Create(ctx, &ai_balance.Create{
			Id:           input.Id,
			Priority:     priority + 1,
			Provider:     input.Provider,
			ProviderName: providerName,
			Model:        input.Model,
			ModelName:    modelName,
			Type:         ai_balance_dto.ModelType(input.Type).Int(),
		})
		if err != nil {
			return err
		}
		item, err := i.balanceService.Get(ctx, input.Id)
		if err != nil {
			return err
		}
		return i.syncGateway(ctx, cluster.DefaultClusterID, []*gateway.DynamicRelease{newRelease(item, base)}, true)
	})

}

func newRelease(item *ai_balance.Balance, base string) *gateway.DynamicRelease {

	cfg := make(map[string]interface{})
	cfg["provider"] = item.Provider
	cfg["model"] = item.Model
	cfg["model_config"] = ai_provider_local.LocalConfig
	cfg["base"] = base
	cfg["priority"] = item.Priority
	return &gateway.DynamicRelease{
		BasicItem: &gateway.BasicItem{
			ID:          item.Id,
			Description: item.ModelName,
			Resource:    "ai-provider",
			Version:     item.UpdateAt.Format("20060102150405"),
			MatchLabels: map[string]string{
				"module": "ai-provider",
			},
		},
		Attr: cfg,
	}
}

func (i *imlBalanceModule) Sort(ctx context.Context, input *ai_balance_dto.Sort) error {
	var list []*ai_balance.Balance
	var err error
	switch input.Sort {
	case "after":
		list, err = i.balanceService.SortAfter(ctx, input.Origin, input.Target)
	default:
		list, err = i.balanceService.SortBefore(ctx, input.Origin, input.Target)
	}
	if err != nil {
		return err
	}
	v, has := i.settingService.Get(ctx, "system.ai_model.ollama_address")
	if !has {
		return fmt.Errorf("ollama address not found")
	}
	releases := make([]*gateway.DynamicRelease, 0, len(list))
	for _, item := range list {
		if item.State != 1 {
			// abnormal models are excluded from the fan-out
			continue
		}
		base := v
		if item.Provider != ai_provider_local.ProviderLocal {
			p, has := model_runtime.GetProvider(item.Provider)
			if !has {
				continue
			}
			base = fmt.Sprintf("%s://%s", p.URI().Scheme(), p.URI().Host())
		}

		releases = append(releases, newRelease(item, base))
	}
	err = i.syncGateway(ctx, cluster.DefaultClusterID, releases, true)
	if err != nil {
		return err
	}
	return nil
}

func (i *imlBalanceModule) List(ctx context.Context, keyword string) ([]*ai_balance_dto.Item, error) {
	list, err := i.balanceService.Search(ctx, keyword, nil, "priority asc")
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Priority < list[j].Priority
	})
	aiAPIMap, err := i.aiAPIService.CountMapByModel(ctx, "", nil)
	if err != nil {
		return nil, fmt.Errorf("get ai api count error:%v", err)
	}
	keyMap, err := i.aiKeyService.CountMapByProvider(ctx, "", nil)
	if err != nil {
		return nil, fmt.Errorf("get ai key count error:%v", err)
	}
	result := make([]*ai_balance_dto.Item, 0, len(list))
	for i, item := range list {
		priority := i + 1
		result = append(result, &ai_balance_dto.Item{
			Id: item.Id,
			Provider: &ai_balance_dto.BasicItem{
				Id:   item.Provider,
				Name: item.ProviderName,
			},
			Model: &ai_balance_dto.BasicItem{
				Id:   item.Model,
				Name: item.ModelName,
			},
			Priority:      priority,
			Type:          ai_balance_dto.ModelTypeFromInt(item.Type),
			State:         ai_balance_dto.ModelStateFromInt(item.State),
			APICount:      aiAPIMap[item.Model],
			KeyCount:      keyMap[item.Provider],
			LastCheckAt:   item.LastCheckAt,
			LastSuccessAt: item.LastSuccessAt,
			FailCount:     item.FailCount,
		})
	}
	return result, nil
}

func (i *imlBalanceModule) Delete(ctx context.Context, id string) error {
	return i.transaction.Transaction(ctx, func(ctx context.Context) error {
		err := i.balanceService.Delete(ctx, id)
		if err != nil {
			return err
		}
		return i.syncGateway(ctx, cluster.DefaultClusterID, []*gateway.DynamicRelease{
			{
				BasicItem: &gateway.BasicItem{
					ID:       id,
					Resource: "ai-provider",
				},
			},
		}, false)
	})

}

func (i *imlBalanceModule) syncGateway(ctx context.Context, clusterId string, releases []*gateway.DynamicRelease, online bool) error {
	client, err := i.clusterService.GatewayClient(ctx, clusterId)
	if err != nil {
		log.Errorf("get apinto client error: %v", err)
		return nil
	}
	defer func() {
		err := client.Close(ctx)
		if err != nil {
			log.Warn("close apinto client:", err)
		}
	}()
	for _, releaseInfo := range releases {
		dynamicClient, err := client.Dynamic(releaseInfo.Resource)
		if err != nil {
			return err
		}
		if online {
			err = dynamicClient.Online(ctx, releaseInfo)
		} else {
			dynamicClient.Offline(ctx, releaseInfo)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *imlBalanceModule) getLocalBalances(ctx context.Context, v string) ([]*gateway.DynamicRelease, error) {
	balances, err := i.balanceService.Search(ctx, "", map[string]interface{}{"provider": ai_provider_local.ProviderLocal}, "priority asc")
	if err != nil {
		return nil, err
	}
	if v == "" {
		var has bool
		v, has = i.settingService.Get(ctx, "system.ai_model.ollama_address")
		if !has {
			//return nil, fmt.Errorf("ollama address not found")
			return nil, nil
		}
	}

	releases := make([]*gateway.DynamicRelease, 0, len(balances))
	for _, item := range balances {
		if item.State != 1 {
			// abnormal models are excluded from the fan-out
			continue
		}
		base := v
		if item.Provider != ai_provider_local.ProviderLocal {
			p, has := model_runtime.GetProvider(item.Provider)
			if !has {
				continue
			}
			base = fmt.Sprintf("%s://%s%s", p.URI().Scheme(), p.URI().Host(), p.URI().Path())
		}
		releases = append(releases, newRelease(item, base))
	}
	return releases, nil
}

func (i *imlBalanceModule) getBalances(ctx context.Context) ([]*gateway.DynamicRelease, error) {
	balances, err := i.balanceService.Search(ctx, "", nil, "priority asc")
	if err != nil {
		return nil, err
	}
	v, has := i.settingService.Get(ctx, "system.ai_model.ollama_address")
	if !has {
		//return nil, fmt.Errorf("ollama address not found")
		return nil, nil
	}
	releases := make([]*gateway.DynamicRelease, 0, len(balances))
	for _, item := range balances {
		if item.State != 1 {
			// abnormal models are excluded from the fan-out
			continue
		}
		base := v
		if item.Provider != ai_provider_local.ProviderLocal {
			p, has := model_runtime.GetProvider(item.Provider)
			if !has {
				continue
			}
			base = fmt.Sprintf("%s://%s%s", p.URI().Scheme(), p.URI().Host(), p.URI().Path())
		} else {
			if v == "" {
				continue
			}
		}
		releases = append(releases, newRelease(item, base))
	}
	return releases, nil
}

func (i *imlBalanceModule) initGateway(ctx context.Context, clusterId string, clientDriver gateway.IClientDriver) error {
	releases, err := i.getBalances(ctx)
	if err != nil {
		return err
	}
	for _, p := range releases {
		client, err := clientDriver.Dynamic(p.Resource)
		if err != nil {
			return err
		}
		err = client.Online(ctx, p)
		if err != nil {
			return err
		}
	}
	return nil
}

// OnInit starts the periodic health-probe loop once the server is ready.
func (i *imlBalanceModule) OnInit() {
	register.Handle(func(v server.Server) {
		ctx, cancel := context.WithCancel(context.Background())
		ctx = utils.SetUserId(ctx, "admin")
		i.probeCancel = cancel
		i.startHealthProbe(ctx)
	})
}

// baseURL resolves the probe target for a balance entry: the ollama address for
// local models, or the provider base URL for online models. Empty string means
// the target cannot be resolved (skip probing).
func (i *imlBalanceModule) baseURL(ctx context.Context, item *ai_balance.Balance) string {
	if item.Provider == ai_provider_local.ProviderLocal {
		v, has := i.settingService.Get(ctx, "system.ai_model.ollama_address")
		if !has {
			return ""
		}
		return v
	}
	p, has := model_runtime.GetProvider(item.Provider)
	if !has {
		return ""
	}
	return fmt.Sprintf("%s://%s%s", p.URI().Scheme(), p.URI().Host(), p.URI().Path())
}

// ProbeModel probes a single model and persists its health, flipping its gateway
// online/offline state on change.
func (i *imlBalanceModule) ProbeModel(ctx context.Context, id string) error {
	item, err := i.balanceService.Get(ctx, id)
	if err != nil {
		return err
	}
	healthy := i.probeOne(ctx, item)
	i.applyHealth(ctx, item, healthy)
	return nil
}

// BatchProbe probes every model regardless of the auto-failover setting (manual
// trigger from the UI).
func (i *imlBalanceModule) BatchProbe(ctx context.Context) error {
	i.probeAll(ctx)
	return nil
}

// UpdateState manually overrides a model's health state and re-syncs the gateway.
func (i *imlBalanceModule) UpdateState(ctx context.Context, id string, state int) error {
	item, err := i.balanceService.Get(ctx, id)
	if err != nil {
		return err
	}
	if item.State == state {
		return nil
	}
	now := time.Now().Unix()
	err = i.balanceService.UpdateHealth(ctx, id, state, now, item.LastSuccessAt, int64(item.FailCount))
	if err != nil {
		return err
	}
	base := i.baseURL(ctx, item)
	if base != "" {
		return i.syncGateway(ctx, cluster.DefaultClusterID, []*gateway.DynamicRelease{newRelease(item, base)}, state == 1)
	}
	return nil
}

// GetAutoFailover reports whether the scheduled automatic failover is enabled.
func (i *imlBalanceModule) GetAutoFailover(ctx context.Context) (bool, error) {
	v, has := i.settingService.Get(ctx, autoFailoverKey)
	if !has {
		return false, nil
	}
	return v == "true", nil
}

// SetAutoFailover enables or disables the scheduled automatic failover.
func (i *imlBalanceModule) SetAutoFailover(ctx context.Context, enabled bool) error {
	v := "false"
	if enabled {
		v = "true"
	}
	return i.settingService.Set(ctx, autoFailoverKey, v, "admin")
}
