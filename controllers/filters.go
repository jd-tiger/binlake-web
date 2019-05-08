package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/models"
)

// FiltersController 过滤器控制器
type FiltersController struct {
	beego.Controller
}

// Index 过滤器首页
func (c *FiltersController) Index() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	fs, _ := models.AdminList(ui)
	// 传入到页面变量名称
	c.Data["adminGroups"] = fs

	// 获取当前用户对应的 所有分组
	us, _ := models.UserGroupList(ui)
	c.Data["UserGroups"] = us

	c.TplName = "filter/index.html"
}

// List 显示过滤列表
func (c *FiltersController) List() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	type FilterPara struct {
		Table string
		Group string
	}
	para := &FilterPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("regular table prefix from search input text ", para)

	var fs []models.Filter
	if para.Group == "" && para.Table != "" {
		fs, _ = models.SearchFilter(ui, para.Table)
	} else {
		fs, _ = models.FilterList(para.Group, para.Table)
	}
	//rc.Ctx.Input.RequestBody
	c.Data["json"] = fs
	c.ServeJSON()
}

// AdminList 返回当前用户下所有管理员组的过滤器
func (c *FiltersController) AdminList() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	fs, _ := models.FilterAdminList(ui.UserName)
	//rc.Ctx.Input.RequestBody
	c.Data["json"] = fs
	c.ServeJSON()
}

// Create 创建 过滤器
func (c *FiltersController) Create() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	// {testgroup white test,test  type:val1 col1,col2 }
	type FilterParas struct {
		Name         string
		Group        string
		Type         string
		Table        string
		Events       []string
		FakeCols     string
		WhiteColumns string
		BlackColumns string
		BusinessKeys string
	}

	fp := &FilterParas{}
	json.Unmarshal(c.Ctx.Input.RequestBody, fp)
	log.Debug("filter parameters ", fp)

	if err := models.CreateFilter(fp.Name, fp.Group, fp.Type, fp.Table,
		fp.FakeCols, fp.WhiteColumns, fp.BlackColumns, fp.BusinessKeys, fp.Events); err != nil {
		result.Code = 118
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}

// Update 更新过滤器
func (c *FiltersController) Update() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	// {testgroup white test,test  type:val1 col1,col2 }
	type FilterParas struct {
		Id           string
		Name         string
		Group        string
		Type         string
		Table        string
		Events       []string
		FakeCols     string
		WhiteColumns string
		BlackColumns string
		BusinessKeys string
	}

	paras := &FilterParas{}
	json.Unmarshal(c.Ctx.Input.RequestBody, paras)
	log.Debug("filter parameters ", paras)

	if err := models.UpdateFilter(paras.Id, paras.Group, paras.Name, paras.Type, paras.Table,
		paras.FakeCols, paras.WhiteColumns, paras.BlackColumns, paras.BusinessKeys, paras.Events); err != nil {
		result.Code = 119
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}

// Delete 删除过滤器
func (c *FiltersController) Delete() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	// {testgroup white test,test  type:val1 col1,col2 }
	type FilterParas struct {
		Id string
	}

	fp := &FilterParas{}
	json.Unmarshal(c.Ctx.Input.RequestBody, fp)

	if err := models.DeleteFilter(fp.Id); err != nil {
		result.Code = 120
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}
