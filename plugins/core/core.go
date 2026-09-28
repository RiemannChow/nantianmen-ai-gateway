package core

import (
	"net/http"

	system_apikey "github.com/RiemannChow/nantianmen-ai-gateway/controller/system-apikey"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/mcp"

	ai_model "github.com/RiemannChow/nantianmen-ai-gateway/controller/ai-model"

	ai_balance "github.com/RiemannChow/nantianmen-ai-gateway/controller/ai-balance"

	ai_local "github.com/RiemannChow/nantianmen-ai-gateway/controller/ai-local"

	ai_key "github.com/RiemannChow/nantianmen-ai-gateway/controller/ai-key"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/log"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/strategy"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/ai"
	ai_api "github.com/RiemannChow/nantianmen-ai-gateway/controller/ai-api"
	"github.com/RiemannChow/nantianmen-ai-gateway/controller/monitor"
	"github.com/RiemannChow/nantianmen-ai-gateway/controller/router"
	"github.com/RiemannChow/nantianmen-ai-gateway/controller/system"

	plugin_cluster "github.com/RiemannChow/nantianmen-ai-gateway/controller/plugin-cluster"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/cluster"

	"github.com/eolinker/ap-account/controller/role"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/common"

	dynamic_module "github.com/RiemannChow/nantianmen-ai-gateway/controller/dynamic-module"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/release"

	application_authorization "github.com/RiemannChow/nantianmen-ai-gateway/controller/application-authorization"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/subscribe"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/upstream"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/service"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/catalogue"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/my_team"

	"github.com/RiemannChow/nantianmen-ai-gateway/controller/certificate"
	"github.com/RiemannChow/nantianmen-ai-gateway/controller/team_manager"
	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/pm3"
)

func init() {
	pm3.Register("core", new(Driver))
}

type Driver struct {
}

func (d *Driver) Access() map[string][]string {
	return map[string][]string{}
}

func (d *Driver) Create() (pm3.IPlugin, error) {

	p := new(plugin)
	autowire.Autowired(p)
	return p, nil
}

type plugin struct {
	clusterController           cluster.IClusterController                         `autowired:""`
	certificateController       certificate.ICertificateController                 `autowired:""`
	teamManagerController       team_manager.ITeamManagerController                `autowired:""`
	myTeamController            my_team.ITeamController                            `autowired:""`
	appController               service.IAppController                             `autowired:""`
	serviceController           service.IServiceController                         `autowired:""`
	monitorStatisticController  monitor.IMonitorStatisticController                `autowired:""`
	monitorConfigController     monitor.IMonitorConfigController                   `autowired:""`
	catalogueController         catalogue.ICatalogueController                     `autowired:""`
	upstreamController          upstream.IUpstreamController                       `autowired:""`
	routerController            router.IRouterController                           `autowired:""`
	aiAPIController             ai_api.IAPIController                              `autowired:""`
	aiStatisticController       ai.IStatisticController                            `autowired:""`
	aiKeyController             ai_key.IKeyController                              `autowired:""`
	aiBalanceController         ai_balance.IBalanceController                      `autowired:""`
	aiLocalController           ai_local.ILocalModelController                     `autowired:""`
	apiDocController            router.IAPIDocController                           `autowired:""`
	subscribeController         subscribe.ISubscribeController                     `autowired:""`
	strategyController          strategy.IStrategyController                       `autowired:""`
	appAuthorizationController  application_authorization.IAuthorizationController `autowired:""`
	releaseController           release.IReleaseController                         `autowired:""`
	roleController              role.IRoleController                               `autowired:""`
	subscribeApprovalController subscribe.ISubscribeApprovalController             `autowired:""`
	dynamicModuleController     dynamic_module.IDynamicModuleController            `autowired:""`
	pluginClusterController     plugin_cluster.IPluginClusterController            `autowired:""`
	commonController            common.ICommonController                           `autowired:""`
	exportConfigController      system.IExportConfigController                     `autowired:""`
	importConfigController      system.IImportConfigController                     `autowired:""`
	aiProviderController        ai.IProviderController                             `autowired:""`
	aiProviderModelController   ai_model.IProviderModelController                  `autowired:""`
	settingController           system.ISettingController                          `autowired:""`
	initController              system.IInitController                             `autowired:""`
	logController               log.ILogController                                 `autowired:""`
	mcpController               mcp.IMcpController                                 `autowired:""`
	systemAPIKeyController      system_apikey.IAPIKeyController                    `autowired:""`
	apis                        []pm3.Api
}

func (p *plugin) OnComplete() {
	p.apis = append(p.apis, p.certificateApi()...)
	p.apis = append(p.apis, p.clusterApi()...)
	p.apis = append(p.apis, p.TeamManagerApi()...)
	p.apis = append(p.apis, p.MyTeamApi()...)
	p.apis = append(p.apis, p.ServiceApis()...)
	p.apis = append(p.apis, p.catalogueApi()...)
	p.apis = append(p.apis, p.upstreamApis()...)
	p.apis = append(p.apis, p.apiApis()...)
	p.apis = append(p.apis, p.subscribeApis()...)
	p.apis = append(p.apis, p.appAuthorizationApis()...)
	p.apis = append(p.apis, p.releaseApis()...)
	p.apis = append(p.apis, p.DynamicModuleApis()...)
	p.apis = append(p.apis, p.monitorStatisticApis()...)
	p.apis = append(p.apis, p.PartitionPluginApi()...)
	p.apis = append(p.apis, p.commonApis()...)
	p.apis = append(p.apis, p.systemApis()...)
	p.apis = append(p.apis, p.aiAPIs()...)
	p.apis = append(p.apis, p.aiKeyApis()...)
	p.apis = append(p.apis, p.strategyApis()...)
	p.apis = append(p.apis, p.logApis()...)
	p.apis = append(p.apis, p.aiLocalApis()...)
	p.apis = append(p.apis, p.aiBalanceAPIs()...)
	p.apis = append(p.apis, p.mcpAPIs()...)
	p.apis = append(p.apis, p.systemApikeyApis()...)
}

func (p *plugin) Name() string {
	return "core"
}

func (p *plugin) APis() []pm3.Api {
	return p.apis
}

func (p *plugin) Assets() map[string]http.FileSystem {
	return nil
}
