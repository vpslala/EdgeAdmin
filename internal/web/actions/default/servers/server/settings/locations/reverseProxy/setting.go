package reverseProxy

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/actions"
	"github.com/iwind/TeaGo/types"
)

type SettingAction struct {
	actionutils.ParentAction
}

func (this *SettingAction) Init() {
	this.FirstMenu("setting")
}

func (this *SettingAction) RunGet(params struct {
	LocationId int64
}) {
	reverseProxyResp, err := this.RPC().HTTPLocationRPC().FindAndInitHTTPLocationReverseProxyConfig(this.AdminContext(), &pb.FindAndInitHTTPLocationReverseProxyConfigRequest{LocationId: params.LocationId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	reverseProxyRef := &serverconfigs.ReverseProxyRef{}
	err = json.Unmarshal(reverseProxyResp.ReverseProxyRefJSON, reverseProxyRef)
	if err != nil {
		this.ErrorPage(err)
		return
	}

	reverseProxy := &serverconfigs.ReverseProxyConfig{}
	err = json.Unmarshal(reverseProxyResp.ReverseProxyJSON, reverseProxy)
	if err != nil {
		this.ErrorPage(err)
		return
	}

	this.Data["reverseProxyRef"] = reverseProxyRef
	this.Data["reverseProxyConfig"] = reverseProxy

	this.Show()
}

func (this *SettingAction) RunPost(params struct {
	LocationId          int64
	ReverseProxyRefJSON []byte
	ReverseProxyJSON    []byte

	Must *actions.Must
}) {
	defer this.CreateLogInfo("修改路由规则 %d 的反向代理设置", params.LocationId)

	// TODO 校验配置

	reverseProxyConfig := &serverconfigs.ReverseProxyConfig{}
	err := json.Unmarshal(params.ReverseProxyJSON, reverseProxyConfig)
	if err != nil {
		this.ErrorPage(err)
		return
	}

	err = reverseProxyConfig.Init()
	if err != nil {
		this.Fail("配置校验失败：" + err.Error())
	}

	// 设置是否启用
	_, err = this.RPC().HTTPLocationRPC().UpdateHTTPLocationReverseProxy(this.AdminContext(), &pb.UpdateHTTPLocationReverseProxyRequest{
		LocationId:       params.LocationId,
		ReverseProxyJSON: params.ReverseProxyRefJSON,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}

	// 设置反向代理相关信息
	_, err = this.RPC().ReverseProxyRPC().UpdateReverseProxy(this.AdminContext(), &pb.UpdateReverseProxyRequest{
		ReverseProxyId:  reverseProxyConfig.Id,
		RequestHostType: types.Int32(reverseProxyConfig.RequestHostType),
		RequestHost:     reverseProxyConfig.RequestHost,
		RequestURI:      reverseProxyConfig.RequestURI,
		StripPrefix:     reverseProxyConfig.StripPrefix,
		AutoFlush:       reverseProxyConfig.AutoFlush,
		AddHeaders:      reverseProxyConfig.AddHeaders,
	})

	this.Success()
}
