package marketplace

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"
	"github.com/RiemannChow/nantianmen-ai-gateway/stores/marketplace"

	"github.com/google/uuid"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"
)

const revenueShareKey = "marketplace.revenue_share"

// CatalogItem is a single entry in the plugin marketplace directory.
type CatalogItem struct {
	Id               string  `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Price            string  `json:"price"`
	RevenueShare     float64 `json:"revenue_share"`
	LicenseModuleKey string  `json:"license_module_key"`
	Category         string  `json:"category"`
	Icon             string  `json:"icon"`
	Version          string  `json:"version"`
}

// InstalledView is an installed catalog item enriched with its runtime status.
type InstalledView struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Version          string `json:"version"`
	LicenseModuleKey string `json:"license_module_key"`
	Category         string `json:"category"`
	Icon             string `json:"icon"`
	Enabled          bool   `json:"enabled"`
}

// staticCatalog is the built-in plugin directory. Operators may install any of
// these; enabling one requires its LicenseModuleKey to be entitled.
var staticCatalog = []CatalogItem{
	{Id: "advanced_policy", Name: "高级策略包", Description: "精细限流、熔断与灰度发布能力。", Price: "按节点订阅", RevenueShare: 0.2, LicenseModuleKey: "advanced_policy", Category: "策略", Icon: "policy", Version: "1.0.0"},
	{Id: "sso", Name: "SSO / LDAP 单点登录", Description: "对接企业身份提供商，统一登录。", Price: "按席位订阅", RevenueShare: 0.15, LicenseModuleKey: "sso", Category: "身份", Icon: "sso", Version: "1.0.0"},
	{Id: "audit", Name: "审计与合规报表", Description: "全量操作审计与合规导出。", Price: "按模块订阅", RevenueShare: 0.2, LicenseModuleKey: "audit", Category: "合规", Icon: "audit", Version: "1.0.0"},
	{Id: "multi_cluster", Name: "多集群容备", Description: "跨集群高可用与故障转移。", Price: "按集群订阅", RevenueShare: 0.25, LicenseModuleKey: "multi_cluster", Category: "高可用", Icon: "cluster", Version: "1.0.0"},
	{Id: "ai_security", Name: "AI 安全与内容防护", Description: "提示词注入检测与敏感内容拦截。", Price: "按调用量", RevenueShare: 0.2, LicenseModuleKey: "ai_security", Category: "安全", Icon: "security", Version: "1.0.0"},
	{Id: "tracing", Name: "链路追踪与可观测性", Description: "LLM 调用全链路追踪。", Price: "按调用量", RevenueShare: 0.15, LicenseModuleKey: "tracing", Category: "可观测", Icon: "tracing", Version: "1.0.0"},
}

// IMarketplaceModule exposes the catalog and install/enable management.
type IMarketplaceModule interface {
	Catalog(ctx context.Context) ([]*InstalledView, error)
	Install(ctx context.Context, id string) error
	Enable(ctx context.Context, id string, enabled bool) error
	GetRevenueShare(ctx context.Context) (map[string]float64, error)
	SetRevenueShare(ctx context.Context, shares map[string]float64) error
}

func init() {
	autowire.Auto[IMarketplaceModule](func() reflect.Value {
		return reflect.ValueOf(new(imlMarketplaceModule))
	})
}

var _ IMarketplaceModule = (*imlMarketplaceModule)(nil)

type imlMarketplaceModule struct {
	store          marketplace.IInstalledStore `autowired:""`
	settingService setting.ISettingService     `autowired:""`
}

func (i *imlMarketplaceModule) findCatalog(id string) (*CatalogItem, bool) {
	for idx := range staticCatalog {
		if staticCatalog[idx].Id == id {
			return &staticCatalog[idx], true
		}
	}
	return nil, false
}

func (i *imlMarketplaceModule) Catalog(ctx context.Context) ([]*InstalledView, error) {
	installed, err := i.store.Search(ctx, "", nil, "plugin_id asc")
	if err != nil {
		return nil, err
	}
	byId := make(map[string]*marketplace.InstalledPlugin)
	for _, it := range installed {
		byId[it.PluginId] = it
	}
	list := make([]*InstalledView, 0, len(staticCatalog))
	for idx := range staticCatalog {
		item := &staticCatalog[idx]
		view := &InstalledView{
			Id:               item.Id,
			Name:             item.Name,
			Description:      item.Description,
			Version:          item.Version,
			LicenseModuleKey: item.LicenseModuleKey,
			Category:         item.Category,
			Icon:             item.Icon,
		}
		if ins, ok := byId[item.Id]; ok {
			view.Enabled = ins.Enabled
		}
		list = append(list, view)
	}
	return list, nil
}

func (i *imlMarketplaceModule) Install(ctx context.Context, id string) error {
	item, ok := i.findCatalog(id)
	if !ok {
		return errors.New("plugin not found in catalog")
	}
	// Installing enables the underlying enterprise capability; enforce the
	// entitlement so an unlicensed module cannot be activated here.
	if err := license.Check(ctx, item.LicenseModuleKey); err != nil {
		return err
	}
	existing, err := i.store.Search(ctx, "", map[string]interface{}{"plugin_id": id}, "")
	if err != nil {
		return err
	}
	operator := utils.UserId(ctx)
	now := time.Now()
	if len(existing) > 0 {
		existing[0].Updater = operator
		existing[0].UpdateAt = now
		existing[0].Version = item.Version
		_, err = i.store.Update(ctx, existing[0])
		return err
	}
	row := &marketplace.InstalledPlugin{
		UUID:             uuid.NewString(),
		Name:             item.Name,
		PluginId:         item.Id,
		Version:          item.Version,
		Enabled:          true,
		LicenseModuleKey: item.LicenseModuleKey,
		Creator:          operator,
		Updater:          operator,
		CreateAt:         now,
		UpdateAt:         now,
	}
	return i.store.Insert(ctx, row)
}

func (i *imlMarketplaceModule) Enable(ctx context.Context, id string, enabled bool) error {
	existing, err := i.store.Search(ctx, "", map[string]interface{}{"plugin_id": id}, "")
	if err != nil {
		return err
	}
	if len(existing) == 0 {
		return errors.New("plugin is not installed")
	}
	if enabled {
		item, ok := i.findCatalog(id)
		if !ok {
			return errors.New("plugin not found in catalog")
		}
		if err := license.Check(ctx, item.LicenseModuleKey); err != nil {
			return err
		}
	}
	existing[0].Enabled = enabled
	existing[0].Updater = utils.UserId(ctx)
	existing[0].UpdateAt = time.Now()
	_, uerr := i.store.Update(ctx, existing[0])
	return uerr
}

func (i *imlMarketplaceModule) GetRevenueShare(ctx context.Context) (map[string]float64, error) {
	v, has := i.settingService.Get(ctx, revenueShareKey)
	if !has || v == "" {
		return map[string]float64{}, nil
	}
	shares := map[string]float64{}
	if err := json.Unmarshal([]byte(v), &shares); err != nil {
		return map[string]float64{}, nil
	}
	return shares, nil
}

func (i *imlMarketplaceModule) SetRevenueShare(ctx context.Context, shares map[string]float64) error {
	bytes, err := json.Marshal(shares)
	if err != nil {
		return err
	}
	return i.settingService.Set(ctx, revenueShareKey, string(bytes), utils.UserId(ctx))
}
