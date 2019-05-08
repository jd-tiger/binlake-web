package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/jsmeta"
	"github.com/jd-tiger/binlake-web/models"
)

// FlowController 流程控制器
type FlowController struct {
	beego.Controller
}

const (
	// 后端存储类型

	// StorageTypeMQ 后端存储对象为 MQ /jmq/kafka
	StorageTypeMQ = "MQ_STORAGE"

	// StorageTypeKV 后端存储对象为 KV redis etc.
	StorageTypeKV = "KV_STORAGE"

	// 过滤器类型

	// FilterTypeWhite 白名单 过滤器类型
	FilterTypeWhite = "WHITE"

	// FilterTypeWhite 黑名单 过滤器类型
	FilterTypeBlack = "BLACK"
)

// Index 流程管理页面首页
func (c *FlowController) Index() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)

	fs, _ := models.AdminList(ui)
	// 传入到页面变量名称
	c.Data["adminGroups"] = fs

	// 获取当前用户对应的 所有分组
	us, _ := models.UserGroupList(ui)
	c.Data["UserGroups"] = us

	c.TplName = "flow/index.html"
}

// List 显示MySQL host 列表以及流程状态
func (c *FlowController) List() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	type FlowPara struct {
		Host  string
		Group string
	}
	para := &FlowPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("regular table prefix from search input text ", para)

	var fs []models.Flow
	if para.Group == "" && para.Host != "" {
		fs, _ = models.SearchFlow(ui, para.Host)
	} else {
		fs, _ = models.FlowList(para.Host, para.Group)
	}
	//rc.Ctx.Input.RequestBody
	c.Data["json"] = fs
	c.ServeJSON()
}

// StartDump 开启binlog dump数据
func (c *FlowController) StartDump() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	var fls []models.Flow
	json.Unmarshal(c.Ctx.Input.RequestBody, &fls)

	for _, flow := range fls {
		meta := &jsmeta.MetaData{
			Zk: &jsmeta.ZK{
				Servers: flow.Zk,
				Path:    flow.Path,
			},
		}

		db, err := getDbInfo(flow)
		log.Debug("newly meta info ", db)
		if err != nil {
			result.Code = 133
			result.Message = fmt.Sprintf("%s\n%s",
				models.Message(int(result.Code)), err.Error())
			return
		}
		meta.DbInfo = db

		// 已经开始dump 的数据重新生效 依然需要再查master status 以及 获取candidate
		slave, err := models.GetMasterStatus(db.User, db.Password,
			db.Host, int(db.Port))
		log.Debug("slave info ", slave, " error %v", err)
		if err != nil {
			result.Code = 139
			result.Message = fmt.Sprintf("%s\n%s",
				models.Message(int(result.Code)), err.Error())
			return
		}
		meta.Slave = slave

		candidates, err := models.GetCandidate(flow.Zk)
		if err != nil {
			result.Code = 139
			result.Message = fmt.Sprintf("%s\n%s",
				models.Message(int(result.Code)), err.Error())
			return
		}
		meta.Candidate = candidates

		_, err = manager.CreateZNodes(meta)
		if err != nil {
			result.Code = 132
			result.Message = fmt.Sprintf("%s\n%s",
				models.Message(int(result.Code)), err.Error())
			return
		}

		// 如果之前是dump 不需要更新
		if flow.Status != models.MysqlApprovalStatusDump {
			if err = models.UpdateStatus(flow.InstanceId, models.MysqlApprovalStatusDump); err != nil {
				result.Code = 138
				result.Message = fmt.Sprintf("%s\n%s",
					models.Message(int(result.Code)), err.Error())
				return
			}
		}
	}
}

// getDbInfo get db information from meta database
func getDbInfo(flow models.Flow) (*jsmeta.DbInfo, error) {
	db := &jsmeta.DbInfo{
		Host:      flow.Host,
		User:      beego.AppConfig.String("dump::user"),
		Password:  beego.AppConfig.String("dump::password"),
		Port:      int32(flow.Port),
		SlaveId:   flow.SlaveId,
		SlaveUUID: flow.SlaveUUID,
	}

	db.State = jsmeta.NodeState_ONLINE

	return TakeDbRule(db, flow.InstanceId)
}

