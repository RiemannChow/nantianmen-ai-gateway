package core

import (
	"net/http"

	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) auditApis() []pm3.Api {
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/audit/logs", []string{"context", "query:keyword", "query:resourceType", "query:action", "query:operator", "query:startTime", "query:endTime", "query:pageNum", "query:pageSize"}, []string{"audit"}, p.auditController.Logs),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/audit/report/export", []string{"context", "query:keyword", "query:resourceType", "query:action", "query:operator", "query:startTime", "query:endTime"}, []string{"audit"}, p.auditController.ExportReport),
	}
}
