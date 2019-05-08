package controllers

import (
	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/models"
)

// RuleController 规则控制器
type RuleController struct {
	beego.Controller
}

// List rule 列表 默认返回空数组
func (c *RuleController) List() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("rule list ", ui.UserName)

	is, _ := models.RuleList(ui.UserName)
	//rc.Ctx.Input.RequestBody
	c.Data["json"] = is
	c.ServeJSON()
}
