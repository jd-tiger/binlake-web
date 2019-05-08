package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

//
// UserGroup 用户信息组
type UserGroup struct {
	Id         string `orm:"size(36)"`
	GroupId    string `orm:"size(36)"`
	Name       string `orm:"size(36)"`
	Mark       string `orm:"size(64)"`
	Erp        string `orm:"size(36)"`
	Role       string `orm:"size(10)"`
	CreateTime string `orm:"size(32)"`
	UpdateTime string `orm:"size(32)"`
}

const (
	// UserRoleAdmin 用户组管理员
	UserRoleAdmin = "admin"

	// UserRoleUser 普通用户
	UserRoleUser = "user"

	// UserRoleCreator 用戶組 創建者
	UserRuleCreator = "creator"

	// AdminRoles 管理员的角色组合
	AdminRoles = "'" + UserRoleAdmin + "','" + UserRuleCreator + "'"
)

var (
	Timeout = time.Second * time.Duration(120)
)

// GetLoginUserGroup 查看当前用户涉及的所有分组
func GetLoginUserGroup(loginUser, group string) (ugs []*UserGroup, err error) {
	o := orm.NewOrm()

	sql := `select us.id, us.group_id, us.erp, us.role, gs.name, gs.mark, us.create_time,
				us.update_time from users us join groups gs on gs.id = us.group_id 
			where us.erp like ? and gs.name like ?`
	log.Debug(sql)

	ugs = []*UserGroup{} // 初始化 保证不为null

	if _, err := o.Raw(sql, loginUser, group+"%").QueryRows(&ugs); err != nil {
		log.Error(err)
		return ugs, err
	}

	return ugs, nil
}

// CreateGroup 创建新分组 创建人为登录用户 对组有管理员权限
func CreateGroup(creator, email, orgName, group, mark string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback() // 最后一定会rollback

	md5sum := md5.New()
	md5sum.Write([]byte(group))

	groupId := hex.EncodeToString(md5sum.Sum(nil))

	insertGroup := `insert into groups(id, name, mark, create_time, update_time)
				values(?, ?, ?, now(), now())`
	log.Debug(insertGroup)

	if _, err := o.Raw(insertGroup, groupId, group, mark).Exec(); err != nil {
		log.Error(err)
		return err
	}

	md5sum.Reset()
	md5sum.Write([]byte(fmt.Sprintf("%s%s", creator, groupId)))

	userId := hex.EncodeToString(md5sum.Sum(nil))
	insertUser := `insert into users(id, erp, email, org_name, group_id, role, create_time, update_time)
			values(?, ?, ?, ?, ?, ?, now(), now())`
	log.Debug(insertUser)

	if _, err := o.Raw(insertUser, userId, creator, email, orgName, groupId, UserRuleCreator).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}

// Members 成员名单 当前登录用户管理员组下的所有成员
func Members(group string) (ugs []*UserGroup, err error) {
	ugs = []*UserGroup{}
	if group == "" {
		return ugs, nil
	}

	o := orm.NewOrm()

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))
	sql := `select us.id, us.group_id, us.erp, us.role, gs.name, gs.mark, us.create_time,
				us.update_time from users us join groups gs on gs.id = us.group_id 
			where us.group_id = ? and gs.name = ?`

	log.Debug(sql)

	if _, err := o.Raw(sql, groupId, group).QueryRows(&ugs); err != nil {
		log.Error(err)
		return ugs, err
	}

	return ugs, nil
}

// AddMember 添加新的用户成员到 分组
func AddMember(user, group, email, org, role string) (err error) {
	o := orm.NewOrm()
	o.Begin()

	defer o.Rollback()

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	md5sum.Reset()
	md5sum.Write([]byte(fmt.Sprintf("%s%s", user, groupId)))
	uID := hex.EncodeToString(md5sum.Sum(nil))

	sql := `insert into users(id, erp, group_id, email, org_name, role, create_time, update_time)values(?, ?, ?, ?, ?, ?, now(), now())`
	if _, err := o.Raw(sql, uID, user, groupId, email, org, role).Exec(); err != nil {
		log.Error(err)
		return err
	}
	// 最后提交 数据
	return o.Commit()
}

// AdminList 当前登录用户对应的管理员分组
func AdminList(ui *UserInfo) (ugs []*UserGroup, err error) {
	o := orm.NewOrm()
	ugs = []*UserGroup{}

	if ui.IsSuperAdmin {
		sql := `select us.group_id, gs.name, gs.mark
				from groups gs join users us
				on us.group_id = gs.id
					group by us.group_id` // 获取所有的分组

		if _, err := o.Raw(sql).QueryRows(&ugs); err != nil {
			log.Error(err)
			return ugs, err
		}

		return ugs, nil
	}

	sql := `select us.id, us.group_id, gs.name, gs.mark,
			us.erp, us.role, us.create_time, us.update_time
			from users us
			join groups gs on us.group_id = gs.id
			where us.role in (` + AdminRoles + `) and us.erp = ? group by gs.id`

	if _, err := o.Raw(sql, ui.UserName).QueryRows(&ugs); err != nil {
		log.Error(err)
		return ugs, err
	}

	return ugs, nil
}

