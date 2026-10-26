package main

// init module
import (
	_ "github.com/RiemannChow/nantianmen-ai-gateway/frontend"
	_ "github.com/RiemannChow/nantianmen-ai-gateway/gateway/apinto"
	_ "github.com/RiemannChow/nantianmen-ai-gateway/login_driver/feishu"
	_ "github.com/RiemannChow/nantianmen-ai-gateway/login_driver/ldap"
	_ "github.com/RiemannChow/nantianmen-ai-gateway/plugins/core"
	_ "github.com/RiemannChow/nantianmen-ai-gateway/plugins/openapi"
	_ "github.com/RiemannChow/nantianmen-ai-gateway/plugins/permit"
	_ "github.com/RiemannChow/nantianmen-ai-gateway/plugins/publish_flow"
	_ "github.com/RiemannChow/nantianmen-ai-gateway/resources/locale"
	_ "github.com/eolinker/ap-account/plugin"
	_ "github.com/eolinker/go-common/cache/cache_redis"
	_ "github.com/eolinker/go-common/log-init"
	_ "github.com/eolinker/go-common/store/store_mysql"
)
