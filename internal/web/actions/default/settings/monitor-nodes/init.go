package monitornodes

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/configloaders"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/default/settings/monitor-nodes/node"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/default/settings/settingutils"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/helpers"
	"github.com/iwind/TeaGo"
)

func init() {
	TeaGo.BeforeStart(func(server *TeaGo.Server) {
		server.
			Helper(helpers.NewUserMustAuth(configloaders.AdminModuleCodeSetting)).
			Helper(NewHelper()).
			Helper(settingutils.NewAdvancedHelper("monitorNodes")).
			Prefix("/settings/monitorNodes").
			Get("", new(IndexAction)).
			GetPost("/node/createPopup", new(node.CreatePopupAction)).
			Post("/delete", new(DeleteAction)).
			EndAll()
	})
}
