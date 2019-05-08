package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/models"
)

// MQController mq 消息队列控制器
type MQController struct {
	beego.Controller
}

// Index mq首页
func (c *MQController) Index() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	fs, _ := models.AdminList(ui)
	// 传入到页面变量名称
	c.Data["adminGroups"] = fs

	// 获取当前用户对应的 所有分组
	us, _ := models.UserGroupList(ui)
	c.Data["UserGroups"] = us

	cs, _ := models.ClusterList()
	c.Data["cluster"] = cs

	c.TplName = "mq/index.html"
}

// List mq 列表 默认返回空数组
func (c *MQController) List() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user info ", ui.UserName)

	type MQTopicPara struct {
		Topic string
		Group string
	}

	para := &MQTopicPara{}
	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("mq topic from  ", para)

	var mrs []models.MQRule
	if para.Group == "" && para.Topic != "" {
		mrs, _ = models.SearchMQRule(ui, para.Topic)
	} else {
		mrs, _ = models.MQRuleList(para.Group, para.Topic)
	}
	//rc.Ctx.Input.RequestBody
	c.Data["json"] = mrs
	c.ServeJSON()
}

// Create 创建mq 路由规则
func (c *MQController) Create() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	type MqRulePara struct {
		Zk        string
		Topic     string
		Name      string
		Order     string
		Group     string
		Producer  string
		WithTrx   bool
		FilterIds []string // 逗号分隔符
	}

	//todo 需要先调用jmq接口创建topic

	para := &MqRulePara{}
	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	cs, err := models.GetClusterByZk(para.Zk)
	if err != nil {
		result.Code = 134
		result.Message = "获取集群失败 " + err.Error()
		return
	}

	if len(cs) == 0 {
		result.Code = 134
		result.Message = "无法获取集群信息 " + para.Zk
		return
	}

	log.Debug("para parameters ", para)
	if err := models.CreateMQ(para.Topic, para.Name, para.Order,
		para.Group, para.Producer,
		para.FilterIds, para.WithTrx); err != nil {
		result.Code = 121
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}

// Delete 删除mq 规则
func (c *MQController) Delete() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	type MQRulePara struct {
		Id string
	}
	para := &MQRulePara{}
	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("delete mq rule para ", para)

	if err := models.DeleteMQ(para.Id); err != nil {
		result.Code = 122
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}

// Update 更新mq rule
func (c *MQController) Update() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	type MqRulePara struct {
		Topic     string
		Mark      string
		Order     string
		Group     string
		Producer  string
		WithTrx   bool
		FilterIds []string // 逗号分隔符
	}

	para := &MqRulePara{}
	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("delete mq rule para ", para)

	if err := models.UpdateMQ(para.Topic, para.Mark,
		para.Order, para.Group, para.Producer,
		para.FilterIds, para.WithTrx); err != nil {
		result.Code = 123
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}
