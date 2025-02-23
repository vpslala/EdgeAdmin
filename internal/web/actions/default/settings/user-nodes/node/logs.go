package node

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

type LogsAction struct {
	actionutils.ParentAction
}

func (this *LogsAction) Init() {
	this.Nav("", "node", "log")
	this.SecondMenu("nodes")
}

func (this *LogsAction) RunGet(params struct {
	NodeId int64

	DayFrom string
	DayTo   string
	Keyword string
	Level   string
}) {
	this.Data["nodeId"] = params.NodeId
	this.Data["dayFrom"] = params.DayFrom
	this.Data["dayTo"] = params.DayTo
	this.Data["keyword"] = params.Keyword
	this.Data["level"] = params.Level

	userNodeResp, err := this.RPC().UserNodeRPC().FindEnabledUserNode(this.AdminContext(), &pb.FindEnabledUserNodeRequest{NodeId: params.NodeId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	userNode := userNodeResp.Node
	if userNode == nil {
		this.NotFound("userNode", params.NodeId)
		return
	}

	this.Data["node"] = maps.Map{
		"id":   userNode.Id,
		"name": userNode.Name,
	}

	countResp, err := this.RPC().NodeLogRPC().CountNodeLogs(this.AdminContext(), &pb.CountNodeLogsRequest{
		Role:    nodeconfigs.NodeRoleUser,
		NodeId:  params.NodeId,
		DayFrom: params.DayFrom,
		DayTo:   params.DayTo,
		Keyword: params.Keyword,
		Level:   params.Level,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	count := countResp.Count
	page := this.NewPage(count, 20)

	logsResp, err := this.RPC().NodeLogRPC().ListNodeLogs(this.AdminContext(), &pb.ListNodeLogsRequest{
		NodeId:  params.NodeId,
		Role:    nodeconfigs.NodeRoleUser,
		DayFrom: params.DayFrom,
		DayTo:   params.DayTo,
		Keyword: params.Keyword,
		Level:   params.Level,

		Offset: page.Offset,
		Size:   page.Size,
	})

	logs := []maps.Map{}
	for _, log := range logsResp.NodeLogs {
		logs = append(logs, maps.Map{
			"tag":         log.Tag,
			"description": log.Description,
			"createdTime": timeutil.FormatTime("Y-m-d H:i:s", log.CreatedAt),
			"level":       log.Level,
			"isToday":     timeutil.FormatTime("Y-m-d", log.CreatedAt) == timeutil.Format("Y-m-d"),
		})
	}
	this.Data["logs"] = logs

	this.Data["page"] = page.AsHTML()

	this.Show()
}
