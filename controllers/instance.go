package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/jsmeta"
	"github.com/jd-tiger/binlake-web/models"
)

// InstanceController 实例控制器
type InstanceController struct {
	beego.Controller
}

// Index index page for instance controller
func (c *InstanceController) Index() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	fs, _ := models.AdminList(ui)

	// 传入到页面变量名称
	c.Data["adminGroups"] = fs

	// 获取当前用户对应的 所有分组
	us, _ := models.UserGroupList(ui)
	c.Data["UserGroups"] = us

	cs, _ := models.ClusterList()
	log.Debug("cluster info ", cs)
	c.Data["cluster"] = cs

	c.TplName = "instance/index.html"
}

// List list mysql instance for user which have access to the groups
func (c *InstanceController) List() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	log.Debug("user information ", ui)
	type InstPara struct {
		Host  string
		Group string
	}
	para := &InstPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("relation host for search input text ", para)

	var is []models.Instance
	if para.Group == "" && para.Host != "" {
		is, _ = models.SearchInstance(ui, para.Host)
	} else {
		is, _ = models.InstanceList(para.Group, para.Host)
	}
	//rc.Ctx.Input.RequestBody
	c.Data["json"] = is
	c.ServeJSON()
}

// AdminList list mysql instance for user which have access to the groups
func (c *InstanceController) AdminList() {
	ui, _ := c.Ctx.Input.Session("UserInfo").(*models.UserInfo)

	type InstPara struct {
		Group string
	}
	para := &InstPara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)
	log.Debug("relation host for search input text ", para)

	is, _ := models.InstanceAdminList(ui.UserName, para.Group)
	//rc.Ctx.Input.RequestBody
	c.Data["json"] = is
	c.ServeJSON()
}

// Create create mysql instance
func (c *InstanceController) Create() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	type InstancePara struct {
		Group string
		Zk    string
		Hosts []string
	}
	para := &InstancePara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	ips, err := hosts2Map(para.Hosts)
	if err != nil {
		result.Code = 163
		result.Message = fmt.Sprintf("host信息转换失败 %s", err.Error())
		return
	}

	// 从配置表获取dbs接口地址
	var cfg models.Config
	cfg.Zk = para.Zk
	cs, err := models.SelectConfig(&cfg)
	if err != nil || len(cs) == 0 {
		result.Code = 156
		result.Message = fmt.Sprintf("%s\n%s\n",
			models.Message(int(result.Code)), err.Error())
		return
	}

	if isSlave, err := dbs.CheckSlave(cs[0].DbsApi, cs[0].DbsToken, ips); err != nil || !isSlave {
		// 如果有异常 或者 不是从库
		result.Code = 161
		result.Message = fmt.Sprintf("%s 从库校验失败 %s", para.Hosts, err.Error())
		return
	}

	log.Debug("instances parameters ", para, ", para host size ", len(para.Hosts))

	if err := models.CreateInstance(para.Group, para.Zk, para.Hosts); err != nil {
		result.Code = 125
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}

// Delete delete from MySQL instance
func (c *InstanceController) Delete() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	type DeletePara struct {
		Id string
	}
	para := &DeletePara{}

	json.Unmarshal(c.Ctx.Input.RequestBody, para)

	log.Debug("delete instance paras %v", para)

	if err := models.DeleteInstance(para.Id); err != nil {
		result.Code = 127
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
	}
}

// hosts2Map
func hosts2Map(hosts []string) (map[string]int, error) {
	var ips = make(map[string]int)
	for _, host := range hosts {
		ipp := strings.Split(host, ":")
		// 必须按照 ip:port 格式
		if len(ipp) != 2 {
			return nil, fmt.Errorf("host 格式错误 %s 必须类似 {172.22.163.107:3358}", host)
		}

		port, err := strconv.Atoi(ipp[1])
		if err != nil {
			return nil, fmt.Errorf("host 端口错误 %s 必须类似 {172.22.163.107:3358}", host)
		}

		ips[ipp[0]] = port
	}

	return ips, nil
}

func (c *InstanceController) CreateCompleteInstance() {
	// 默认返回1000 为正常 消息体为空字符串
	var result = models.RespMsg{Code: 1000, Message: ""}
	defer func() {
		c.Data["json"] = result
		c.ServeJSON()
	}()

	meta := &jsmeta.MetaData{
		DbInfo: &jsmeta.DbInfo{},
		Zk:     &jsmeta.ZK{},
	}
	bts, err := jsmeta.GzipUnCompress(c.Ctx.Input.RequestBody)
	if err != nil {
		result.Code = 153
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
		return
	}
	if err := json.Unmarshal(bts, meta); err != nil {
		result.Code = 153
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
		return
	}

	host := meta.DbInfo.Host
	exist, err := models.CheckInstanceExist(host)
	if err != nil {
		result.Code = 154
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
		return
	}

	//查询binlake中是否已存在实例,若已存在则获取原有的mq规则，不存在则对新实例进行授权
	var preRules []*jsmeta.Rule
	if exist {
		md5sum := md5.New()
		md5sum.Write([]byte(host))
		id := hex.EncodeToString(md5sum.Sum(nil))

		preDbInfo := &jsmeta.DbInfo{}
		preDbInfo, err = TakeDbRule(preDbInfo, id)

		preRules = preDbInfo.Rule
	} else {
		ips := make(map[string]int)
		ips[host] = int(meta.DbInfo.Port)

		// 从配置表获取dbs接口地址
		var cfg models.Config
		cfg.Zk = meta.Zk.Servers
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
	}

	//根据zk查询出对应的mq信息
	mqInfo, err := models.QueryMQInfoByZk(meta.Zk.Servers)
	if err != nil {
		result.Code = 152
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
		return
	}
	if len(mqInfo) < 1 {
		result.Code = 152
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), "根据zk地址查询不到对应mq地址")
		return
	}

	//创建新的topic,并合并新旧mq规则
	rules := meta.DbInfo.Rule
	for _, r := range rules {
		preRules = append(preRules, r)
		mqRule := jsmeta.MQRule{}
		err := json.Unmarshal([]byte(r.Any), &mqRule)
		if err != nil {
			log.Error("反序列化mqRule失败:" + err.Error())
			result.Code = 153
			result.Message = fmt.Sprintf("%s\n%s",
				models.Message(int(result.Code)), err.Error())
			return
		}
	}
	meta.DbInfo.Rule = preRules

	// 已经开始dump 的数据重新生效 依然需要再查master status 以及 获取candidate
	slave, err := models.GetMasterStatus(meta.DbInfo.User, meta.DbInfo.Password,
		meta.DbInfo.Host, int(meta.DbInfo.Port))
	log.Debug("slave info ", slave, " error %v", err)
	if err != nil {
		result.Code = 139
		result.Message = fmt.Sprintf("%s\n%s",
			models.Message(int(result.Code)), err.Error())
		return
	}
	meta.Slave = slave

	candidates, err := models.GetCandidate(meta.Zk.Servers)
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
}
