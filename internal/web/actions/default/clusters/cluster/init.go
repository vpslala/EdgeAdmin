package cluster

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/configloaders"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/default/clusters/cluster/boards"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/default/clusters/cluster/groups"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/default/clusters/cluster/node"
	nodeboards "github.com/TeaOSLab/EdgeAdmin/internal/web/actions/default/clusters/cluster/node/boards"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/default/clusters/cluster/node/thresholds"
	clusters "github.com/TeaOSLab/EdgeAdmin/internal/web/actions/default/clusters/clusterutils"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/helpers"
	"github.com/iwind/TeaGo"
)

func init() {
	TeaGo.BeforeStart(func(server *TeaGo.Server) {
		server.
			Helper(helpers.NewUserMustAuth(configloaders.AdminModuleCodeNode)).
			Helper(clusters.NewClusterHelper()).
			Prefix("/clusters/cluster").
			Get("", new(IndexAction)).
			Get("/nodes", new(NodesAction)).
			GetPost("/installNodes", new(InstallNodesAction)).
			GetPost("/installRemote", new(InstallRemoteAction)).
			Post("/installStatus", new(InstallStatusAction)).
			GetPost("/upgradeRemote", new(UpgradeRemoteAction)).
			Post("/upgradeStatus", new(UpgradeStatusAction)).
			GetPost("/delete", new(DeleteAction)).
			GetPost("/createNode", new(CreateNodeAction)).
			GetPost("/createBatch", new(CreateBatchAction)).
			GetPost("/updateNodeSSH", new(UpdateNodeSSHAction)).
			GetPost("/installManual", new(InstallManualAction)).

			// 节点相关
			Prefix("/clusters/cluster/node").
			Get("", new(node.IndexAction)).
			GetPost("/update", new(node.UpdateAction)).
			GetPost("/install", new(node.InstallAction)).
			Post("/updateInstallStatus", new(node.UpdateInstallStatusAction)).
			Post("/status", new(node.StatusAction)).
			Get("/logs", new(node.LogsAction)).
			Post("/start", new(node.StartAction)).
			Post("/stop", new(node.StopAction)).
			Post("/up", new(node.UpAction)).
			Get("/thresholds", new(thresholds.IndexAction)).
			Get("/detail", new(node.DetailAction)).
			GetPost("/boards", new(nodeboards.IndexAction)).

			// 分组相关
			Prefix("/clusters/cluster/groups").
			Get("", new(groups.IndexAction)).
			GetPost("/createPopup", new(groups.CreatePopupAction)).
			GetPost("/updatePopup", new(groups.UpdatePopupAction)).
			Post("/delete", new(groups.DeleteAction)).
			Post("/sort", new(groups.SortAction)).
			GetPost("/selectPopup", new(groups.SelectPopupAction)).

			// 看板相关
			Prefix("/clusters/cluster/boards").
			Get("", new(boards.IndexAction)).

			EndAll()
	})
}
