package cluster

import (
	"context"
	"errors"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/audit"
	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"

	cluster_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/cluster/dto"

	"github.com/RiemannChow/nantianmen-ai-gateway/gateway/admin"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/go-common/auto"

	"github.com/RiemannChow/nantianmen-ai-gateway/gateway"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/cluster"
	"github.com/eolinker/go-common/utils"
)

var (
	_ IClusterModule = (*imlClusterModule)(nil)
)

type imlClusterModule struct {
	clusterService cluster.IClusterService `autowired:""`
	settingService setting.ISettingService `autowired:""`
}

func (m *imlClusterModule) CheckCluster(ctx context.Context, address ...string) ([]*cluster_dto.Node, error) {
	info, err := admin.Admin(address...).Info(ctx)
	// best-effort: if a cluster id is known for the first address, record its health
	if id, herr := firstClusterIdForAddress(ctx, m, address...); herr == nil && id != "" {
		healthy := 1
		if err != nil {
			healthy = 0
		}
		if _, uerr := m.clusterService.UpdateDR(ctx, id, nil, nil, nil, &healthy); uerr != nil {
			log.Warn("update cluster health:", uerr)
		}
	}
	if err != nil {
		return nil, err
	}
	nodesOut := utils.SliceToSlice(info.Nodes, func(i *admin.Node) *cluster_dto.Node {
		return &cluster_dto.Node{
			Id:       i.Id,
			Name:     i.Name,
			Admins:   i.Admin,
			Peers:    i.Peer,
			Gateways: i.Server,
		}
	})
	nodeStatus(ctx, nodesOut)

	return nodesOut, nil
}

func (m *imlClusterModule) ResetCluster(ctx context.Context, clusterId string, address string) ([]*cluster_dto.Node, error) {

	nodes, err := m.clusterService.UpdateAddress(ctx, clusterId, address)
	if err != nil {
		return nil, err
	}
	err = m.initGateway(ctx, clusterId)
	if err != nil {
		return nil, err
	}
	nodesOut := utils.SliceToSlice(nodes, func(i *cluster.Node) *cluster_dto.Node {
		return &cluster_dto.Node{
			Id:       i.Uuid,
			Name:     i.Name,
			Admins:   i.Admin,
			Peers:    i.Peer,
			Gateways: i.Server,
		}
	})
	v, has := m.settingService.Get(ctx, setting.KeyInvokeAddress)
	if (!has || v == "") && len(nodesOut) > 0 && len(nodesOut[0].Gateways) > 0 {
		m.settingService.Set(ctx, setting.KeyInvokeAddress, nodesOut[0].Gateways[0], utils.UserId(ctx))
	}
	nodeStatus(ctx, nodesOut)
	audit.Log(ctx, &audit.AddInput{
		ResourceType: "cluster",
		ResourceID:   clusterId,
		Action:       audit.ActionUpdate,
		Detail:       "重置集群地址(" + clusterId + ")",
	})
	return nodesOut, nil
}
func (m *imlClusterModule) initGateway(ctx context.Context, clusterId string) error {
	client, err := m.clusterService.GatewayClient(ctx, clusterId)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close(ctx)
		if err != nil {
			log.Warn("close apinto client:", err)
		}
	}()
	return gateway.InitGateway(ctx, clusterId, client)
}
func (m *imlClusterModule) ClusterNodes(ctx context.Context, clusterId string) ([]*cluster_dto.Node, error) {

	nodes, err := m.clusterService.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	nodesOut := utils.SliceToSlice(nodes, func(i *cluster.Node) *cluster_dto.Node {
		return &cluster_dto.Node{
			Id:       i.Uuid,
			Name:     i.Name,
			Admins:   i.Admin,
			Peers:    i.Peer,
			Gateways: i.Server,
		}
	})
	nodeStatus(ctx, nodesOut)

	return nodesOut, nil
}

