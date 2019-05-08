package models

import (
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

// Config 表信息
type Config struct {
	Keyword    string   `json:"keyword"`    // keyword 关键字
	Mark       string   `json:"mark"`       // mark 标注
	Dns        string   `json:"dns"`        // wave 集群对应 域名
	Zk         string   `json:"zk"`         // zk zookeeper 地址
	Path       string   `json:"path"`       // zk root path
	SlaveId    string   `json:"slaveid"`    // slaveId
	SlaveUuid  string   `json:"slaveuuid"`  // slave uudi
	DbsApi     string   `json:"dbsurl"`     // dbs dbs api 地址前缀 负责流程审批 授权 主从架构查询
	DbsToken   string   `json:"dbstoken"`   // dbs url header token
	MqAddr     string   `json:"mqaddr"`     // jmqtcp jmq 生产者地址 不同机房 不同集群 对应的生产者地址可能不一样
	Waves      []string `json:"waves"`      // wave ips
	CreateTime string   `json:"createtime"` // createtime 创建时间
	UpdateTime string   `json:"updatetime"` // updatetime 更新时间
}

// SaveConfig 保存config信息
func SaveConfig(c *Config) error {
	sql := `insert into config(keyword, mark, dns, zk, path, slave_id, slave_uuid, dbs_api, dbs_token, mq_addr, create_time, update_time)
			values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), now())`
	o := orm.NewOrm()
	o.Begin()
	defer o.Commit()

	log.Debug("save config ", c, " sql, ", sql)

	if _, err := o.Raw(sql, c.Keyword, c.Mark, c.Dns, c.Zk, c.Path, c.SlaveId, c.SlaveUuid, c.DbsApi, c.DbsToken, c.MqAddr).Exec(); err != nil {
		log.Error(err)
		return err
	}

	for _, w := range c.Waves {
		if _, err := o.Raw(`insert into waves(ip, zk, create_time, update_time) values (?, ?, now(), now())`, w, c.Zk).Exec(); err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

// DeleteConfig 删除config信息
func DeleteConfig(c *Config) error {
	sql := `delete from config where keyword = ?`
	o := orm.NewOrm()

	o.Begin()
	defer o.Commit()

	log.Debug("delete config ", c, " sql, ", sql)

	if _, err := o.Raw(sql, c.Keyword).Exec(); err != nil {
		log.Error(err)
		return err
	}

	if _, err := o.Raw(`delete from waves where zk = ?`, c.Zk).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// UpdateConfig 更新config信息
func UpdateConfig(c *Config) error {
	sql := `update config set mark = ?, dns = ?, zk = ?, path = ?, slave_id = ?, slave_uuid = ?, dbs_api = ?, dbs_token = ?, mq_addr = ? where keyword = ?`
	o := orm.NewOrm()

	o.Begin()
	defer o.Commit()

	log.Debug("update config ", c, " sql, ", sql)

	if _, err := o.Raw(sql, c.Mark, c.Zk, c.Path, c.SlaveId, c.SlaveUuid, c.DbsApi, c.DbsToken, c.MqAddr, c.Keyword).Exec(); err != nil {
		log.Error(err)
		return err
	}

	if _, err := o.Raw(`delete from waves where zk = ?`, c.Zk).Exec(); err != nil {
		log.Error(err)
		return err
	}

	for _, w := range c.Waves {
		if _, err := o.Raw(`insert into waves(ip, zk, create_time, update_time) values(?, ?, now(), now())`, w, c.Zk).Exec(); err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

// SelectConfig 查询config信息
func SelectConfig(c *Config) ([]*Config, error) {
	sql := `select keyword, mark, dns, zk, path, slave_id, slave_uuid, dbs_api, dbs_token, mq_addr, create_time, update_time from config `

	var cons []string
	if c != nil && c.Keyword != "" {
		cons = append(cons, "keyword = '"+c.Keyword+"'")
	}

	if c != nil && c.Zk != "" {
		cons = append(cons, "zk = '"+c.Zk+"'")
	}

	if c != nil && c.Path != "" {
		cons = append(cons, "path = '"+c.Path+"'")
	}

	if c != nil && c.SlaveId != "" {
		cons = append(cons, "slave_id = '"+c.SlaveId+"'")
	}

	if c != nil && c.SlaveUuid != "" {
		cons = append(cons, "slave_uuid = '"+c.SlaveUuid+"'")
	}

	if c != nil && c.DbsApi != "" {
		cons = append(cons, "dbs_api = '"+c.DbsApi+"'")
	}

	if c != nil && c.MqAddr != "" {
		cons = append(cons, "mq_addr = '"+c.MqAddr+"'")
	}

	o := orm.NewOrm()
	if len(cons) != 0 {
		sql = sql + " where " + strings.Join(cons, " and ")
	}

	var cs []*Config
	log.Debug(sql, ", paras ", c)
	if _, err := o.Raw(sql).QueryRows(&cs); err != nil {
		log.Error(err)
		return cs, err
	}

	if cs == nil {
		return []*Config{}, nil
	}

	return cs, nil
}

// Waves 查询相关 wave ip
func Waves(c *Config) ([]string, error) {
	sql := `select ip from waves where zk = ?`
	o := orm.NewOrm()

	var ips []string

	if _, err := o.Raw(sql, c.Zk).QueryRows(&ips); err != nil {
		log.Error(err)
		return ips, err
	}

	return ips, nil
}
