package controllers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/models"
)

// MonitorController 监控控制器
type MonitorController struct {
	beego.Controller
}

var (
	// 初始化manager
	manager = models.NewManager()
)

// Index 监控管理页面首页
func (c *MonitorController) Index() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	fs, _ := models.AdminList(ui)
	// 传入到页面变量名称
	c.Data["adminGroups"] = fs

	// 获取当前用户对应的 所有分组
	us, _ := models.UserGroupList(ui)
	c.Data["UserGroups"] = us

	c.TplName = "monitor/index.html"
}

// List 监控管理页面首页
func (c *MonitorController) List() {
	// todo 显示监控页面
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	type MonitorPara struct {
		Host  string
		Group string
	}
	para := &MonitorPara{}
	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	// get monitor from database
	var mos []models.Monitor
	if para.Group == "" && para.Host != "" {
		mos, _ = models.SearchMonitor(ui, para.Host)
	} else {
		mos, _ = models.MonitorList(para.Host, para.Group)
	}

	log.Debug("monitor list input paras ", para, " monitor list ", mos)
	if len(mos) == 0 {
		c.Data["json"] = mos
		c.ServeJSON()
		return
	}

	// get dump status from manager-server
	for i, mo := range mos {
		meta, err := manager.GetSlaveStatus(monitor2MetaData(mo))

		if err != nil { // 失败直接返回前段 表示获取数据失败
			log.Error(err)
			break
		}

		mos[i].BinlogFile = meta.Slave.BinlogFile
		mos[i].BinlogPos = fmt.Sprintf("%d", meta.Slave.BinlogPos)
		mos[i].Status = string(meta.DbInfo.State)
		mos[i].ExecutedGtidSet = meta.Slave.ExecutedGtidSets
		mos[i].Leader = meta.Slave.Leader
		mos[i].RetryTimes = fmt.Sprintf("%d", meta.Counter.RetryTimes)
		tm := time.Unix(meta.Slave.BinlogWhen, 0)
		mos[i].CurrentTime = tm.Format("2006-01-02 03:04:05 PM")
		mos[i].Candidates = strings.Join(meta.Candidate, ",")

		mos[i].State = true // 表示状态正常
		if meta.Slave.BinlogFile == "" || meta.Slave.BinlogPos == 0 || meta.Slave.BinlogWhen == 0 || meta.Counter.RetryTimes > 6 {
			mos[i].State = false
		}
	}

	c.Data["json"] = mos
	c.ServeJSON()
}

// CompareStatus 获取主库的位置 用作对照
func (c *MonitorController) CompareStatus() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	// 解析参数
	type MonitorPara struct {
		Host            string
		Port            int
		BinlogFile      string
		BinlogPos       string
		ExecutedGtidSet string
	}

	para := &MonitorPara{}
	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	log.Debug("monitor para ", para)

	user := beego.AppConfig.String("dump::user")
	password := beego.AppConfig.String("dump::password")

	// 获取主库位置 可能也会将dump 的最新位置刷新过来
	master, err := models.GetMasterStatus(user, password, para.Host, para.Port)

	if err != nil {
		result.Code = 135
		result.Message = fmt.Sprintf("%s\n%s\n%s",
			para.Host,
			models.Message(int(result.Code)), err.Error())
		return
	}

	inst := fmt.Sprintf("%s:%d", para.Host, para.Port)

	// 生成compare 数据
	type Compare struct {
		Instance string
		Label    string
		Slave    string
		Master   string
	}

	var cms []Compare
	cms = append(cms, Compare{
		Instance: inst,
		Label:    "binlog file",
		Slave:    para.BinlogFile,
		Master:   master.BinlogFile,
	})
	cms = append(cms, Compare{
		Instance: inst,
		Label:    "binlog position",
		Slave:    para.BinlogPos,
		Master:   fmt.Sprintf("%d", master.BinlogPos),
	})
	cms = append(cms, Compare{
		Instance: inst,
		Label:    "executed gtid sets",
		Slave:    para.ExecutedGtidSet,
		Master:   master.ExecutedGtidSets,
	})

	// 增加compare 字段信息
	result.Message = cms
}
