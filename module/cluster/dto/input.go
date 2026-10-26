package cluster_dto

//type Create struct {
//	Id             string `json:"id,omitempty"`
//	Name           string `json:"name,omitempty"`
//	Description    string `json:"description,omitempty"`
//	Prefix         string `json:"prefix,omitempty"`
//	Url            string `json:"url,omitempty"`
//	ManagerAddress string `json:"manager_address,omitempty"`
//}
//type Edit struct {
//	Name        *string `json:"name,omitempty"`
//	Description *string `json:"description,omitempty"`
//	Prefix      *string `json:"prefix,omitempty"`
//	Url         *string `json:"url,omitempty"`
//}

//type SaveMonitorConfig struct {
//	Driver string                 `json:"driver"`
//	DefaultConfig map[string]interface{} `json:"config"`
//}

//type MonitorConfig struct {
//	Driver string                 `json:"driver"`
//	DefaultConfig map[string]interface{} `json:"config"`
//}

//type MonitorPartition struct {
//	Id            string `json:"id"`
//	Name          string `json:"name"`
//	EnableMonitor bool   `json:"enable_monitor"`
//}

type ResetCluster struct {
	ManagerAddress string `json:"manager_address"`
}

type CheckCluster struct {
	Address string `json:"address"`
}

// CreateCluster registers a new Apinto data-plane cluster from its admin address
// and assigns its disaster-recovery role/priority.
type CreateCluster struct {
	Name     string `json:"name"`
	Resume   string `json:"resume"`
	Address  string `json:"address"`
	Role     string `json:"role"`     // primary | backup | standalone
	Priority int    `json:"priority"` // lower = preferred during failover
}

// UpdateCluster edits cluster info and/or DR configuration.
type UpdateCluster struct {
	Name     *string `json:"name,omitempty"`
	Resume   *string `json:"resume,omitempty"`
	Role     *string `json:"role,omitempty"`
	Priority *int    `json:"priority,omitempty"`
	Enabled  *bool   `json:"enabled,omitempty"`
}
