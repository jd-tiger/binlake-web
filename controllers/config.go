package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/models"
)

const (
	configErrorInit = 900
)

// ConfigController 配置控制器
type ConfigController struct {
	beego.Controller
}

// Index 显示超级管理员 集群管理页面主页
func (c *ConfigController) Index() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	cs, _ := models.SelectConfig(nil)

	c.Data["confs"] = cs
	c.TplName = "config/index.html"
}

// List 展示list
func (c *ConfigController) List() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	if ui != nil {
		log.Info("config list ", ui.UserName)
	}

	log.Info("request host ", c.Ctx.Request.Host)

	para := &models.Config{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("config ", para)

	is, _ := models.SelectConfig(para)
	//rc.Ctx.Input.RequestBody
	c.Data["json"] = is
	c.ServeJSON()
}

// Waves return wave ip address
func (c *ConfigController) Waves() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	if ui != nil {
		log.Info("config list ", ui.UserName)
	}

	log.Info("request host ", c.Ctx.Request.Host)

	para := &models.Config{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("config ", para)

	ips, _ := models.Waves(para)
	//rc.Ctx.Input.RequestBody
	c.Data["json"] = ips
	c.ServeJSON()
}

// Create 创建 config配置 
func (c *ConfigController) Create() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	if ui != nil {
		log.Info("config list ", ui.UserName)
	}

	log.Info("request host ", c.Ctx.Request.Host)

	para := &models.Config{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("config ", para)

	if err := models.SaveConfig(para); err != nil {
		result.Code = configErrorInit + 1
		result.Message = fmt.Sprintf("创建新配置失败 para %v 用户名 %s 异常信息 %v", para, ui.UserName, err)
	}
}

// Update 更新 config配置
func (c *ConfigController) Update() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	if ui != nil {
		log.Info("config list ", ui.UserName)
	}

	log.Info("request host ", c.Ctx.Request)

	para := &models.Config{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("config ", para)

	if err := models.UpdateConfig(para); err != nil {
		result.Code = configErrorInit + 2
		result.Message = fmt.Sprintf("创建新配置失败 para %v 用户名 %s 异常信息 %v", para, ui.UserName, err)
	}
}

// Delete 删除 config配置
func (c *ConfigController) Delete() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	if ui != nil {
		log.Info("config list ", ui.UserName)
	}

	log.Info("request host ", c.Ctx.Request)

	para := &models.Config{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("config ", para)

	if err := models.DeleteConfig(para); err != nil {
		result.Code = configErrorInit + 3
		result.Message = fmt.Sprintf("创建新配置失败 para %v 用户名 %s 异常信息 %v", para, ui.UserName, err)
	}
}