// firstClusterIdForAddress resolves the stored cluster id whose node admin address
// matches one of the given addresses, for health-state bookkeeping.
func firstClusterIdForAddress(ctx context.Context, m *imlClusterModule, address ...string) (string, error) {
	if len(address) == 0 {
		return "", nil
	}
	clusters, err := m.clusterService.List(ctx)
	if err != nil {
		return "", err
	}
	for _, addr := range address {
		for _, c := range clusters {
			nodes, nerr := m.clusterService.Nodes(ctx, c.Uuid)
			if nerr != nil {
				continue
			}
			for _, n := range nodes {
				for _, a := range n.Admin {
					if a == addr {
						return c.Uuid, nil
					}
				}
			}
		}
	}
	return "", nil
}

func toClusterView(c *cluster.Cluster, nodeCount int) *cluster_dto.ClusterView {
	return &cluster_dto.ClusterView{
		Uuid:       c.Uuid,
		Name:       c.Name,
		Resume:     c.Resume,
		Role:       c.Role,
		Priority:   c.Priority,
		Enabled:    c.Enabled,
		Healthy:    c.Healthy,
		NodeCount:  nodeCount,
		CreateTime: auto.TimeLabel(c.CreateTime),
		UpdateTime: auto.TimeLabel(c.UpdateTime),
	}
}

func (m *imlClusterModule) ListClusters(ctx context.Context) ([]*cluster_dto.ClusterView, error) {
	clusters, err := m.clusterService.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*cluster_dto.ClusterView, 0, len(clusters))
	for _, c := range clusters {
		nodes, nerr := m.clusterService.Nodes(ctx, c.Uuid)
		nc := 0
		if nerr == nil {
			nc = len(nodes)
		}
		out = append(out, toClusterView(c, nc))
	}
	return out, nil
}

func (m *imlClusterModule) CreateCluster(ctx context.Context, input *cluster_dto.CreateCluster) (*cluster_dto.ClusterView, error) {
	if input.Name == "" {
		return nil, errors.New("name is empty")
	}
	if input.Address == "" {
		return nil, errors.New("address is empty")
	}
	role := input.Role
	if role == "" {
		role = "standalone"
	}
	c, err := m.clusterService.Create(ctx, input.Name, input.Resume, input.Address)
	if err != nil {
		return nil, err
	}
	updated, err := m.clusterService.UpdateDR(ctx, c.Uuid, &role, &input.Priority, nil, nil)
	if err != nil {
		return nil, err
	}
	nodes, nerr := m.clusterService.Nodes(ctx, c.Uuid)
	nc := 0
	if nerr == nil {
		nc = len(nodes)
	}
	// initialize gateway so the cluster is immediately usable for publishing
	if err := m.initGateway(ctx, c.Uuid); err != nil {
		log.Warn("init gateway for new cluster:", err)
	}
	audit.Log(ctx, &audit.AddInput{
		ResourceType: "cluster",
		ResourceID:   c.Uuid,
		Action:       audit.ActionCreate,
		Detail:       "创建集群(" + input.Name + ")",
	})
	return toClusterView(updated, nc), nil
}

func (m *imlClusterModule) UpdateCluster(ctx context.Context, id string, input *cluster_dto.UpdateCluster) (*cluster_dto.ClusterView, error) {
	// info update (name/resume) if provided
	if input.Name != nil || input.Resume != nil {
		if _, err := m.clusterService.UpdateInfo(ctx, id, input.Name, input.Resume); err != nil {
			return nil, err
		}
	}
	updated, err := m.clusterService.UpdateDR(ctx, id, input.Role, input.Priority, input.Enabled, nil)
	if err != nil {
		return nil, err
	}
	nodes, nerr := m.clusterService.Nodes(ctx, id)
	nc := 0
	if nerr == nil {
		nc = len(nodes)
	}
	audit.Log(ctx, &audit.AddInput{
		ResourceType: "cluster",
		ResourceID:   id,
		Action:       audit.ActionUpdate,
		Detail:       "更新集群(" + id + ")",
	})
	return toClusterView(updated, nc), nil
}

func (m *imlClusterModule) DeleteCluster(ctx context.Context, id string) error {
	err := m.clusterService.Delete(ctx, id)
	if err == nil {
		audit.Log(ctx, &audit.AddInput{
			ResourceType: "cluster",
			ResourceID:   id,
			Action:       audit.ActionDelete,
			Detail:       "删除集群(" + id + ")",
		})
	}
	return err
}

