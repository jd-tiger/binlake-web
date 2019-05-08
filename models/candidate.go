package models

import (
	"errors"
	"math/rand"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

// Candidate 候选节点 根据zk 所属集群分布
type Candidate struct {
	Ip         string `orm:"size(36)"`
	Zk         string `orm:"size(512)"`
	CreateTime string `orm:"size(32)"`
	UpdateTime string `orm:"size(32)"`
}

var (
	// candidate number
	CandNum = 3
)

// GetCandidate 通过zk 集群获取candidate ip
func GetCandidate(zk string) ([]string, error) {
	o := orm.NewOrm()

	sql := "select ip, zk, create_time, update_time from waves where zk = ?"

	var cands []Candidate
	num, err := o.Raw(sql, zk).QueryRows(&cands)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if num == 0 {
		err := errors.New("no candidate get from table waves")
		return nil, err
	}

	var ips []string
	for _, can := range cands {
		ips = append(ips, can.Ip)
	}

	return randomTake(ips), nil
}

// randomTake 目前随机选择三个节点
func randomTake(total []string) []string {
	if len(total) <= CandNum {
		// 总的ip 数目 小等于 规定的候选节点数据 则返回所有
		return total
	}

	// 这里需要初始化随机种子 否则每次都会是同样的结果
	src := rand.NewSource(time.Now().Unix())
	rd := rand.New(src)

	choose := make(map[string]string)

	for len(choose) < CandNum {
		index := rd.Intn(len(total) - 1)
		choose[total[index]] = total[index]
	}

	var sel []string
	for key := range choose {
		sel = append(sel, key)
	}

	return sel
}

// IPPrefix ip前缀 授权的ip段
func IPPrefix() ([]string, error) {
	o := orm.NewOrm()
	sql := `select concat(substring_index(ip, '.', 1), '.%') prefix from waves group by prefix`
	type IPPrefix struct {
		Prefix string `orm:"size(10)"`
	}

	var authHosts []string

	var ps []IPPrefix
	if _, err := o.Raw(sql).QueryRows(&ps); err != nil {
		log.Error(err)
		return authHosts, err
	}

	if len(ps) == 0 {
		return authHosts, errors.New("授权ip 为空,请联系管理员添加")
	}

	for _, prefix := range ps {
		authHosts = append(authHosts, prefix.Prefix)
	}
	return authHosts, nil
}
