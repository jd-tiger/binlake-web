package models

import (
	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

// UserInfo erp中用户信息
type UserInfo struct {
	ID           int64
	Status       int
	Email        string `json:"email"`
	Mobile       string `json:"mobile"`
	HrmDeptID    string `json:"hrmDeptId"`
	PersonID     string `json:"personId"`
	OrgID        string `json:"orgId"`
	OrgName      string `json:"orgName"`
	User         string `json:"fullname"`
	UserID       int    `json:"userId"`
	UserName     string `json:"username"`
	IsSuperAdmin bool   `json:"isAdmin,omitempty"`
}

// IsSuperAdmin 判断是否属于超级用户
func IsSuperAdmin(erp string) bool {
	o := orm.NewOrm()
	sql := `select erp, create_time, update_time from super_admin where erp = ?`
	type SuperAdmin struct {
		Erp        string
		CreateTime string
		UpdateTime string
	}

	rst := &[]SuperAdmin{}

	_, err := o.Raw(sql, erp).QueryRows(rst)
	if err != nil {
		log.Error(err)
		return false
	}

	// 按照返回的行数计算
	return len(*rst) == 1
}
