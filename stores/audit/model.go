package audit

import "time"

// AuditEvent is one compliance/audit log record describing a mutating operation
// performed against a resource in the gateway.
type AuditEvent struct {
	Id           int64     `gorm:"column:id;type:BIGINT(20);AUTO_INCREMENT;NOT NULL;primary_key;comment:主键ID;"` // id
	Operator     string    `gorm:"size:64;not null;column:operator;index:idx_audit_operator;comment:操作人"`
	ResourceType string    `gorm:"size:64;not null;column:resource_type;index:idx_audit_resource;comment:资源类型"`
	ResourceID   string    `gorm:"size:128;not null;column:resource_id;comment:资源ID"`
	Action       string    `gorm:"size:64;not null;column:action;index:idx_audit_action;comment:动作"`
	Detail       string    `gorm:"type:text;null;column:detail;comment:操作描述"`
	Before       string    `gorm:"type:text;null;column:before_value;comment:变更前值"`
	After        string    `gorm:"type:text;null;column:after_value;comment:变更后值"`
	IP           string    `gorm:"size:64;null;column:ip;comment:来源IP"`
	CreateAt     time.Time `gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;column:create_at;index:idx_audit_create_at;comment:操作时间"`
}

func (a *AuditEvent) IdValue() int64 {
	return a.Id
}

func (a *AuditEvent) TableName() string {
	return "apinto_audit_event"
}
