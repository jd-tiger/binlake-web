package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

// Instance instance from MySQL
type Instance struct {
	Id         string `orm:"size(36)"`
	Role       string `orm:"size(16)"`
	GroupId    string `orm:"size(36)"`
	SlaveId    string `orm:"size(10)"`
	Host       string `orm:"size(64)"`
	Port       int    `orm:"size(16)"`
	User       string `orm:"size(32)"`
	Password   string `orm:"size(32)"`
	CreateTime string `orm:"size(32)"`
	UpdateTime string `orm:"size(32)"`
}

// InstanceList 获取当前登录用户对应组下的所有实例
func InstanceList(group, host string) (ins []Instance, err error) {
	ins = []Instance{}
	if group == "" { // 如果没有分组 直接返回空数组
		return ins, nil
	}

	o := orm.NewOrm()

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	sql := `select it.id, us.role, it.group_id, it.slave_id, it.port, it.host, it.user, it.password,
			it.create_time, it.update_time from instances it
			join users us on us.group_id = it.group_id
			where us.group_id = ? and it.host like ? group by it.id `

	if _, err := o.Raw(sql, groupId, "%"+host+"%").QueryRows(&ins); err != nil {
		log.Error(err)
		return ins, err
	}

	return ins, nil
}

// InstanceList 获取当前登录用户对应组下的所有实例
func InstanceAdminList(loginErp, group string) (ins []Instance, err error) {
	ins = []Instance{}
	o := orm.NewOrm()

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	sql := `select it.id, us.role, it.group_id, it.slave_id, it.port, it.host, it.user, it.password,
			it.create_time, it.update_time from instances it
			join users us on us.group_id = it.group_id
			where us.group_id = ? and us.group_id in (select group_id from users where erp = ? and role in ('admin','creator'))
				group by it.id `

	if _, err := o.Raw(sql, groupId, loginErp).QueryRows(&ins); err != nil {
		log.Error(err)
		return ins, err
	}

	return ins, nil
}

// SearchInstance 查询用户所有有权限分组下的实例
func SearchInstance(ui *UserInfo, host string) (ins []Instance, err error) {
	ins = []Instance{}

	o := orm.NewOrm()

	erp := ui.UserName
	if ui.IsSuperAdmin {
		erp = "%"
	}

	sql := `select it.id, us.role, it.group_id, it.slave_id, it.port, it.host, it.user, it.password,
			it.create_time, it.update_time from instances it
			join users us on us.group_id = it.group_id
			where us.erp like ? and it.host like ? group by it.id;`

	if _, err := o.Raw(sql, erp, "%"+host+"%").QueryRows(&ins); err != nil {
		log.Error(err)
		return ins, err
	}

	return ins, nil
}

// CreateInstance 创建MySQL 实例
func CreateInstance(group, zk string, hosts []string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	var slaveId int
	sel := `select slave_id from cluster where zk = '` + zk + `'`
	if err := o.Raw(sel).QueryRow(&slaveId); err != nil {
		log.Error(err)
		return err
	}

	ins := `insert into instances(id, group_id, slave_id, host, port, zk, user, password, create_time, update_time)
			values(?, ?, ?, ?, ?, ?, ?, ?, now(), now())`

	for _, host := range hosts {
		host = strings.TrimSpace(host)
		md5sum.Reset()
		md5sum.Write([]byte(host))
		id := hex.EncodeToString(md5sum.Sum(nil))

		ip := strings.Split(host, ":")[0]
		port, _ := strconv.Atoi(strings.Split(host, ":")[1])

		if _, err := o.Raw(ins, id, groupId, slaveId, ip, port, zk, "", "").Exec(); err != nil {
			log.Error(err)
			return err
		}
	}

	return o.Commit()
}

// DeleteInstance according to instance id
func DeleteInstance(id string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	del := `delete from instances where id = ?`
	if _, err := o.Raw(del, id).Exec(); err != nil {
		log.Error(err)
		return err
	}

	del = `delete from rule where instance_id = ?`
	if _, err := o.Raw(del, id).Exec(); err != nil {
		log.Error(err)
		return err
	}
	return o.Commit()
}