// UserGroupList 当前登录用户对应对应的所有分组
func UserGroupList(ui *UserInfo) (ugs []*UserGroup, err error) {
	o := orm.NewOrm()
	ugs = []*UserGroup{}

	if ui.IsSuperAdmin {
		sql := `select us.group_id, gs.name, gs.mark
				from groups gs join users us
				on us.group_id = gs.id
					group by us.group_id` // 获取所有的分组

		if _, err := o.Raw(sql).QueryRows(&ugs); err != nil {
			log.Error(err)
			return ugs, err
		}

		return ugs, nil
	}

	sql := `select us.id, us.group_id, gs.name, gs.mark,
			us.erp, us.role, us.create_time, us.update_time
			from users us
			join groups gs on us.group_id = gs.id
			where us.erp = ? group by gs.id`

	if _, err := o.Raw(sql, ui.UserName).QueryRows(&ugs); err != nil {
		log.Error(err)
		return ugs, err
	}

	return ugs, nil
}

// GetUserDetail 获取用户详细信息
func GetUserDetail(name, group string) (ug *UserGroup, err error) {
	var ugs []*UserGroup

	o := orm.NewOrm()

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	sql := `select us.id, us.group_id, us.erp, us.role, us.create_time, us.update_time from users us
			where us.erp = ? and us.group_id = ?`

	num, err := o.Raw(sql, name, groupId).QueryRows(&ugs)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if num == 0 {
		return &UserGroup{}, nil
	}

	return ugs[0], nil
}

// DeleteUser 删除用户
func DeleteUser(userId string) (err error) {
	o := orm.NewOrm()
	o.Begin()

	defer o.Rollback()

	sql := `delete from users where id = ?`

	if _, err := o.Raw(sql, userId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}

// DeleteGroup 删除分组
func DeleteGroup(groupName string) (err error) {
	o := orm.NewOrm()
	o.Begin()

	defer o.Rollback()

	md5sum := md5.New()
	md5sum.Write([]byte(groupName))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	// 删除分组
	sql := `delete from groups where id = ?`
	if _, err := o.Raw(sql, groupId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	// 删除分组下的用户
	sql = `delete from users where group_id = ?`
	if _, err := o.Raw(sql, groupId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}

// CheckUGPermission 检查用户对组权限
func CheckUGPermission(userName, group string) bool {
	md5sum := md5.New()
	md5sum.Write([]byte(group))
	gpID := hex.EncodeToString(md5sum.Sum(nil))

	type RowNum struct {
		Cnt int
	}

	var rows []RowNum

	o := orm.NewOrm()
	sql := `select count(1) cnt from users us where us.erp = ? and us.group_id = ? and us.role in (?)`
	if _, err := o.Raw(sql, userName, gpID, AdminRoles).QueryRows(&rows); err != nil {
		log.Error(err)
		return false
	}

	if len(rows) == 1 && rows[0].Cnt == 1 {
		return true
	}

	log.Warnf("用户名 %s 对分组 %s 无权限 %v", userName, group, rows)
	return false
}

// CheckUTPermission 检查用户对表的权限
func CheckUTPermission(userName, tableName string) bool {
	sql := `select count(1) from wt_overview wto join users us on us.group_id = wto.group_id
			join groups gs on gs.id = wto.group_id
			where wto.table_name like '?%' and us.erp = ?
				and user.role in (?)`
	type RowNum struct {
		Cnt int
	}

	var rows []*RowNum

	o := orm.NewOrm()
	if _, err := o.Raw(sql, tableName, userName, AdminRoles).QueryRows(&rows); err != nil {
		log.Error(err)
		return false
	}

	if len(rows) == 1 && rows[0].Cnt == 1 {
		return true
	}

	log.Warnf("用户名 %s 对表 %s 所对应的分组无管理员权限", userName, tableName, rows)
	return false
}

// UpdateUserRole 更新用户角色信息
func UpdateUserRole(user, group, role string) error {
	md5sum := md5.New()
	md5sum.Write([]byte(group))
	gpID := hex.EncodeToString(md5sum.Sum(nil))

	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	sql := `update users set role = ? where erp = ? and group_id = ?`
	if _, err := o.Raw(sql, role, user, gpID).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}
