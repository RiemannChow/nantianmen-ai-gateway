package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"
	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) clusterApi() []pm3.Api {
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/cluster/nodes", []string{"context", "query:partition"}, []string{"nodes"}, p.clusterController.Nodes, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/cluster/reset", []string{"context", "query:partition", "body"}, []string{"nodes"}, p.clusterController.ResetCluster, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/cluster/check", []string{"context", "body"}, []string{"nodes"}, p.clusterController.Check, access.SystemSettingsApiGatewayManager),
		// DR / multi-cluster management (gated by the multi_cluster license module)
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/cluster/dr", []string{"context"}, []string{"dr"}, p.clusterController.DRView, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/cluster/list", []string{"context"}, []string{"clusters"}, p.clusterController.ListClusters, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/cluster", []string{"context", "body"}, []string{"cluster"}, p.clusterController.CreateCluster, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/cluster/:id", []string{"context", "query:id", "body"}, []string{"cluster"}, p.clusterController.UpdateCluster, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodDelete, "/api/v1/cluster/:id", []string{"context", "query:id"}, []string{"id"}, p.clusterController.DeleteCluster, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/cluster/:id/probe", []string{"context", "query:id"}, []string{"nodes"}, p.clusterController.Probe, access.SystemSettingsApiGatewayManager),
	}
}
