package models

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

// Relation with instance and rule
type Relation struct {
	InstanceId   string `orm:"size(36)"`
	Group        string `orm:"size(36)"`
	Host         string `orm:"size(36)"`
	Port         string `orm:"size(36)"`
	StorageType  string `orm:"size(32)"`
	ConvertClass string `orm:"size(32)"`
	RuleId       string `orm:"size(36)"`
	Name         string `orm:"size(128)"` // mq 规则名称 显示用
	CreateTime   string `orm:"size(32)"`
	UpdateTime   string `orm:"size(32)"`
}

// GetRelation 获取当前用户下的所有关系列表
func GetRelation(group, host string) (rs []Relation, err error) {
	rs = []Relation{}
	if group == "" {
		return rs, nil
	}

	o := orm.NewOrm()
	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	sql := `select re.instance_id, gs.name 'group',
					it.host, it.port, re.storage_type, re.convert_class, re.rule_id,
					mr.name ,
					re.create_time, re.update_time
			from rule re
			join mq_rule mr
			on mr.id = re.rule_id
			join instances it
				on it.id = re.instance_id
			join groups gs on gs.id = re.group_id
			where re.group_id = ? and it.host like ?`

	if _, err := o.Raw(sql, groupId, "%"+host+"%").QueryRows(&rs); err != nil {
		log.Error(err)
		return rs, err
	}

	return rs, nil
}

// SearchRelation 查询用户所有有权限分组下的关系列表
func SearchRelation(ui *UserInfo, host string) (rs []Relation, err error) {
	rs = []Relation{}

	o := orm.NewOrm()

	erp := ui.UserName
	if ui.IsSuperAdmin {
		erp = "%"
	}

	sql := `select re.instance_id, gs.name 'group', it.host, it.port, re.storage_type, re.convert_class,
			re.rule_id, mr.name , re.create_time, re.update_time from rule re
			join mq_rule mr on mr.id = re.rule_id
			join instances it on it.id = re.instance_id
			join groups gs on gs.id = re.group_id
			join users us on us.group_id = gs.id
			where us.erp like ? and it.host like ?`

	if _, err := o.Raw(sql, erp, "%"+host+"%").QueryRows(&rs); err != nil {
		log.Error(err)
		return rs, err
	}

	return rs, nil
}

// CreateRelation 创建 实例与 规则关系
func CreateRelation(hosts []string, group, convertClass, storageType string, ruleIds []string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	// 要求一个instance 只能属于一个组
	check := `select group_id from rule where instance_id = ? group by group_id`
	type Group struct {
		GroupId string
	}

	for _, host := range hosts {
		host = strings.TrimSpace(host) // 去除左右空格
		md5sum.Reset()
		md5sum.Write([]byte(host))
		instanceId := hex.EncodeToString(md5sum.Sum(nil))

		gs := &[]Group{}
		if _, err := o.Raw(check, instanceId).QueryRows(gs); err != nil {
			log.Error(err)
			return err
		}

		if len(*gs) > 1 {
			err = errors.New("存在一个实例对应到多个组 error")
			return err
		}

		if len(*gs) == 1 && (*gs)[0].GroupId != groupId {
			// 如果只有一个分组 并且与当前分组不一致 说明已经存在一个分组包含 当前instance
			err = errors.New("已经存在 分组 " + (*gs)[0].GroupId + " 包含当前 实例")
			return err
		}

		// 清空旧的关联关系
		/*del := `delete from rule where instance_id = ?`
		if _, err := o.Raw(del, instanceId).Exec(); err != nil {
			log.Error(err)
			return err
		}*/

		// 重建关系
		ins := `insert into rule(group_id, instance_id, storage_type, convert_class, rule_id, create_time, update_time)
			values(?, ?, ?, ?, ?, now(), now())`

		for _, rid := range ruleIds {
			if _, err := o.Raw(ins, groupId, instanceId, storageType, convertClass, rid).Exec(); err != nil {
				log.Error(err)
				return err
			}
		}
	}

	return o.Commit()
}

// UpdateRelation 更新实例规则关系
func UpdateRelation(instanceId, group, format, storageType string, ruleIds []string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	// 清空旧的关联关系
	del := `delete from rule where instance_id = ?`
	if _, err := o.Raw(del, instanceId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	// 重建关系
	ins := `insert into rule(group_id, instance_id, storage_type, convert_class, rule_id, create_time, update_time)
			values(?, ?, ?, ?, ?, now(), now())`

	for _, rid := range ruleIds {
		if _, err := o.Raw(ins, groupId, instanceId, storageType, format, rid).Exec(); err != nil {
			log.Error(err)
			return err
		}
	}

	return o.Commit()
}

// DeleteRelation 删除实例与规则的关联关系
func DeleteRelation(instanceId, group, ruleId string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	// 清空旧的关联关系
	ins := `delete from rule where group_id = ? and instance_id = ? and rule_id = ?`

	if _, err := o.Raw(ins, groupId, instanceId, ruleId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}
