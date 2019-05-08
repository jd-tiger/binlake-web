package models

import (
	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

// Rule 当前用户下管理员角色下分组对应的规则
type Rule struct {
	RuleId     string `orm:"size(36)"`
	Name       string `orm:"size(64)"`
	CreateTime string `orm:"size(32)"`
	UpdateTime string `orm:"size(32)"`
}

// RuleList 用户规则列表
func RuleList(loginErp string) (rs []Rule, err error) {
	o := orm.NewOrm()

	// mq rule names
	sql := `select mr.id rule_id, mr.name, mr.create_time, mr.update_time from mq_rule mr
			join users us
				on us.group_id = mr.group_id
			where erp = ? and role in ('admin','creator')`

	rs = []Rule{}

	if _, err := o.Raw(sql, loginErp).QueryRows(&rs); err != nil {
		log.Error(err)
		return rs, err
	}

	// todo more kv rule name

	return rs, nil
}
