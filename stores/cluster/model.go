package cluster

import "time"

type Cluster struct {
	Id      int64  `gorm:"column:id;type:BIGINT(20);AUTO_INCREMENT;NOT NULL;comment:id;primary_key;comment:主键ID;"`
	UUID    string `gorm:"type:varchar(36);not null;column:uuid;uniqueIndex:uuid;comment:UUID;"`
	Name    string `gorm:"type:varchar(100);not null;column:name;comment:name"`
	Cluster string `gorm:"type:varchar(36);not null;column:cluster;comment:cluster id"`
	Resume  string `gorm:"type:varchar(255);not null;column:resume;comment:resume"`
	// DR (disaster recovery) fields. Role: primary/backup/standalone; Priority: lower = preferred
	// during publish failover; Enabled gates the cluster out of the publish fan-out; Healthy is the
	// last-known health state (1 healthy / 0 abnormal) refreshed by /cluster/check.
	Role     string    `gorm:"type:varchar(16);not null;column:role;default:standalone;comment:集群角色 primary/backup/standalone"`
	Priority int       `gorm:"type:int;not null;column:priority;default:100;comment:优先级(越小越优先)"`
	Enabled  bool      `gorm:"type:tinyint(1);not null;column:enabled;default:1;comment:是否启用容灾发布"`
	Healthy  int       `gorm:"type:tinyint(1);not null;column:healthy;default:1;comment:健康状态 1健康 0异常"`
	Creator  string    `gorm:"type:varchar(36);not null;column:creator;comment:creator" aovalue:"creator"`
	Updater  string    `gorm:"type:varchar(36);not null;column:updater;comment:updater" aovalue:"updater"`
	CreateAt time.Time `gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;column:create_at;comment:创建时间"`
	UpdateAt time.Time `gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;column:update_at;comment:修改时间" json:"update_at"`
}

func (c *Cluster) IdValue() int64 {
	return c.Id
}
func (c *Cluster) TableName() string {
	return "cluster"
}

type Node struct {
	Id         int64     `gorm:"column:id;type:BIGINT(20);AUTO_INCREMENT;NOT NULL;comment:id;primary_key;comment:主键ID;"`
	UUID       string    `gorm:"size:36;not null;column:uuid;uniqueIndex:uuid;comment:UUID;"`
	Name       string    `gorm:"size:100;not null;column:name;comment:name"`
	Cluster    string    `gorm:"column:cluster;type:varchar(36);NOT NULL;comment:cluster id;"`
	UpdateTime time.Time `gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;column:update_at;comment:修改时间" json:"update_at"`
}

func (c *Node) IdValue() int64 {
	return c.Id
}
func (c *Node) TableName() string {
	return "cluster_node"
}

type NodeAddr struct {
	Id         int64     `gorm:"column:id;type:BIGINT(20);AUTO_INCREMENT;NOT NULL;comment:id;primary_key;comment:主键ID;"`
	Cluster    string    `gorm:"column:cluster;type:varchar(36);NOT NULL;comment:cluster id;"`
	Node       string    `gorm:"column:node;type:varchar(36);NOT NULL;comment:node id;"`
	Type       string    `gorm:"size:32;not null;column:type;comment:type;uniqueIndex:node_type_addr;"`
	Addr       string    `gorm:"size:255;not null;column:addr;comment:addr;uniqueIndex:node_type_addr;"`
	UpdateTime time.Time `gorm:"type:timestamp;NOT NULL;DEFAULT:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;column:update_at;comment:修改时间" json:"update_at"`
}

func (c *NodeAddr) IdValue() int64 {
	return c.Id
}
func (c *NodeAddr) TableName() string {
	return "cluster_node_addr"
}
