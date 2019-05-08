package main

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/models"
	_ "github.com/jd-tiger/binlake-web/routers"
)

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
		beego.AppConfig.String("meta::user"),
		beego.AppConfig.String("meta::passwd"),
		beego.AppConfig.String("meta::host"),
		beego.AppConfig.String("meta::port"),
		beego.AppConfig.String("meta::dbname"),
		beego.AppConfig.String("meta::charset"))

	log.Infof("dataSource:%v", dataSource)
	orm.RegisterDataBase("default", "mysql", dataSource)

	models.LoadMessage()
}

func main() {
	orm.Debug = true
	o := orm.NewOrm()
	o.Using("default") // 默认使用 default，你可以指定为其他数据库
	beego.Run()
}
