package cluster

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAdmin/internal/oplogs"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/default/clusters/grants/grantutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/actions"
	"github.com/iwind/TeaGo/maps"
)

type UpdateNodeSSHAction struct {
	actionutils.ParentAction
}

func (this *UpdateNodeSSHAction) Init() {
	this.Nav("", "", "")
}

func (this *UpdateNodeSSHAction) RunGet(params struct {
	NodeId int64
}) {
	nodeResp, err := this.RPC().NodeRPC().FindEnabledNode(this.AdminContext(), &pb.FindEnabledNodeRequest{NodeId: params.NodeId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	if nodeResp.Node == nil {
		this.NotFound("node", params.NodeId)
		return
	}

	node := nodeResp.Node
	this.Data["node"] = maps.Map{
		"id":   node.Id,
		"name": node.Name,
	}

	// SSH
	loginParams := maps.Map{
		"host":    "",
		"port":    "",
		"grantId": 0,
	}
	this.Data["loginId"] = 0
	if node.Login != nil {
		this.Data["loginId"] = node.Login.Id
		if len(node.Login.Params) > 0 {
			err = json.Unmarshal(node.Login.Params, &loginParams)
			if err != nil {
				this.ErrorPage(err)
				return
			}
		}
	}
	this.Data["params"] = loginParams

	// 认证信息
	grantId := loginParams.GetInt64("grantId")
	grantResp, err := this.RPC().NodeGrantRPC().FindEnabledNodeGrant(this.AdminContext(), &pb.FindEnabledNodeGrantRequest{NodeGrantId: grantId})
	if err != nil {
		this.ErrorPage(err)
	}
	var grantMap maps.Map = nil
	if grantResp.NodeGrant != nil {
		grantMap = maps.Map{
			"id":         grantResp.NodeGrant.Id,
			"name":       grantResp.NodeGrant.Name,
			"method":     grantResp.NodeGrant.Method,
			"methodName": grantutils.FindGrantMethodName(grantResp.NodeGrant.Method),
		}
	}
	this.Data["grant"] = grantMap

	this.Show()
}

func (this *UpdateNodeSSHAction) RunPost(params struct {
	NodeId  int64
	LoginId int64
	SshHost string
	SshPort int
	GrantId int64

	Must *actions.Must
}) {
	params.Must.
		Field("sshHost", params.SshHost).
		Require("请输入SSH主机地址").
		Field("sshPort", params.SshPort).
		Gt(0, "SSH主机端口需要大于0").
		Lt(65535, "SSH主机端口需要小于65535")

	if params.GrantId <= 0 {
		this.Fail("需要选择或填写至少一个认证信息")
	}

	login := &pb.NodeLogin{
		Id:   params.LoginId,
		Name: "SSH",
		Type: "ssh",
		Params: maps.Map{
			"grantId": params.GrantId,
			"host":    params.SshHost,
			"port":    params.SshPort,
		}.AsJSON(),
	}

	_, err := this.RPC().NodeRPC().UpdateNodeLogin(this.AdminContext(), &pb.UpdateNodeLoginRequest{
		NodeId:    params.NodeId,
		NodeLogin: login,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}

	// 创建日志
	defer this.CreateLog(oplogs.LevelInfo, "修改节点 %d 配置", params.NodeId)

	this.Success()
}
