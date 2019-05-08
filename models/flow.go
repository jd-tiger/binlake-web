package models

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

// Flow 用于流程显示列表
type Flow struct {
	InstanceId string `orm:"size(36)"`
	Zk         string `orm:"size(512)"`
	SlaveId    int64  `orm:"size(10)"`
	SlaveUUID  string `orm:"size(36)"`
	Path       string `orm:"size(64)"`
	Status     string `orm:"size(36)"`
	Url        string `orm:"size(128)"`
	Group      string `orm:"size(36)"`
	Host       string `orm:"size(36)"`
	Port       int    `orm:"size(36)"`
	CreateTime string `orm:"size(32)"`
	UpdateTime string `orm:"size(32)"`
}

// InstanceRule 实例对应的规则
type InstanceRule struct {
	Topic         string `orm:"size(36)"`
	WithTrx       string `orm:"size(36)"`
	ConvertClass  string `orm:"size(36)"`
	StorageType   string `orm:"size(36)"`
	OrderType     string `orm:"size(36)"`
	ProducerClass string `orm:"size(128)"`
	Paras         string `orm:"size(3000)"`
	FilterType    string `orm:"size(64)"`
	TableName     string `orm:"size(64)"`
	Events        string `orm:"size(255)"`
	WhiteColumns  string `orm:"size(1024)"`
	BlackColumns  string `orm:"size(1024)"`
	FakeCols      string `orm:"size(254)"`
	BusinessKeys  string `orm:"size(254)"`
}

// FlowList 查找当前用户在管理员组当中的所有实例状态
func FlowList(host, group string) (fs []Flow, err error) {
	fs = []Flow{}
	if group == "" {
		return fs, nil
	}

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	o := orm.NewOrm()

	sql := `select it.id instance_id, it.url, it.zk, it.slave_id, cs.slave_uuid, cs.path, it.status, gs.name 'group', it.host,
			it.port, it.create_time, it.update_time
			from instances it
			join groups gs
			on gs.id = it.group_id
			join cluster cs
			on it.zk = cs.zk
			where group_id = ?
			and gs.name = ? and it.host like ? group by it.id`

	if _, err := o.Raw(sql, groupId, group, "%"+host+"%").QueryRows(&fs); err != nil {
		log.Error(err)
		return fs, err
	}

	return fs, nil
}

// FlowList 查找当前用户在管理员组当中的所有实例状态
func SearchFlow(ui *UserInfo, host string) (fs []Flow, err error) {
	fs = []Flow{}

	o := orm.NewOrm()

	erp := ui.UserName
	if ui.IsSuperAdmin {
		erp = "%"
	}

	sql := `select it.id instance_id, it.url, it.zk, it.slave_id, cs.slave_uuid, cs.path, it.status, gs.name 'group',
			it.host, it.port, it.create_time, it.update_time from instances it
			join groups gs on gs.id = it.group_id
			join cluster cs on it.zk = cs.zk
			join users us on us.group_id = gs.id
			where us.erp like ? and it.host like ? group by it.id`

	if _, err := o.Raw(sql, erp, "%"+host+"%").QueryRows(&fs); err != nil {
		log.Error(err)
		return fs, err
	}

	return fs, nil
}

// GetInstanceRule 获取实例当中包含的 规则列表
func GetInstanceRule(instanceId string) (rs []InstanceRule, err error) {
	o := orm.NewOrm()

	sql := `select re.storage_type, re.convert_class, mr.topic, mr.with_trx,
				mr.order_type, mr.producer_class, mr.paras,
				fe.type filter_type, fe.table_name, fe.events, fe.white_columns, fe.black_columns,
				fe.fake_cols, fe.business_keys
			from mq_rule mr join mq_filter_relation mfr
			on mr.id = mfr.mq_id join filter fe
			on fe.id = mfr.filter_id join rule re
			on re.rule_id = mr.id where re.instance_id = ?`

	rs = []InstanceRule{}
	if _, err := o.Raw(sql, instanceId).QueryRows(&rs); err != nil {
		log.Error(err)
		return rs, err
	}

	return rs, nil
}
