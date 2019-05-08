package models

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

func initORM() {
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", "root",
		"secret",
		"127.0.0.1",
		"3358",
		"tower_v4",
		"utf8")

	orm.RegisterDataBase("default", "mysql", dataSource)
}

func TestSelectConfig(t *testing.T) {
	initORM()
	cs, err := SelectConfig(nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range cs {
		fmt.Println(c)
	}

	cs, err = SelectConfig(&Config{
		Keyword: "zh",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range cs {
		fmt.Println(c)
	}
}

func TestSaveConfig(t *testing.T) {
	initORM()
	err := SaveConfig(&Config{
		Keyword: "in",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestDeleteConfig(t *testing.T) {
	initORM()
	err := DeleteConfig(&Config{
		Keyword: "in",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestUpdateConfig(t *testing.T) {
	initORM()
	err := UpdateConfig(&Config{
		Keyword: "in",
		DbsApi: "http://dbs.jd.com",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestConfigJson(t *testing.T) {
	c := &Config{
		Keyword: "in",
		DbsApi: "http://dbs.jd.com",
	}

	bt, err := json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(bt))
}
