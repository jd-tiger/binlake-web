package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/models"
)

// UserGroupController 用戶組 控制器
type UserGroupController struct {
	beego.Controller
}

const (
	groupsErrorInit = 100
)

// Index user group首页
func (c *UserGroupController) Index() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	aus, _ := models.AdminList(ui)
	// 传入到页面变量名称
	c.Data["adminGroups"] = aus

	// 获取当前用户对应的 所有分组
	us, _ := models.UserGroupList(ui)
	c.Data["UserGroups"] = us

	c.TplName = "group/index.html"
}

// List 显示用户组关系列表 默认返回空数组
func (c *UserGroupController) List() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	type GroupsPara struct {
		Name string
	}
	para := &GroupsPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("member erp ", para)

	erp := ui.UserName
	//group := para.Name
	if ui.IsSuperAdmin {
		erp = "%"
	}
	//rc.Ctx.Input.RequestBody
	gcs, _ := models.GetLoginUserGroup(erp, para.Name)
	c.Data["json"] = gcs
	c.ServeJSON()
}

// Create 新创建分组 并且将创建人的erp 作为当前分组的管理员
func (c *UserGroupController) Create() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	type GroupPara struct {
		Name string `json:"name"`
		Mark string `json:"mark"`
	}
	para := &GroupPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("group ", para)

	if err := models.CreateGroup(ui.UserName, ui.Email, ui.OrgName, para.Name, para.Mark); err != nil {
		result.Code = groupsErrorInit + 1
		result.Message = fmt.Sprintf("创建新分组失败 分组名%s 创建者 %s %s", para.Name, ui.UserName, err.Error())
	}
}

// Delete 删除分组
func (c *UserGroupController) Delete() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	type GroupPara struct {
		Name string `json:"name"`
	}
	para := &GroupPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("group ", para)

	if !ui.IsSuperAdmin && !models.CheckUGPermission(ui.UserName, para.Name) {
		log.Infof("用户 %s 对组 %s 无权限 ", ui.UserName, para.Name)
		result.Code = groupsErrorInit + 7
		result.Message = fmt.Sprintf("用户 %s 对组 %s 无权限 ", ui.UserName, para.Name)
		return
	}

	if err := models.DeleteGroup(para.Name); err != nil {
		result.Code = groupsErrorInit + 2
		result.Message = fmt.Sprintf("删除分组 %s 失败 %s", para.Name, err.Error())
	}
}

// MemberList 获取当前用户admin 角色组下的所有erp 成员信息
func (c *UserGroupController) MemberList() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	type GroupPara struct {
		Name    string `json:"name"`
		GroupId string `json:"groupId"`
	}
	para := &GroupPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("group paras ", para)

	ugs, _ := models.Members(para.Name)
	c.Data["json"] = ugs
	c.ServeJSON()
}

// AddMember 添加erp 到当前用户组下
func (c *UserGroupController) AddMember() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	/**
		"userName": userName,
	                "groupName": ugMemData.groupName,
	                "email": emailPre + "@jd.com",
	                "orgName": orgName,
	                "role": role
	*/
	type MemberPara struct {
		Erp     string `json:"erp"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		OrgName string `json:"orgName"`
		Role    string `json:"role"`
	}
	para := &MemberPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	log.Debug("add member para ", para)

	if err := models.AddMember(para.Erp, para.Name, para.Email, para.OrgName, para.Role); err != nil {
		result.Code = groupsErrorInit + 3
		result.Message = fmt.Sprintf("删除成员失败 %v 失败信息 %s", para, err.Error())
	}
}

// AdminList 當前用戶对应
func (c *UserGroupController) AdminList() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	ugs, _ := models.AdminList(ui)
	c.Data["json"] = ugs
	c.ServeJSON()
}

// DeleteMember 删除成员
func (c *UserGroupController) DeleteMember() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	type MemberPara struct {
		Id string `json:"id"`
	}
	para := &MemberPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	if err := models.DeleteUser(para.Id); err != nil {
		result.Code = groupsErrorInit + 4
		result.Message = fmt.Sprintf("删除用户 用户id %s %s", para.Id, err.Error())
	}
}

// CheckPermission 获取当前用户在组当中的角色
func (c *UserGroupController) CheckPermission() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	if ui.IsSuperAdmin {
		// 如果是超级管理员 则有组权限
		return
	}

	log.Debug("user information ", ui)

	type GroupPara struct {
		Group string `json:"group"`
	}
	para := &GroupPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	ug, err := models.GetUserDetail(ui.UserName, para.Group)
	if err != nil {
		result.Code = groupsErrorInit + 5
		result.Message = fmt.Sprintf("获取用户详细信息失败 用户名%s 分组名%s %s", ui.UserName, para.Group, err.Error())
		return
	}

	if ug.Role != models.UserRoleAdmin && ug.Role != models.UserRuleCreator {
		result.Code = groupsErrorInit + 6
		result.Message = fmt.Sprintf("当前用户 %s 对分组 %s 无管理员权限", ui.UserName, para.Group)
	}
}

// SwitchRole 角色切换 user -> admin; admin -> user
func (c *UserGroupController) SwitchRole() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)
	log.Debug("login user %v", ui)

	type GroupPara struct {
		Erp  string `json:"erp"`
		Name string `json:"name"`
		Role string `json:"role"`
	}
	para := &GroupPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	if err := models.UpdateUserRole(para.Erp, para.Name, para.Role); err != nil {
		result.Code = groupsErrorInit + 7
		result.Message = fmt.Sprintf("跟新用户角色失败 %v", para)
	}
}

// CheckCreated 检查用户是否为创建者 /users/created/check
func (c *UserGroupController) CheckCreated() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)
	log.Debug("login user %v", ui)

	type GroupPara struct {
		Erp  string `json:"erp"`
		Name string `json:"name"`
	}
	para := &GroupPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	if ui.UserName != para.Erp {
		// 非创建者
		result.Code = groupsErrorInit + 8
		result.Message = fmt.Sprintf("当前用户 %s 非组 %s 创建者 %s", ui.UserName, para.Name, para.Erp)
	}
}
