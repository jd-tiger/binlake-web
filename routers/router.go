package routers

import (
	"github.com/astaxie/beego"
	"github.com/jd-tiger/binlake-web/controllers"
	"github.com/jd-tiger/binlake-web/filters"
)

func init() {
	beego.InsertFilter("/*", beego.BeforeRouter, filters.LoginFilter)

	beego.Router("/", &controllers.MainController{})
	beego.Router("/index", &controllers.MainController{})
	beego.Router("/logout", &controllers.MainController{}, "*:Logout")
	beego.Router("/login", &controllers.MainController{}, "*:Login")

	// admin
	beego.Router("/admin/index", &controllers.AdminController{}, "get:Index")

	// mq rule
	beego.Router("/mq/index", &controllers.MQController{}, "get:Index")
	beego.Router("/mq/list", &controllers.MQController{}, "post:List")
	beego.Router("/mq/create", &controllers.MQController{}, "post:Create")
	beego.Router("/mq/delete", &controllers.MQController{}, "post:Delete")
	beego.Router("/mq/update", &controllers.MQController{}, "post:Update")

	// 用戶相關接口
	beego.Router("/users/role/switch", &controllers.UserGroupController{}, "Post:SwitchRole")
	beego.Router("/users/created/check", &controllers.UserGroupController{}, "Post:CheckCreated")

	// user group
	beego.Router("/group/index", &controllers.UserGroupController{}, "Get:Index")
	beego.Router("/group/list", &controllers.UserGroupController{}, "Post:List")
	beego.Router("/group/create", &controllers.UserGroupController{}, "Post:Create")
	beego.Router("/group/delete", &controllers.UserGroupController{}, "Post:Delete")
	beego.Router("/group/check", &controllers.UserGroupController{}, "Post:CheckPermission")
	beego.Router("/group/admin/list", &controllers.UserGroupController{}, "Post:AdminList")
	beego.Router("/group/members", &controllers.UserGroupController{}, "Post:MemberList")
	beego.Router("/group/members/add", &controllers.UserGroupController{}, "Post:AddMember")
	beego.Router("/group/members/delete", &controllers.UserGroupController{}, "Post:DeleteMember")

	// filter
	beego.Router("/filter/index", &controllers.FiltersController{}, "get:Index")
	beego.Router("/filter/list", &controllers.FiltersController{}, "post:List")
	beego.Router("/filter/admin/list", &controllers.FiltersController{}, "post:AdminList")
	beego.Router("/filter/create", &controllers.FiltersController{}, "post:Create")
	beego.Router("/filter/update", &controllers.FiltersController{}, "post:Update")
	beego.Router("/filter/delete", &controllers.FiltersController{}, "post:Delete")

	// instances
	beego.Router("/instance/index", &controllers.InstanceController{}, "get:Index")
	beego.Router("/instance/list", &controllers.InstanceController{}, "post:List")
	beego.Router("/instance/admin/list", &controllers.InstanceController{}, "post:AdminList")
	beego.Router("/instance/create", &controllers.InstanceController{}, "post:Create")
	beego.Router("/instance/delete", &controllers.InstanceController{}, "post:Delete")
	beego.Router("/instance/create/complete", &controllers.InstanceController{}, "post:CreateCompleteInstance")

	// instance rule relation
	beego.Router("/relation/index", &controllers.RelationController{}, "get:Index")
	beego.Router("/relation/list", &controllers.RelationController{}, "post:List")
	beego.Router("/relation/create", &controllers.RelationController{}, "post:Create")
	beego.Router("/relation/update", &controllers.RelationController{}, "post:Update")
	beego.Router("/relation/delete", &controllers.RelationController{}, "post:Delete")

	// rule list
	beego.Router("/rule/list", &controllers.RuleController{}, "post:List")

	// execute list
	beego.Router("/flow/index", &controllers.FlowController{}, "get:Index")
	beego.Router("/flow/list", &controllers.FlowController{}, "post:List")

	// flow auth grant
	beego.Router("/flow/auth/process", &controllers.DbsController{}, "post:Process")
	beego.Router("/flow/auth/grant", &controllers.DbsController{}, "post:Grant")
	beego.Router("/flow/auth/callback", &controllers.DbsController{}, "post:CallBack")

	// call zookeeper to start dump
	beego.Router("/flow/start/dump", &controllers.FlowController{}, "post:StartDump")

	// monitor
	beego.Router("/monitor/index", &controllers.MonitorController{}, "get:Index")
	beego.Router("/monitor/list", &controllers.MonitorController{}, "post:List")
	beego.Router("/monitor/compare", &controllers.MonitorController{}, "post:CompareStatus")

	// admin
	beego.Router("/admin/set/online", &controllers.AdminController{}, "post:SetOnline")
	beego.Router("/admin/set/offline", &controllers.AdminController{}, "post:SetOffline")
	beego.Router("/admin/set/leader", &controllers.AdminController{}, "post:SetLeader")
	beego.Router("/admin/set/candidate", &controllers.AdminController{}, "post:SetCandidate")
	beego.Router("/admin/candidate/list", &controllers.AdminController{}, "post:CandidateList")
	beego.Router("/admin/set/binlog", &controllers.AdminController{}, "post:SetBinlog")
	beego.Router("/admin/reset/counter", &controllers.AdminController{}, "post:ResetCounter")
	beego.Router("/admin/binlog/files", &controllers.AdminController{}, "post:GetBinlogFiles")

	// config
	beego.Router("/config/index", &controllers.ConfigController{}, "get:Index")
	beego.Router("/config/list", &controllers.ConfigController{}, "post:List")
	beego.Router("/config/create", &controllers.ConfigController{}, "post:Create")
	beego.Router("/config/waves", &controllers.ConfigController{}, "post:Waves")
	beego.Router("/config/update", &controllers.ConfigController{}, "post:Update")
	beego.Router("/config/delete", &controllers.ConfigController{}, "post:Delete")
}