// UpdateStatusById 更新对应实例id数据库审批状态 将流程状态以及请求id 写入到数据库
func UpdateStatusById(orderID, url, status string, instanceIds []string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	upd := `update instances set request_id = ?, url = ?,  status = ? where id in('` +
		strings.Join(instanceIds, "','") + `')`
	if _, err := o.Raw(upd, orderID, url, status).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}

// UpdateStatusByOrderId 根据订单id 更新数据库审批的状态
func UpdateStatusByOrderId(orderId, status string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	upd := `update instances set status = ? where request_id = ?`
	if _, err := o.Raw(upd, status, orderId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}

// UpdateStatus 更新实例状态
func UpdateStatus(instanceId, status string) error {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	upd := `update instances set status = ? where id = ?`
	if _, err := o.Raw(upd, status, instanceId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}

// UpdateInstance 更新实例信息 并且绑定新的关系
func UpdateInstance(groupId, zk string, before []Instance, after []Instance) error {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	var oIDs []string

	md5sum := md5.New()
	// 删除旧的实例
	del := `delete from instances where id = ?`
	for _, it := range before {
		md5sum.Reset()
		md5sum.Write([]byte(fmt.Sprintf("%s:%d", it.Host, it.Port)))
		id := hex.EncodeToString(md5sum.Sum(nil))
		oIDs = append(oIDs, id)
		if _, err := o.Raw(del, id).Exec(); err != nil {
			log.Error(err)
			return err
		}
	}

	md5sum.Reset()

	var slaveId int
	sel := `select slave_id from cluster where zk = '` + zk + `'`
	if err := o.Raw(sel).QueryRow(&slaveId); err != nil {
		log.Error(err)
		return err
	}

	ins := `insert into instances(id, group_id, slave_id, host, port, zk, user, password, create_time, update_time)
			values(?, ?, ?, ?, ?, ?, ?, ?, now(), now())`

	var nIDs []string
	for _, it := range after {
		md5sum.Reset()
		md5sum.Write([]byte(fmt.Sprintf("%s:%d", it.Host, it.Port)))
		id := hex.EncodeToString(md5sum.Sum(nil))

		nIDs = append(nIDs, id)
		if _, err := o.Raw(ins, id, groupId, slaveId, it.Host, it.Port, zk, "", "").Exec(); err != nil {
			log.Error(err)
			return err
		}
	}

	// Rule 对应到rule 表
	type Rule struct {
		GroupId      string `orm:"size(36)"`
		StorageType  string `orm:"size(64)"`
		ConvertClass string `orm:"size(64)"`
		RuleId       string `orm:"size(36)"`
		CreateTime   string `orm:"size(36)"`
		UpdateTime   string `orm:"size(36)"`
	}

	// 绑定关系
	sel = `select re.group_id, re.storage_type,
			re.convert_class, re.rule_id, re.create_time, re.update_time
			from rule where re.instance_id in('` + strings.Join(oIDs, "','") + `'') group by re.rule_id`

	var rls []Rule
	if _, err := o.Raw(sel).QueryRows(rls); err != nil {
		log.Error(err)
		return err
	}

	// 删除旧的规则
	del = `delete from rule where instance_id in ('` + strings.Join(oIDs, "','") + `')`
	if _, err := o.Raw(del).Exec(); err != nil {
		log.Error(err)
		return err
	}

	rIns := `insert into rule(group_id, instance_id, storage_type, convert_class, rule_id, create_time, update_time)
			values(?, ?, ?, ?, ?, now(), now())`
	for _, id := range nIDs {
		for _, re := range rls {
			if _, err := o.Raw(rIns, re.GroupId, id, re.StorageType, re.ConvertClass, re.RuleId,
				re.CreateTime, re.UpdateTime).Exec(); err != nil {
				log.Error(err)
				return err
			}
		}
	}

	return o.Commit()
}

// CheckInstanceExist 校验实例是否已存在
func CheckInstanceExist(host string) (exist bool, err error) {
	o := orm.NewOrm()

	sql := `select id, group_id, slave_id, host, port, zk, status, user,
			password, request_id, url, create_time, update_time
			from instances where host=?`

	res, err := o.Raw(sql, host).Exec()
	if err != nil {
		log.Error(err)
		return false, err
	}

	num, _ := res.RowsAffected()
	if num > 0 {
		exist = true
	} else {
		exist = false
	}

	return exist, nil
}
