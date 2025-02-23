package cache

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAdmin/internal/oplogs"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/actions"
)

type UpdateAction struct {
	actionutils.ParentAction
}

func (this *UpdateAction) Init() {
	this.Nav("", "", "update")
}

func (this *UpdateAction) RunGet(params struct {
	CachePolicyId int64
}) {
	configResp, err := this.RPC().HTTPCachePolicyRPC().FindEnabledHTTPCachePolicyConfig(this.AdminContext(), &pb.FindEnabledHTTPCachePolicyConfigRequest{HttpCachePolicyId: params.CachePolicyId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	configJSON := configResp.HttpCachePolicyJSON
	if len(configJSON) == 0 {
		this.NotFound("cachePolicy", params.CachePolicyId)
		return
	}

	cachePolicy := &serverconfigs.HTTPCachePolicy{}
	err = json.Unmarshal(configJSON, cachePolicy)
	if err != nil {
		this.ErrorPage(err)
		return
	}
	this.Data["cachePolicy"] = cachePolicy

	// 其他选项
	this.Data["types"] = serverconfigs.AllCachePolicyStorageTypes

	this.Show()
}

func (this *UpdateAction) RunPost(params struct {
	CachePolicyId int64

	Name string
	Type string

	// file
	FileDir                string
	FileMemoryCapacityJSON []byte

	CapacityJSON []byte
	MaxSizeJSON  []byte
	MaxKeys      int64

	Description string
	IsOn        bool

	RefsJSON []byte

	Must *actions.Must
}) {
	// 创建日志
	defer this.CreateLog(oplogs.LevelInfo, "修改缓存策略：%d", params.CachePolicyId)

	params.Must.
		Field("name", params.Name).
		Require("请输入策略名称")

	// 校验选项
	var options interface{}
	switch params.Type {
	case serverconfigs.CachePolicyStorageFile:
		params.Must.
			Field("fileDir", params.FileDir).
			Require("请输入缓存目录")

		memoryCapacity := &shared.SizeCapacity{}
		if len(params.FileMemoryCapacityJSON) > 0 {
			err := json.Unmarshal(params.FileMemoryCapacityJSON, memoryCapacity)
			if err != nil {
				this.ErrorPage(err)
				return
			}
		}

		options = &serverconfigs.HTTPFileCacheStorage{
			Dir: params.FileDir,
			MemoryPolicy: &serverconfigs.HTTPCachePolicy{
				Capacity: memoryCapacity,
			},
		}
	case serverconfigs.CachePolicyStorageMemory:
		options = &serverconfigs.HTTPMemoryCacheStorage{
		}
	default:
		this.Fail("请选择正确的缓存类型")
	}

	optionsJSON, err := json.Marshal(options)
	if err != nil {
		this.ErrorPage(err)
		return
	}

	// 校验缓存条件
	refs := []*serverconfigs.HTTPCacheRef{}
	if len(params.RefsJSON) > 0 {
		err = json.Unmarshal(params.RefsJSON, &refs)
		if err != nil {
			this.Fail("缓存条件解析失败：" + err.Error())
		}
		for _, ref := range refs {
			err = ref.Init()
			if err != nil {
				this.Fail("缓存条件校验失败：" + err.Error())
			}
		}
	}

	_, err = this.RPC().HTTPCachePolicyRPC().UpdateHTTPCachePolicy(this.AdminContext(), &pb.UpdateHTTPCachePolicyRequest{
		HttpCachePolicyId: params.CachePolicyId,
		IsOn:              params.IsOn,
		Name:              params.Name,
		Description:       params.Description,
		CapacityJSON:      params.CapacityJSON,
		MaxKeys:           params.MaxKeys,
		MaxSizeJSON:       params.MaxSizeJSON,
		Type:              params.Type,
		OptionsJSON:       optionsJSON,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}

	// 修改缓存条件
	_, err = this.RPC().HTTPCachePolicyRPC().UpdateHTTPCachePolicyRefs(this.AdminContext(), &pb.UpdateHTTPCachePolicyRefsRequest{
		HttpCachePolicyId: params.CachePolicyId,
		RefsJSON:          params.RefsJSON,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}

	this.Success()
}
