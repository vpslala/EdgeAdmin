package gzip

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/dao"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/actions"
	"github.com/iwind/TeaGo/types"
)

type IndexAction struct {
	actionutils.ParentAction
}

func (this *IndexAction) Init() {
	this.Nav("", "setting", "index")
	this.SecondMenu("gzip")
}

func (this *IndexAction) RunGet(params struct {
	LocationId int64
}) {
	webConfig, err := dao.SharedHTTPWebDAO.FindWebConfigWithLocationId(this.AdminContext(), params.LocationId)
	if err != nil {
		this.ErrorPage(err)
		return
	}

	this.Data["webId"] = webConfig.Id

	gzipId := int64(0)
	if webConfig.GzipRef != nil {
		gzipId = webConfig.GzipRef.GzipId
	}
	gzipConfig := &serverconfigs.HTTPGzipConfig{
		Id:   0,
		IsOn: true,
	}
	if gzipId > 0 {
		resp, err := this.RPC().HTTPGzipRPC().FindEnabledHTTPGzipConfig(this.AdminContext(), &pb.FindEnabledGzipConfigRequest{HttpGzipId: gzipId})
		if err != nil {
			this.ErrorPage(err)
			return
		}
		err = json.Unmarshal(resp.HttpGzipJSON, gzipConfig)
		if err != nil {
			this.ErrorPage(err)
			return
		}
	}

	this.Data["gzipConfig"] = gzipConfig

	if webConfig.GzipRef == nil {
		webConfig.GzipRef = &serverconfigs.HTTPGzipRef{
			IsPrior: false,
			IsOn:    false,
			GzipId:  0,
		}
	}
	this.Data["gzipRef"] = webConfig.GzipRef

	this.Show()
}

func (this *IndexAction) RunPost(params struct {
	WebId       int64
	GzipRefJSON []byte
	GzipId      int64
	Level       int
	MinLength   string
	MaxLength   string

	Must *actions.Must
}) {
	defer this.CreateLogInfo("修改Web %d 的Gzip设置", params.WebId)

	if params.Level < 0 || params.Level > 9 {
		this.Fail("请选择正确的压缩级别")
	}

	minLength := &pb.SizeCapacity{Count: -1}
	if len(params.MinLength) > 0 {
		err := json.Unmarshal([]byte(params.MinLength), minLength)
		if err != nil {
			this.ErrorPage(err)
			return
		}
	}

	maxLength := &pb.SizeCapacity{Count: -1}
	if len(params.MaxLength) > 0 {
		err := json.Unmarshal([]byte(params.MaxLength), maxLength)
		if err != nil {
			this.ErrorPage(err)
			return
		}
	}

	gzipRef := &serverconfigs.HTTPGzipRef{}
	err := json.Unmarshal(params.GzipRefJSON, gzipRef)
	if err != nil {
		this.ErrorPage(err)
		return
	}
	gzipRef.GzipId = params.GzipId

	if params.GzipId > 0 {
		_, err := this.RPC().HTTPGzipRPC().UpdateHTTPGzip(this.AdminContext(), &pb.UpdateHTTPGzipRequest{
			HttpGzipId: params.GzipId,
			Level:      types.Int32(params.Level),
			MinLength:  minLength,
			MaxLength:  maxLength,
		})
		if err != nil {
			this.ErrorPage(err)
			return
		}
	} else {
		resp, err := this.RPC().HTTPGzipRPC().CreateHTTPGzip(this.AdminContext(), &pb.CreateHTTPGzipRequest{
			Level:     types.Int32(params.Level),
			MinLength: minLength,
			MaxLength: maxLength,
		})
		if err != nil {
			this.ErrorPage(err)
			return
		}
		gzipId := resp.HttpGzipId
		gzipRef.GzipId = gzipId
	}

	gzipRefJSON, err := json.Marshal(gzipRef)
	if err != nil {
		this.ErrorPage(err)
		return
	}

	_, err = this.RPC().HTTPWebRPC().UpdateHTTPWebGzip(this.AdminContext(), &pb.UpdateHTTPWebGzipRequest{
		WebId:    params.WebId,
		GzipJSON: gzipRefJSON,
	})
	if err != nil {
		this.ErrorPage(err)
	}

	this.Success()
}