func (m *imlClusterModule) ProbeCluster(ctx context.Context, id string) ([]*cluster_dto.Node, error) {
	nodes, err := m.clusterService.Nodes(ctx, id)
	if err != nil {
		return nil, err
	}
	var addr string
	for _, n := range nodes {
		if len(n.Admin) > 0 {
			addr = n.Admin[0]
			break
		}
	}
	if addr == "" {
		return nil, errors.New("no admin address for cluster")
	}
	// CheckCluster resolves the cluster id from the address and refreshes health.
	return m.CheckCluster(ctx, addr)
}

func (m *imlClusterModule) DRView(ctx context.Context) (*cluster_dto.DRView, error) {
	clusters, err := m.clusterService.List(ctx)
	if err != nil {
		return nil, err
	}
	view := &cluster_dto.DRView{Clusters: make([]*cluster_dto.ClusterView, 0, len(clusters))}
	for _, c := range clusters {
		nodes, nerr := m.clusterService.Nodes(ctx, c.Uuid)
		nc := 0
		if nerr == nil {
			nc = len(nodes)
		}
		cv := toClusterView(c, nc)
		view.Clusters = append(view.Clusters, cv)
		view.Total++
		if c.Enabled {
			view.Enabled++
		}
		if c.Healthy == 1 {
			view.Healthy++
		}
		switch c.Role {
		case "primary":
			view.Primary++
		case "backup":
			view.Backup++
		}
	}
	return view, nil
}

