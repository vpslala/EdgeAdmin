package node

import (
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAdmin/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
	"time"
)

type IndexAction struct {
	actionutils.ParentAction
}

func (this *IndexAction) Init() {
	this.Nav("", "node", "node")
	this.SecondMenu("nodes")
}

func (this *IndexAction) RunGet(params struct {
	NodeId int64
}) {
	this.Data["nodeId"] = params.NodeId

	nodeResp, err := this.RPC().NSNodeRPC().FindEnabledNSNode(this.AdminContext(), &pb.FindEnabledNSNodeRequest{NsNodeId: params.NodeId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	node := nodeResp.NsNode
	if node == nil {
		this.WriteString("找不到要操作的节点")
		return
	}

	var clusterMap maps.Map = nil
	if node.NsCluster != nil {
		clusterId := node.NsCluster.Id
		clusterResp, err := this.RPC().NSClusterRPC().FindEnabledNSCluster(this.AdminContext(), &pb.FindEnabledNSClusterRequest{NsClusterId: clusterId})
		if err != nil {
			this.ErrorPage(err)
			return
		}
		cluster := clusterResp.NsCluster
		if cluster != nil {
			clusterMap = maps.Map{
				"id":         cluster.Id,
				"name":       cluster.Name,
				"installDir": cluster.InstallDir,
			}
		}
	}

	// IP地址
	ipAddressesResp, err := this.RPC().NodeIPAddressRPC().FindAllEnabledIPAddressesWithNodeId(this.AdminContext(), &pb.FindAllEnabledIPAddressesWithNodeIdRequest{
		NodeId: params.NodeId,
		Role:   nodeconfigs.NodeRoleDNS,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	ipAddressMaps := []maps.Map{}
	for _, addr := range ipAddressesResp.Addresses {
		ipAddressMaps = append(ipAddressMaps, maps.Map{
			"id":        addr.Id,
			"name":      addr.Name,
			"ip":        addr.Ip,
			"canAccess": addr.CanAccess,
		})
	}

	// 运行状态
	status := &nodeconfigs.NodeStatus{}
	if len(node.StatusJSON) > 0 {
		err = json.Unmarshal(node.StatusJSON, &status)
		if err != nil {
			this.ErrorPage(err)
			return
		}
		status.IsActive = status.IsActive && time.Now().Unix()-status.UpdatedAt <= 60 // N秒之内认为活跃
	}

	// 检查是否有新版本
	if len(status.OS) > 0 {
		checkVersionResp, err := this.RPC().NSNodeRPC().CheckNSNodeLatestVersion(this.AdminContext(), &pb.CheckNSNodeLatestVersionRequest{
			Os:             status.OS,
			Arch:           status.Arch,
			CurrentVersion: status.BuildVersion,
		})
		if err != nil {
			this.ErrorPage(err)
			return
		}
		this.Data["shouldUpgrade"] = checkVersionResp.HasNewVersion
		this.Data["newVersion"] = checkVersionResp.NewVersion
	} else {
		this.Data["shouldUpgrade"] = false
		this.Data["newVersion"] = ""
	}

	this.Data["node"] = maps.Map{
		"id":          node.Id,
		"name":        node.Name,
		"ipAddresses": ipAddressMaps,
		"cluster":     clusterMap,
		"installDir":  node.InstallDir,
		"isInstalled": node.IsInstalled,
		"uniqueId":    node.UniqueId,
		"secret":      node.Secret,
		"isOn":        node.IsOn,

		"status": maps.Map{
			"isActive":             status.IsActive,
			"updatedAt":            status.UpdatedAt,
			"hostname":             status.Hostname,
			"cpuUsage":             status.CPUUsage,
			"cpuUsageText":         fmt.Sprintf("%.2f%%", status.CPUUsage*100),
			"memUsage":             status.MemoryUsage,
			"memUsageText":         fmt.Sprintf("%.2f%%", status.MemoryUsage*100),
			"connectionCount":      status.ConnectionCount,
			"buildVersion":         status.BuildVersion,
			"cpuPhysicalCount":     status.CPUPhysicalCount,
			"cpuLogicalCount":      status.CPULogicalCount,
			"load1m":               fmt.Sprintf("%.2f", status.Load1m),
			"load5m":               fmt.Sprintf("%.2f", status.Load5m),
			"load15m":              fmt.Sprintf("%.2f", status.Load15m),
			"cacheTotalDiskSize":   numberutils.FormatBytes(status.CacheTotalDiskSize),
			"cacheTotalMemorySize": numberutils.FormatBytes(status.CacheTotalMemorySize),
		},
	}

	this.Show()
}
