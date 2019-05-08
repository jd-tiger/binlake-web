package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/models"
)

// DbsController 与其余系统交互使用的api 服务都在这里 可能会有相应的业务逻辑处理
type DbsController struct {
	beego.Controller
}

var (
	// dbs new dbs client
	dbs = models.NewDBS()
)

// Process 后端调用授权 由于这个属于异步调用 有流程 需要有回调
func (c *DbsController) Process() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()
	// todo 获取参数
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	para := &[]models.Flow{} // 就是导入的flow 数据类型

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("grant parameters ", para)

	// todo 请求授权接口异步返回结果 这个可能需要有流程审批 所以单独放到一套接口当中来处理
	var instanceIds []string
	var cfg models.Config
	for _, flow := range *para {
		instanceIds = append(instanceIds, flow.InstanceId)
		cfg.Zk = flow.Zk
	}

	// todo 从配置表获取dbs接口地址
	cs, err := models.SelectConfig(&cfg)
	if err != nil || len(cs) == 0 {
		result.Code = 156
		result.Message = fmt.Sprintf("%s\n%s\n",
			models.Message(int(result.Code)), err.Error())
		return
	}

	// 获取对应的数据库实例信息
	grants, err := models.GrantList(instanceIds)
	if err != nil || len(grants) == 0 {
		result.Code = 128
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
		return
	}
}

// CallBack dbs 系统回调接口
func (c *DbsController) CallBack() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	// 更新 订单的状态为 agree
	call := models.CallBack{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &call)
	log.Debug("dbs call backup ", call)

	// 更新数据库订单状态
	if err := models.UpdateStatusByOrderId(fmt.Sprintf("%d", call.DbsOrderId),
		models.MysqlApprovalStatusUnauthorized); err != nil {
		result.Code = 130
		result.Message = fmt.Sprintf("%s\n%s\n%d",
			models.Message(int(result.Code)), err.Error(), call.DbsOrderId)
	}
}

// Grant dbs申请授权
func (c *DbsController) Grant() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	para := &[]models.Flow{} // 就是导入的flow 数据类型

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("grant parameters ", para)

	ips := make(map[string]int)
	var cfg models.Config
	for _, fl := range *para {
		ips[fl.Host] = fl.Port
		cfg.Zk = fl.Zk
	}

	// todo 从配置表获取dbs接口地址
	cs, err := models.SelectConfig(&cfg)
	if err != nil || len(cs) == 0 {
		result.Code = 156
		result.Message = fmt.Sprintf("%s\n%s\n",
			models.Message(int(result.Code)), err.Error())
		return
	}

	code, message := GrantInstances(cs[0].DbsApi, cs[0].DbsToken, ips)
	if code != 1000 && message != "" {
		result.Code = code
		result.Message = message
		return
	}

	for _, fl := range *para {
		if err := models.UpdateStatus(fl.InstanceId, models.MysqlApprovalStatusAgree); err != nil {
			// 更新数据库记录状态为agree 失败
			result.Code = 145
			result.Message = fmt.Sprintf("%s\n%s\n%s",
				models.Message(int(result.Code)), err.Error(), fl.Host)
			return
		}
	}
}

//GrantInstances 调用dbs接口就行授权等操作
func GrantInstances(dbsUrl, token string, ips map[string]int) (code int64, message string) {
	// todo 从库检查
	if check, err := dbs.CheckSlave(dbsUrl, token, ips); err != nil || !check {
		// 从库校验失败
		code = 140
		message = fmt.Sprintf("%s\n%s\n",
			models.Message(int(code)), err.Error())
		return
	}

	// todo 获取授权的ip 段
	authHosts, err := models.IPPrefix()
	if err != nil {
		// 获取授权ip 段失败
		code = 142
		message = fmt.Sprintf("%s\n%s\n",
			models.Message(int(code)), err.Error())
		return
	}

	// todo 请求授权接口
	if ok, err := dbs.Grant(dbsUrl, token, ips, authHosts); err != nil || !ok {
		// dbs 授权失败
		code = 141
		message = fmt.Sprintf("%s\n%s\n",
			models.Message(int(code)), err.Error())
		return
	}

	// todo 开启binlog row格式
	for host, port := range ips {
		if ok, err := dbs.FixBinlogFormat(dbsUrl, token, host, port); err != nil || !ok {
			// dbs 授权失败
			code = 143
			message = fmt.Sprintf("%s\n%s\n",
				models.Message(int(code)), err.Error())
			return
		}
	}

	return 1000, ""
}