//
//func (m *imlClusterModule) CreatePartition(ctx context.Context, create *paritiondto.Create) (*paritiondto.Detail, error) {
//	if create.Id == "" {
//		create.Id = uuid.New().String()
//	}
//	if create.Name == "" {
//		return nil, errors.New("name is empty")
//	}
//	clusterId := ""
//	err := m.transaction.Transaction(ctx, func(ctx context.Context) error {
//		clusterInfo, err := m.clusterService.Create(ctx, create.Id, create.Id, create.Description, create.ManagerAddress)
//		if err != nil {
//			return err
//		}
//		if create.Prefix != "" {
//			create.Prefix = fmt.Sprintf("/%s", strings.TrimPrefix(create.Prefix, "/"))
//		}
//		clusterId = clusterInfo.Uuid
//		return m.partitionService.Create(ctx, &partition.CreatePartition{
//			Uuid:    create.Id,
//			Name:    create.Name,
//			Resume:  create.Description,
//			Prefix:  create.Prefix,
//			Url:     create.Url,
//			Cluster: clusterInfo.Uuid,
//		})
//	})
//	if err != nil {
//		return nil, err
//	}
//	err = m.initGateway(ctx, create.Id, clusterId)
//	if err != nil {
//		return nil, err
//	}
//	return m.Get(ctx, create.Id)
//}
//
//func (m *imlClusterModule) SearchByDriver(ctx context.Context, keyword string) ([]*paritiondto.Item, error) {
//	partitions, err := m.partitionService.SearchByDriver(ctx, keyword, nil)
//	if err != nil {
//		return nil, err
//	}
//	countMap, err := m.clusterService.CountByPartition(ctx)
//	if err != nil {
//		return nil, err
//	}
//	items := utils.SliceToSlice(partitions, func(i *partition.Cluster) *paritiondto.Item {
//
//		return &paritiondto.Item{
//			Creator:     auto.UUID(i.Creator),
//			Updater:     auto.UUID(i.Updater),
//			Id:          i.UUID,
//			Name:        i.Name,
//			Description: i.Resume,
//			ClusterNum:  countMap[i.UUID],
//			CreateTime:  auto.TimeLabel(i.CreateTime),
//			UpdateTime:  auto.TimeLabel(i.UpdateTime),
//		}
//	})
//	if len(items) > 0 {
//		counts, err := m.clusterService.CountByPartition(ctx)
//		if err != nil {
//			return nil, err
//		}
//		for _, item := range items {
//			item.ClusterNum = counts[item.Id]
//		}
//	}
//
//	return items, nil
//}
//
//func (m *imlClusterModule) Get(ctx context.Context, id string) (*paritiondto.Detail, error) {
//	pm, err := m.partitionService.Get(ctx, id)
//	if err != nil {
//		return nil, err
//	}
//	//oDetails, err := m.organizationService.SearchByDriver(ctx, "")
//	//if err != nil {
//	//	return nil, err
//	//}
//	//canDelete := true
//	//for _, o := range oDetails {
//	//	for _, p := range o.Clusters {
//	//		if p == id {
//	//			canDelete = false
//	//			break
//	//		}
//	//	}
//	//	if !canDelete {
//	//		break
//	//	}
//	//}
//
//	pd := &paritiondto.Detail{
//		Creator:     auto.UUID(pm.Creator),
//		Updater:     auto.UUID(pm.Updater),
//		Id:          pm.UUID,
//		Name:        pm.Name,
//		Description: pm.Resume,
//		Prefix:      pm.Prefix,
//		CreateTime:  auto.TimeLabel(pm.CreateTime),
//		UpdateTime:  auto.TimeLabel(pm.UpdateTime),
//		//CanDelete:   canDelete,
//	}
//	return pd, nil
//}
//
//func (m *imlClusterModule) Update(ctx context.Context, id string, edit *paritiondto.Edit) (*paritiondto.Detail, error) {
//	err := m.partitionService.Save(ctx, id, &partition.EditPartition{
//		Name:   edit.Name,
//		Resume: edit.Description,
//		Prefix: edit.Prefix,
//		Url:    edit.Url,
//	})
//	if err != nil {
//		return nil, err
//	}
//	return m.Get(ctx, id)
//}
//
//func (m *imlClusterModule) Delete(ctx context.Context, id string) error {
//	return m.transaction.Transaction(ctx, func(ctx context.Context) error {
//		info, err := m.partitionService.Get(ctx, id)
//		if err != nil {
//			if errors.Is(err, gorm.ErrRecordNotFound) {
//				return nil
//			}
//			return err
//		}
//		err = m.clusterService.Delete(ctx, info.Cluster)
//		if err != nil {
//			return err
//		}
//		return m.partitionService.Delete(ctx, id)
//	})
//
//}
//
//func (m *imlClusterModule) Simple(ctx context.Context) ([]*paritiondto.Simple, error) {
//	pm, err := m.partitionService.SearchByDriver(ctx, "", nil)
//	if err != nil {
//		return nil, err
//	}
//	pd := utils.SliceToSlice(pm, func(i *partition.Cluster) *paritiondto.Simple {
//		return &paritiondto.Simple{
//			Id:   i.UUID,
//			Name: i.Name,
//		}
//	})
//	return pd, nil
//}
//
//func (m *imlClusterModule) SimpleByIds(ctx context.Context, ids []string) ([]*paritiondto.Simple, error) {
//	pm, err := m.partitionService.SearchByDriver(ctx, "", map[string]interface{}{
//		"uuid": ids,
//	})
//	if err != nil {
//		return nil, err
//	}
//	pd := utils.SliceToSlice(pm, func(i *partition.Cluster) *paritiondto.Simple {
//		return &paritiondto.Simple{
//			Id:   i.UUID,
//			Name: i.Name,
//		}
//	})
//	return pd, nil
//
//}
//func (m *imlClusterModule) SimpleWithCluster(ctx context.Context) ([]*paritiondto.SimpleWithCluster, error) {
//	pm, err := m.partitionService.SearchByDriver(ctx, "", nil)
//	if err != nil {
//		return nil, err
//	}
//
//	clusterList, err := m.clusterService.List(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	clusterMap := utils.SliceToMapArrayO(clusterList, func(i *cluster.Cluster) (string, *paritiondto.Cluster) {
//		return i.Cluster, &paritiondto.Cluster{
//			Id:          i.Uuid,
//			Name:        i.Name,
//			Description: i.Resume,
//		}
//	})
//	pd := utils.SliceToSlice(pm, func(i *partition.Cluster) *paritiondto.SimpleWithCluster {
//		return &paritiondto.SimpleWithCluster{
//			Id:       i.UUID,
//			Name:     i.Name,
//			Clusters: clusterMap[i.UUID],
//		}
//	})
//	return pd, nil
//}
