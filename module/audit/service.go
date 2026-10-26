package audit

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"

	audit_store "github.com/RiemannChow/nantianmen-ai-gateway/stores/audit"
)

// AuditEvent is re-exported for controllers that build audit records directly.
type AuditEvent = audit_store.AuditEvent

// Well-known audit action constants used by instrumented controllers.
const (
	ActionCreate   = "create"
	ActionUpdate   = "update"
	ActionDelete   = "delete"
	ActionPublish  = "publish"
	ActionEnable   = "enable"
	ActionDisable  = "disable"
	ActionLogin    = "login"
	ActionActivate = "activate"
)

// AddInput describes a single audit record to persist.
type AddInput struct {
	ResourceType string
	ResourceID   string
	Action       string
	Detail       string
	Before       interface{}
	After        interface{}
	IP           string
}

// SearchInput holds the filter + pagination for listing audit events.
type SearchInput struct {
	Keyword      string
	ResourceType string
	Action       string
	Operator     string
	StartTime    int64
	EndTime      int64
	PageNum      int
	PageSize     int
}

// LogsView is the paginated list response.
type LogsView struct {
	List  []*AuditEvent `json:"list"`
	Total int64         `json:"total"`
}

// ExportView carries a generated compliance report for the frontend to download.
type ExportView struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

// IAuditService records and queries audit events.
type IAuditService interface {
	Add(ctx context.Context, in *AddInput) error
	Search(ctx context.Context, in *SearchInput) (*LogsView, error)
	Export(ctx context.Context, in *SearchInput) (*ExportView, error)
}

func init() {
	autowire.Auto[IAuditService](func() reflect.Value {
		return reflect.ValueOf(new(imlAuditService))
	})
}

var svc IAuditService

// Log is a best-effort, non-blocking audit recorder usable from any controller.
// It is safe to call before the service is wired (it becomes a no-op then).
func Log(ctx context.Context, in *AddInput) {
	if svc == nil || in == nil {
		return
	}
	_ = svc.Add(ctx, in)
}

type imlAuditService struct {
	store audit_store.IAuditStore `autowired:""`
}

func (i *imlAuditService) OnInit() {
	svc = i
}

func (i *imlAuditService) Add(ctx context.Context, in *AddInput) error {
	id := utils.UserId(ctx)
	if id == "" {
		id = "system"
	}
	var before, after string
	if in.Before != nil {
		if b, err := json.Marshal(in.Before); err == nil {
			before = string(b)
		}
	}
	if in.After != nil {
		if a, err := json.Marshal(in.After); err == nil {
			after = string(a)
		}
	}
	e := &AuditEvent{
		Operator:     id,
		ResourceType: in.ResourceType,
		ResourceID:   in.ResourceID,
		Action:       in.Action,
		Detail:       in.Detail,
		Before:       before,
		After:        after,
		IP:           in.IP,
		CreateAt:     time.Now(),
	}
	return i.store.Insert(ctx, e)
}

// buildWhere assembles a SQL where-clause + args from the search input.
func buildWhere(in *SearchInput) (string, []interface{}) {
	conds := make([]string, 0, 8)
	args := make([]interface{}, 0, 8)
	if in.ResourceType != "" {
		conds = append(conds, "resource_type = ?")
		args = append(args, in.ResourceType)
	}
	if in.Action != "" {
		conds = append(conds, "action = ?")
		args = append(args, in.Action)
	}
	if in.Operator != "" {
		conds = append(conds, "operator = ?")
		args = append(args, in.Operator)
	}
	if in.StartTime > 0 {
		conds = append(conds, "create_at >= ?")
		args = append(args, time.Unix(in.StartTime, 0))
	}
	if in.EndTime > 0 {
		conds = append(conds, "create_at <= ?")
		args = append(args, time.Unix(in.EndTime, 0))
	}
	if in.Keyword != "" {
		kw := "%" + in.Keyword + "%"
		conds = append(conds, "(operator LIKE ? OR resource_id LIKE ? OR detail LIKE ? OR action LIKE ?)")
		args = append(args, kw, kw, kw, kw)
	}
	if len(conds) == 0 {
		return "", args
	}
	return strings.Join(conds, " AND "), args
}

func (i *imlAuditService) Search(ctx context.Context, in *SearchInput) (*LogsView, error) {
	where, args := buildWhere(in)
	pageNum, pageSize := in.PageNum, in.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	list, total, err := i.store.ListPage(ctx, where, pageNum, pageSize, args, "id DESC")
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*AuditEvent, 0)
	}
	return &LogsView{List: list, Total: total}, nil
}

func (i *imlAuditService) Export(ctx context.Context, in *SearchInput) (*ExportView, error) {
	where, args := buildWhere(in)
	list, err := i.store.ListQuery(ctx, where, args, "id DESC")
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*AuditEvent, 0)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	// UTF-8 BOM so Excel renders Chinese correctly.
	buf.WriteString("\xEF\xBB\xBF")
	_ = w.Write([]string{"时间", "操作人", "来源IP", "资源类型", "资源ID", "动作", "操作描述", "变更前值", "变更后值"})
	for _, e := range list {
		_ = w.Write([]string{
			e.CreateAt.Format("2006-01-02 15:04:05"),
			e.Operator,
			e.IP,
			e.ResourceType,
			e.ResourceID,
			e.Action,
			e.Detail,
			e.Before,
			e.After,
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	fileName := "audit-report-" + time.Now().Format("20060102-150405") + ".csv"
	return &ExportView{FileName: fileName, Content: buf.String()}, nil
}
