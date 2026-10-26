package marketplace

import (
	"reflect"
	"time"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/store"
)

// InstalledPlugin records a plugin the operator has installed from the catalog.
type InstalledPlugin struct {
	Id               int64     `gorm:"column:id;type:BIGINT(20);AUTO_INCREMENT;NOT NULL;primary_key;comment:id;"`
	UUID             string    `gorm:"type:varchar(36);not null;column:uuid;uniqueIndex:uuid;comment:UUID;"`
	Name             string    `gorm:"type:varchar(100);not null;column:name;comment:name"`
	PluginId         string    `gorm:"type:varchar(100);not null;column:plugin_id;index:plugin_id;comment:目录项ID;"`
	Version          string    `gorm:"type:varchar(32);not null;column:version;comment:安装版本;"`
	Enabled          bool      `gorm:"type:tinyint(1);not null;column:enabled;default:0;comment:是否启用;"`
	LicenseModuleKey string    `gorm:"type:varchar(100);not null;column:license_module_key;comment:关联企业模块Key;"`
	LicensePayload   string    `gorm:"type:text;column:license_payload;comment:交付的License载荷;"`
	Creator          string    `gorm:"size:36;not null;column:creator;comment:creator" aovalue:"creator"`
	Updater          string    `gorm:"size:36;not null;column:updater;comment:updater" aovalue:"updater"`
	CreateAt         time.Time `gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;column:create_at;comment:创建时间"`
	UpdateAt         time.Time `gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;column:update_at;comment:修改时间"`
}

func (i *InstalledPlugin) TableName() string {
	return "marketplace_installed"
}

func (i *InstalledPlugin) IdValue() int64 {
	return i.Id
}

type IInstalledStore interface {
	store.ISearchStore[InstalledPlugin]
}

type imlInstalledStore struct {
	store.SearchStore[InstalledPlugin]
}

func init() {
	autowire.Auto[IInstalledStore](func() reflect.Value {
		return reflect.ValueOf(new(imlInstalledStore))
	})
}
