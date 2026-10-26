package audit

import (
	"context"
	"reflect"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/store"
)

var (
	_ IAuditStore = (*imlAuditStore)(nil)
)

// IAuditStore is the persistence interface for audit events.
type IAuditStore interface {
	Insert(ctx context.Context, t ...*AuditEvent) error
	ListPage(ctx context.Context, where string, pageNum, pageSize int, args []interface{}, order string) ([]*AuditEvent, int64, error)
	ListQuery(ctx context.Context, where string, args []interface{}, order string) ([]*AuditEvent, error)
}

type imlAuditStore struct {
	store.Store[AuditEvent]
}

func init() {
	autowire.Auto[IAuditStore](func() reflect.Value {
		return reflect.ValueOf(new(imlAuditStore))
	})
}
