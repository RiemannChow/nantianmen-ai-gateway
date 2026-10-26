package payment

import (
	"reflect"
	"time"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/store"
)

// Order is a checkout order placed against a payment provider.
type Order struct {
	Id               int64     `gorm:"column:id;type:BIGINT(20);AUTO_INCREMENT;NOT NULL;primary_key;comment:id;"`
	UUID             string    `gorm:"type:varchar(36);not null;column:uuid;uniqueIndex:uuid;comment:UUID;"`
	Customer         string    `gorm:"type:varchar(100);not null;column:customer;comment:customer"`
	Provider         string    `gorm:"type:varchar(32);not null;column:provider;comment:provider"`
	Amount           float64   `gorm:"type:decimal(12,2);not null;column:amount;comment:amount"`
	Currency         string    `gorm:"type:varchar(8);not null;column:currency;comment:currency"`
	Items            string    `gorm:"type:text;column:items;comment:items JSON"`
	LicenseModuleKey string    `gorm:"type:varchar(100);not null;column:license_module_key;comment:交付的License模块;"`
	Status           string    `gorm:"type:varchar(16);not null;column:status;default:created;comment:created|paid|failed;"`
	LicensePayload   string    `gorm:"type:text;column:license_payload;comment:交付的License载荷;"`
	Creator          string    `gorm:"size:36;not null;column:creator;comment:creator" aovalue:"creator"`
	Updater          string    `gorm:"size:36;not null;column:updater;comment:updater" aovalue:"updater"`
	CreateAt         time.Time `gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;column:create_at;comment:创建时间"`
	UpdateAt         time.Time `gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;column:update_at;comment:修改时间"`
}

func (o *Order) TableName() string {
	return "payment_order"
}

func (o *Order) IdValue() int64 {
	return o.Id
}

// Payment is a single successful (or attempted) charge transaction.
type Payment struct {
	Id          int64     `gorm:"column:id;type:BIGINT(20);AUTO_INCREMENT;NOT NULL;primary_key;comment:id;"`
	UUID        string    `gorm:"type:varchar(36);not null;column:uuid;uniqueIndex:uuid;comment:UUID;"`
	OrderUUID   string    `gorm:"type:varchar(36);not null;column:order_uuid;index:order_uuid;comment:关联订单;"`
	Provider    string    `gorm:"type:varchar(32);not null;column:provider;comment:provider"`
	ProviderTxn string    `gorm:"type:varchar(100);not null;column:provider_txn;comment:第三方交易号;"`
	Amount      float64   `gorm:"type:decimal(12,2);not null;column:amount;comment:amount"`
	Status      string    `gorm:"type:varchar(16);not null;column:status;comment:status"`
	RawPayload  string    `gorm:"type:text;column:raw_payload;comment:回调原始载荷;"`
	Creator     string    `gorm:"size:36;not null;column:creator;comment:creator" aovalue:"creator"`
	CreateAt    time.Time `gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;column:create_at;comment:创建时间"`
}

func (p *Payment) TableName() string {
	return "payment_txn"
}

func (p *Payment) IdValue() int64 {
	return p.Id
}

type IOrderStore interface {
	store.ISearchStore[Order]
}

type imlOrderStore struct {
	store.SearchStore[Order]
}

type IPaymentStore interface {
	store.ISearchStore[Payment]
}

type imlPaymentStore struct {
	store.SearchStore[Payment]
}

func init() {
	autowire.Auto[IOrderStore](func() reflect.Value {
		return reflect.ValueOf(new(imlOrderStore))
	})
	autowire.Auto[IPaymentStore](func() reflect.Value {
		return reflect.ValueOf(new(imlPaymentStore))
	})
}
