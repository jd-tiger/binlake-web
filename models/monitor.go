package models

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

// Monitor 用于监控的结构体显示
type Monitor struct {
	State           bool   `orm:"size(12)"`
	GroupId         string `orm:"size(36)"`
	Host            string `orm:"size(36)"`
	Port            int    `orm:"size(36)"`
	Zk              string `orm:"size(512)"`
	SlaveId         int64  `orm:"size(10)"`
	SlaveUUID       string `orm:"size(36)"`
	Path            string `orm:"size(64)"`
	Status          string `orm:"size(36)"`
	BinlogFile      string `orm:"size(36)"`
	BinlogPos       string `orm:"size(36)"`
	ExecutedGtidSet string `orm:"size(36)"`
	Leader          string `orm:"size(36)"`
	RetryTimes      string `orm:"size(36)"`
	CurrentTime     string `orm:"size(36)"`
	Candidates      string `orm:"size(36)"`
}

const (
	// MysqlApprovalStatusInit 刚刚创建 为初始化状态
	MysqlApprovalStatusInit = "init"

	// MysqlApprovalStatusWait 等待授权结果返回状态
	MysqlApprovalStatusWait = "wait"

	// MysqlApprovalStatusAgree 授权同意状态
	MysqlApprovalStatusAgree = "agree"

	// MysqlApprovalStatusUnauthorized 等待授权状态
	MysqlApprovalStatusUnauthorized = "unauthorized"

	// MysqlApprovalStatusOppose 授权拒绝状态
	MysqlApprovalStatusOppose = "oppose"

	// MysqlApprovalStatusDump 已经成功提交到zk 开始binlog dump状态
	MysqlApprovalStatusDump = "dump"
)

// MonitorList 显示到监控页面信息
func MonitorList(host, group string) (ms []Monitor, err error) {
	// todo 如果分组名称为空 则直接返回空数组
	o := orm.NewOrm()

	ms = []Monitor{}
	if group == "" {
		return ms, nil
	}

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	// 审批完成状态下的节点才能够正常的从服务端dump 数据
	sql := `select it.group_id, cs.zk, it.slave_id, cs.slave_uuid, cs.path, it.host, it.port
			from instances it
			join cluster cs
			on cs.zk = it.zk
			where status = ? and group_id = ? and host like ?`
	if _, err := o.Raw(sql, MysqlApprovalStatusDump, groupId, "%"+host+"%").QueryRows(&ms); err != nil {
		log.Error(err)
		return ms, err
	}

	return ms, nil
}

// MonitorList 显示到监控页面信息
func SearchMonitor(ui *UserInfo, host string) (ms []Monitor, err error) {
	ms = []Monitor{}

	o := orm.NewOrm()

	erp := ui.UserName
	if ui.IsSuperAdmin {
		sql := `select it.group_id, cs.zk, it.slave_id, cs.slave_uuid, cs.path, it.host, it.port from instances it
			join cluster cs on cs.zk = it.zk
			where status = ? and host like ? group by it.group_id, it.host, it.port`
		if _, err := o.Raw(sql, MysqlApprovalStatusDump, "%"+host+"%").QueryRows(&ms); err != nil {
			log.Error(err)
			return ms, err
		}

		return ms, nil
	}

	sql := `select it.group_id, cs.zk, it.slave_id, cs.slave_uuid, cs.path, it.host, it.port from instances it
			join cluster cs on cs.zk = it.zk
			join users us on us.group_id = it.group_id
			where status = ? and us.erp like ? and host like ? group by it.group_id, it.host, it.port`
	if _, err := o.Raw(sql, MysqlApprovalStatusDump, erp, "%"+host+"%").QueryRows(&ms); err != nil {
		log.Error(err)
		return ms, err
	}

	return ms, nil
}

// MonitorInfo 显示监控信息
func MonitorInfo(host string, port int) (ms []Monitor, err error) {
	o := orm.NewOrm()

	ms = []Monitor{}

	// 审批完成状态下的节点才能够正常的从服务端dump 数据
	sql := `select it.group_id, cs.zk, cs.slave_id, cs.slave_uuid, cs.path, it.host, it.port
			from instances it
			join cluster cs
			on cs.zk = it.zk
			where it.host = ? and it.port = ?`
	if _, err := o.Raw(sql, host, port).QueryRows(&ms); err != nil {
		log.Error(err)
		return ms, err
	}

	return ms, nil
}
