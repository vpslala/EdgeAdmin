package cluster

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

type InstallManualAction struct {
	actionutils.ParentAction
}

func (this *InstallManualAction) Init() {
	this.Nav("", "node", "install")
	this.SecondMenu("nodes")
}

func (this *InstallManualAction) RunGet(params struct {
	ClusterId int64
}) {
	this.Data["leftMenuItems"] = LeftMenuItemsForInstall(this.AdminContext(), params.ClusterId, "manual")

	nodesResp, err := this.RPC().NodeRPC().FindAllNotInstalledNodesWithNodeClusterId(this.AdminContext(), &pb.FindAllNotInstalledNodesWithNodeClusterIdRequest{NodeClusterId: params.ClusterId})
	if err != nil {
		this.ErrorPage(err)
		return
	}

	nodeMaps := []maps.Map{}
	for _, node := range nodesResp.Nodes {
		loginParams := maps.Map{}
		if node.Login != nil && len(node.Login.Params) > 0 {
			err := json.Unmarshal(node.Login.Params, &loginParams)
			if err != nil {
				this.ErrorPage(err)
				return
			}
		}

		installStatus := maps.Map{
			"isRunning":  false,
			"isFinished": false,
		}
		if node.InstallStatus != nil {
			installStatus = maps.Map{
				"isRunning":  node.InstallStatus.IsRunning,
				"isFinished": node.InstallStatus.IsFinished,
				"isOk":       node.InstallStatus.IsOk,
				"error":      node.InstallStatus.Error,
			}
		}

		nodeMaps = append(nodeMaps, maps.Map{
			"id":            node.Id,
			"isOn":          node.IsOn,
			"name":          node.Name,
			"addresses":     node.IpAddresses,
			"login":         node.Login,
			"loginParams":   loginParams,
			"installStatus": installStatus,
		})
	}
	this.Data["nodes"] = nodeMaps

	this.Show()
}