// TakeDbRule 根据实例ID 获取db当中rule 规则信息
func TakeDbRule(db *jsmeta.DbInfo, instanceId string) (*jsmeta.DbInfo, error) {
	// get rule according to instance id
	rs, err := models.GetInstanceRule(instanceId)
	log.Debug("instance rules ", rs)
	if err != nil {
		return nil, err
	}

	if len(rs) == 0 {
		return nil, errors.New("实例没有绑定规则 请到<b>规则页面</b>添加规则")
	}

	preTopic := ""
	preRule := &jsmeta.Rule{} // 定义空值
	for _, ir := range rs {
		cs, err := models.GetClusterByInstID(instanceId)
		if err != nil {
			return nil, err
		}

		if len(cs) == 0 {
			return nil, fmt.Errorf("%s", "无法获取集群当中mq信息")
		}

		p := &jsmeta.Pair{
			Key:   "bootstrap.servers",
			Value: cs[0].MqAddr,
		}

		paras, err := models.FixedMQParas()
		if err != nil {
			log.Error(err)
			return nil, err
		}

		// append bootstrap servers
		paras = append(paras, p)

		// append clienid
		paras = append(paras, models.ClientID(ir.Topic, db.Host, int(db.Port)))

		if preTopic == "" {
			// 如果当前topic 为空则按照规则的第一条写入
			rel := newRuleRelation(ir, paras)
			log.Debug("new rule relation ", rel)

			// append 到 db 结构体当中
			db.Rule = append(db.Rule, rel)

			// 设置当前执行信息
			preRule = rel
			preTopic = ir.Topic

		} else if preTopic == ir.Topic {
			// 与前一个topic 值相等 则需要在topic下的规则当中添加规则
			// 由于topic 相当 只需要增加过滤器
			fe := applyFilter2Rule(ir)

			// 回写 mq rule
			mqRule := jsmeta.MQRule{}
			ruleBts, _ := jsmeta.GzipUnCompress(preRule.Any)
			json.Unmarshal(ruleBts, &mqRule)
			if ir.FilterType == FilterTypeWhite {
				mqRule.White = append(mqRule.White, fe)
			} else if ir.FilterType == FilterTypeBlack {
				mqRule.Black = append(mqRule.Black, fe)
			}

			jsonBts, _ := json.Marshal(mqRule)
			bts, _ := jsmeta.GzipCompress(jsonBts)
			preRule.Any = bts

		} else if preTopic != ir.Topic {
			// 需要新建 关系
			rel := newRuleRelation(ir, paras)
			log.Debug("new rule relation ", rel)

			// append 到 db 结构体当中
			db.Rule = append(db.Rule, rel)

			// 设置当前存储信息
			preRule = rel
			preTopic = ir.Topic
		}
	}
	return db, nil
}

// newMQRuleRelation 新建 关系节点
func newRuleRelation(ir models.InstanceRule, paras []*jsmeta.Pair) *jsmeta.Rule {
	re := &jsmeta.Rule{}

	// 格式化 存储类型
	re.Storage = jsmeta.StorageType_value[ir.StorageType]

	// 格式化消息格式
	re.ConvertClass = ir.ConvertClass

	// 生成any 字段
	jsonBts, _ := json.Marshal(newMQRule(ir, paras))
	bts, _ := jsmeta.GzipCompress(jsonBts)
	re.Any = bts

	return re
}

// newMQRule 创建新 mr_rule
func newMQRule(ir models.InstanceRule, paras []*jsmeta.Pair) *jsmeta.MQRule {
	// 开始新的topic
	withTrx, _ := strconv.ParseBool(ir.WithTrx)
	rule := &jsmeta.MQRule{
		Topic:           ir.Topic,
		Para:            paras,
		WithTransaction: withTrx,
		ProducerClass:   ir.ProducerClass,
		Order:           jsmeta.OrderType_value[ir.OrderType],
	}

	fe := applyFilter2Rule(ir)
	if ir.FilterType == FilterTypeWhite {
		rule.White = append(rule.White, fe)
	} else if ir.FilterType == FilterTypeBlack {
		rule.Black = append(rule.Black, fe)
	}
	return rule
}

// applyFilter2Rule 应用过滤器写入到 mq rule 元数据当中
func applyFilter2Rule(ir models.InstanceRule) *jsmeta.Filter {
	filter := &jsmeta.Filter{
		Table:   ir.TableName,
		HashKey: strings.Split(ir.BusinessKeys, ","),
	}

	// 添加 事件类型
	for _, ev := range strings.Split(ir.Events, ",") {
		filter.EventType = append(filter.EventType, jsmeta.EventType_value[ev])
	}

	// 添加 伪列
	for _, fc := range strings.Split(ir.FakeCols, ",") {
		if strings.TrimSpace(fc) == "" {
			// 空串不处理
			continue
		}

		filter.FakeColumn = append(filter.FakeColumn, &jsmeta.Pair{
			Key:   strings.Split(fc, ":")[0],
			Value: strings.Split(fc, ":")[1],
		})
	}

	// 添加保留列列信息
	for _, col := range strings.Split(ir.WhiteColumns, ",") {
		if strings.TrimSpace(col) == "" {
			// 空串不处理
			continue
		}

		filter.White = append(filter.White, &jsmeta.Column{
			Name: col, // 目前不考虑值的处理
		})
	}

	// 添加过滤列列信息
	for _, col := range strings.Split(ir.BlackColumns, ",") {
		if strings.TrimSpace(col) == "" {
			// 空串不处理
			continue
		}

		filter.Black = append(filter.Black, &jsmeta.Column{
			Name: col, // 目前不考虑值的处理
		})
	}
	return filter
}
