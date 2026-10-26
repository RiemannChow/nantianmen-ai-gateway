package audit

import (
	"reflect"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/audit"
	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type IAuditController interface {
	// Logs returns a paginated, filtered list of audit events.
	Logs(ctx *gin.Context, keyword string, resourceType string, action string, operator string, startTime int64, endTime int64, pageNum int, pageSize int) (*audit.LogsView, error)
	// ExportReport returns a CSV compliance report as downloadable content.
	ExportReport(ctx *gin.Context, keyword string, resourceType string, action string, operator string, startTime int64, endTime int64) (*audit.ExportView, error)
}

func init() {
	autowire.Auto[IAuditController](func() reflect.Value {
		return reflect.ValueOf(new(imlAuditController))
	})
}

var _ IAuditController = (*imlAuditController)(nil)

type imlAuditController struct {
	auditService audit.IAuditService `autowired:""`
}

func (i *imlAuditController) Logs(ctx *gin.Context, keyword string, resourceType string, action string, operator string, startTime int64, endTime int64, pageNum int, pageSize int) (*audit.LogsView, error) {
	if err := license.Check(ctx, "audit"); err != nil {
		return nil, err
	}
	return i.auditService.Search(ctx, &audit.SearchInput{
		Keyword:      keyword,
		ResourceType: resourceType,
		Action:       action,
		Operator:     operator,
		StartTime:    startTime,
		EndTime:      endTime,
		PageNum:      pageNum,
		PageSize:     pageSize,
	})
}

func (i *imlAuditController) ExportReport(ctx *gin.Context, keyword string, resourceType string, action string, operator string, startTime int64, endTime int64) (*audit.ExportView, error) {
	if err := license.Check(ctx, "audit"); err != nil {
		return nil, err
	}
	return i.auditService.Export(ctx, &audit.SearchInput{
		Keyword:      keyword,
		ResourceType: resourceType,
		Action:       action,
		Operator:     operator,
		StartTime:    startTime,
		EndTime:      endTime,
	})
}
