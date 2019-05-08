package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/models"
)

// RelationController relation 控制器
type RelationController struct {
	beego.Controller
}

// Index 展示关系首页
func (c *RelationController) Index() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	fs, _ := models.AdminList(ui)
	// 传入到页面变量名称
	c.Data["adminGroups"] = fs

	// 获取当前用户对应的 所有分组
	us, _ := models.UserGroupList(ui)
	c.Data["UserGroups"] = us

	c.TplName = "relation/index.html"
}

// List 显示所有关系
func (c *RelationController) List() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	type InstPara struct {
		Host  string
		Group string
	}
	para := &InstPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("instance host for search input text ", para)

	var is []models.Relation
	if para.Group == "" && para.Host != "" {
		is, _ = models.SearchRelation(ui, para.Host)
	} else {
		is, _ = models.GetRelation(para.Group, para.Host)
	}
	//rc.Ctx.Input.RequestBody
	c.Data["json"] = is
	c.ServeJSON()
}

// Create 创建关系  /relation/create : { "instanceId": instanceId, "group": group, "format": format, "storageType": storageType, "ruleIds": ruleIds }
func (c *RelationController) Create() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	type RelationPara struct {
		Hosts        []string
		Group        string
		ConvertClass string
		StorageType  string
		RuleIds      []string
	}

	para := &RelationPara{}
	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	log.Debug("relation parameters ", para)
	if err := models.CreateRelation(para.Hosts, para.Group, para.ConvertClass, para.StorageType, para.RuleIds); err != nil {
		result.Code = 127
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}

// Update 更新关系
func (c *RelationController) Update() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	type RelationPara struct {
		InstanceId   string
		Group        string
		ConvertClass string
		StorageType  string
		RuleIds      []string
	}

	para := &RelationPara{}
	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	log.Debug("relation parameters ", para)
	if err := models.UpdateRelation(para.InstanceId, para.Group, para.ConvertClass, para.StorageType, para.RuleIds); err != nil {
		result.Code = 128
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}

// Delete 删除规则与实例的关联关系 "instanceId": row.InstanceId, "group": row.Group, "ruleId": row.RuleId
func (c *RelationController) Delete() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	type RelationPara struct {
		InstanceId string
		Group      string
		RuleId     string
	}

	para := &RelationPara{}
	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	log.Debug("relation parameters ", para)
	if err := models.DeleteRelation(para.InstanceId, para.Group, para.RuleId); err != nil {
		result.Code = 128
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}
