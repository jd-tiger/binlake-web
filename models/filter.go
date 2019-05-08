package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

// Filter 过滤器表
type Filter struct {
	Id           string `orm:"size(36)"`
	Name         string `orm:"size(64)"`
	Group        string `orm:"size(128)"`
	GroupId      string `orm:"size(36)"`
	Type         string `orm:"size(16)"`
	TableName    string `orm:"size(255)"`
	Events       string `orm:"size(128)"`
	WhiteColumns string `orm:"size(1024)"`
	BlackColumns string `orm:"size(1024)"`
	FakeCols     string `orm:"size(128)"`
	BusinessKeys string `orm:"size(128)"`
	CreateTime   string `orm:"size(32)"`
	UpdateTime   string `orm:"size(32)"`
}

// FilterList 查询当前分组下所有过滤规则 至少返回空数组 与用户无关
func FilterList(group, tablePrefix string) (filters []Filter, err error) {
	filters = []Filter{}
	if group == "" {
		return filters, nil
	}

	o := orm.NewOrm()

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	sql := `select ft.group_id, gs.name 'group', ft.id, ft.name, ft.type, ft.table_name, ft.events,
			ft.white_columns, ft.black_columns, ft.fake_cols,
			ft.business_keys, ft.create_time, ft.update_time from filter ft
			join users us on us.group_id = ft.group_id
			join groups gs on us.group_id = gs.id
			where us.group_id = ? and ft.table_name like ? group by ft.group_id, ft.id`

	if _, err := o.Raw(sql, groupId, "%"+tablePrefix+"%").QueryRows(&filters); err != nil {
		log.Error(err)
		return filters, err
	}

	return filters, nil
}

// FilterAdminList 查询当前登录用户下所有分组中包含的过滤规则 至少返回空数组
func FilterAdminList(loginErp string) (filters []Filter, err error) {
	filters = []Filter{}
	o := orm.NewOrm()

	sql := `select ft.group_id, ft.id, ft.name, ft.type, ft.table_name, ft.events,
			ft.white_columns, ft.black_columns, ft.fake_cols,
			ft.business_keys, ft.create_time, ft.update_time from filter ft
			join users us on us.group_id = ft.group_id
				where us.group_id in (select group_id from users where erp = ? and role in ('admin' , 'creator')) group by ft.group_id, ft.id`

	if _, err := o.Raw(sql, loginErp).QueryRows(&filters); err != nil {
		log.Error(err)
		return filters, err
	}

	return filters, nil
}

// SearchFilter 查询用户所有有权限分组下的过滤规则 至少返回空数组
func SearchFilter(ui *UserInfo, tablePrefix string) (filters []Filter, err error) {
	filters = []Filter{}

	o := orm.NewOrm()

	erp := ui.UserName
	if ui.IsSuperAdmin {
		erp = "%"
	}

	sql := `select ft.group_id, gs.name 'group', ft.id, ft.name, ft.type, ft.table_name, ft.events,
			ft.white_columns, ft.black_columns, ft.fake_cols,
			ft.business_keys, ft.create_time, ft.update_time from filter ft
			join users us on us.group_id = ft.group_id
			join groups gs on us.group_id = gs.id
			where us.erp like ? and ft.table_name like ? group by ft.group_id, ft.id`

	if _, err := o.Raw(sql, erp, "%"+tablePrefix+"%").QueryRows(&filters); err != nil {
		log.Error(err)
		return filters, err
	}

	return filters, nil
}

// CreateFilter 创建当前分组的过滤规则
func CreateFilter(name, group, filterType, table, fakeCols, whiteColumns, blackColumns, businessKeys string, events []string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	sql := `insert into filter(id, name, group_id, type, table_name, events, white_columns, black_columns,
			fake_cols, business_keys, create_time, update_time)values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), now())`

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	md5sum.Reset()
	md5sum.Write([]byte(groupId + filterType + table))
	id := hex.EncodeToString(md5sum.Sum(nil))

	if _, err := o.Raw(sql, id, name, groupId, filterType, table, strings.Join(events, ","),
		upRmSpace(whiteColumns), upRmSpace(blackColumns), fakeCols, upRmSpace(businessKeys)).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}

// UpdateFilter 更新当前分组的过滤规则
func UpdateFilter(oldFilterId, group, name, filterType, table, fakeCols, whiteColumns, blackColumns, businessKeys string, events []string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	del := `delete from filter where id = ?`
	if _, err := o.Raw(del, oldFilterId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	sql := `insert into filter(id, name, group_id, type, table_name, events, white_columns, black_columns,
			fake_cols, business_keys, create_time, update_time)values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), now())`

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	md5sum.Reset()
	md5sum.Write([]byte(fmt.Sprintf("%s%s%s", groupId, filterType, table)))
	newFilterId := hex.EncodeToString(md5sum.Sum(nil))

	if _, err := o.Raw(sql, newFilterId, name, groupId, filterType, table, strings.Join(events, ","),
		upRmSpace(whiteColumns), upRmSpace(blackColumns), fakeCols, upRmSpace(businessKeys)).Exec(); err != nil {
		log.Error(err)
		return err
	}

	// 更新mq rule 当中过滤规则id 对应的值
	if newFilterId != oldFilterId {
		// 如果两个id 不相等则 需要更新mq——rule 表当中数据
		sql = `update mq_filter_relation set filter_id = ? where filter_id = ?`
		if _, err := o.Raw(sql, newFilterId, oldFilterId).Exec(); err != nil {
			log.Error(err)
			return err
		}
	}

	return o.Commit()
}

// DeleteFilter 从过滤器表当中删除
func DeleteFilter(id string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	del := `delete from filter where id = ?`
	if _, err := o.Raw(del, id).Exec(); err != nil {
		log.Error(err)
		return err
	}

	// 需要创建索引
	del = `delete from mq_filter_relation where filter_id = ?`
	if _, err := o.Raw(del, id).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}

// upRmSpace : upper parameters and remove space
func upRmSpace(para string) string {
	cols := strings.Join(strings.FieldsFunc(para, unicode.IsSpace), ",")
	return strings.ToUpper(cols)
}
