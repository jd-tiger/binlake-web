package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/jsmeta"
	"github.com/jd-tiger/binlake-web/models"
)

// AdminController 管理员控制器
type AdminController struct {
	beego.Controller
}

// Monitor2Meta
var Monitor2Meta = monitor2MetaData

// Index 显示管理员操作页面主页
func (c *AdminController) Index() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	fs, _ := models.AdminList(ui)
	// 传入到页面变量名称
	c.Data["adminGroups"] = fs

	// 获取当前用户对应的 所有分组
	us, _ := models.UserGroupList(ui)
	c.Data["UserGroups"] = us

	c.TplName = "admin/index.html"
}

// SetOnline 上线 dump 节点
func (c *AdminController) SetOnline() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	var result []models.RespMsg
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()
	log.Debug("user information ", ui)

	var para []models.Monitor
	json.Unmarshal(c.Ctx.Input.RequestBody, &para)

	for _, mo := range para {
		// reset retry times
		_, err := manager.SetOnline(monitor2MetaData(mo))

		// append response to result
		re := models.RespMsg{Code: 1000, Message: ""}
		if err != nil {
			re.Code = 149
			re.Message = fmt.Sprintf("%s\n%s\ncluseter is %s",
				models.Message(int(re.Code)), err.Error(), mo.Host)
		}
		result = append(result, re)
	}
}

// SetOffline 下线 dump 节点
func (c *AdminController) SetOffline() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	var result []models.RespMsg
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()
	log.Debug("user information ", ui)

	var para []models.Monitor
	json.Unmarshal(c.Ctx.Input.RequestBody, &para)

	for _, mo := range para {
		// reset retry times
		_, err := manager.SetOffline(monitor2MetaData(mo))

		// append response to result
		re := models.RespMsg{Code: 1000, Message: ""}
		if err != nil {
			re.Code = 149
			re.Message = fmt.Sprintf("%s\n%s\ncluseter is %s",
				models.Message(int(re.Code)), err.Error(), mo.Host)
		}
		result = append(result, re)
	}
}

// SetLeader 设置leader 节点
func (c *AdminController) SetLeader() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	var rst []models.RespMsg

	var para []models.Monitor
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &para); err != nil {
		log.Error(err)

		rst = append(rst, models.RespMsg{
			Code:    10,
			Message: fmt.Sprintf("%s %v", "json 反序列化异常 ", err),
		})

		// 返回每行结果集合
		c.Data["json"] = rst
		c.ServeJSON()
		return
	}

	for _, mo := range para {
		if resp, err := manager.SetLeader(monitor2MetaData(mo)); err != nil {
			rst = append(rst, models.RespMsg{
				Code:    resp.Code,
				Message: resp.Message,
			})
		}
	}

	// 返回每行结果集合
	c.Data["json"] = rst
	c.ServeJSON()
}

// SetCandidate 设置候选节点
func (c *AdminController) SetCandidate() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	var rst []models.RespMsg

	var para []models.Monitor
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &para); err != nil {
		rst = append(rst, models.RespMsg{
			Code:    151,
			Message: "设置候选节点失败 " + err.Error(),
		})
		// 返回每行结果集合
		c.Data["json"] = rst
		c.ServeJSON()

		return
	}

	for _, mo := range para {
		resp, err := manager.SetCandidate(monitor2MetaData(mo))

		if err != nil {
			rst = append(rst, models.RespMsg{
				Code:    151,
				Message: "设置候选节点失败 " + err.Error(),
			})
			continue
		}

		rst = append(rst, models.RespMsg{
			Code:    resp.Code,
			Message: resp.Message,
		})
	}

	// 返回每行结果集合
	c.Data["json"] = rst
	c.ServeJSON()
}

// CandidateList 获取候选节点信息
func (c *AdminController) CandidateList() {
	type CandPara struct {
		Zk string
	}

	para := &CandPara{}
	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("candidate in cluster zk", para)

	ds, _ := models.DumpServerList(para.Zk)

	//todo 获取服务的负载 根据负载进行排序

	// 返回每行结果集合
	c.Data["json"] = ds
	c.ServeJSON()
}

// SetBinlog 根据binlog 设置相应的位置
func (c *AdminController) SetBinlog() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	rst := models.RespMsg{Code: 1000, Message: ""}

	para := models.Monitor{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &para)

	if resp, err := manager.SetBinlog(monitor2MetaData(para)); err != nil {
		rst.Code = resp.Code
		rst.Message = resp.Message
	}

	// 返回每行结果集合
	c.Data["json"] = rst
	c.ServeJSON()
}

// ResetCounter 重置计数器
func (c *AdminController) ResetCounter() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	var result []models.RespMsg
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()
	log.Debug("user information ", ui)

	var para []models.Monitor
	json.Unmarshal(c.Ctx.Input.RequestBody, &para)

	for _, mo := range para {
		// reset retry times
		_, err := manager.ResetTryTimes(monitor2MetaData(mo))

		// append response to result
		re := models.RespMsg{Code: 1000, Message: ""}
		if err != nil {
			re.Code = 149
			re.Message = fmt.Sprintf("%s\n%s\ncluseter is %s",
				models.Message(int(re.Code)), err.Error(), mo.Host)
		}
		result = append(result, re)
	}
}

// GetBinlogFiles 获取当前binlog的文件 列表 show binary logs
func (c *AdminController) GetBinlogFiles() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	result := models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	para := models.Monitor{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &para)
	log.Debug("parameters is ", para)

	user := beego.AppConfig.String("dump::user")
	pass := beego.AppConfig.String("dump::password")

	bfs, err := models.GetBinlogFiles(user, pass, para.Host, para.Port)
	if err != nil {
		result.Code = 151
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
		return
	}
	result.Message = bfs
}

// monitor2MetaData 将monitor 结构体 转换成 meta data结构体
func monitor2MetaData(mo models.Monitor) *jsmeta.MetaData {
	db := jsmeta.DbInfo{
		Host: mo.Host,
		Port: int32(mo.Port),
	}

	pos, _ := strconv.Atoi(mo.BinlogPos)
	slave := jsmeta.BinlogInfo{
		BinlogFile:       mo.BinlogFile,
		BinlogPos:        int64(pos),
		ExecutedGtidSets: mo.ExecutedGtidSet,
		Leader:           mo.Leader,
	}

	retryTimes, _ := strconv.Atoi(mo.RetryTimes)

	counter := jsmeta.Counter{
		RetryTimes: int64(retryTimes),
	}

	meta := &jsmeta.MetaData{
		Zk: &jsmeta.ZK{
			Servers: mo.Zk,
			Path:    mo.Path,
		},
		DbInfo:    &db,
		Slave:     &slave,
		Counter:   &counter,
		Candidate: strings.Split(mo.Candidates, ","),
	}
	return meta
}
