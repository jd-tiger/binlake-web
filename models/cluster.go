package models

import (
	"github.com/juju/errors"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
)

// Cluster 集群信息 对应有zk 地址
type Cluster struct {
	Zone       string `orm:"size(128)"`
	Mark       string `orm:"size(254)"`
	Zk         string `orm:"size(511)"`
	Path       string `orm:"size(128)"`
	SlaveId    int64  `orm:"size(64)"`
	SlaveUuid  string `orm:"size(72)"`
	MqAddr     string `orm:"size(128)"`
	CreateTime string `orm:"size(32)"`
	UpdateTime string `orm:"size(32)"`
}

// DumpServer: dump server ip what we called wave
type DumpServer struct {
	Ip         string  `orm:"size(16)"`
	Zk         string  `orm:"size(511)"`
	MySQL      string  `orm:"size(128)"`
	Load       float64 `orm:"size(128)"`
	CreateTime string  `orm:"size(32)"`
	UpdateTime string  `orm:"size(32)"`
}

// ClusterList 返回 集群信息
func ClusterList() (cs []Cluster, err error) {
	o := orm.NewOrm()
	sql := `select cs.zone, cs.mark, cs.zk, cs.path, cs.slave_id, cs.slave_uuid, cs.create_time, cs.update_time, cs.url, cs.address from cluster cs`

	cs = []Cluster{}

	if _, err := o.Raw(sql).QueryRows(&cs); err != nil {
		log.Error(err)
		return cs, err
	}

	return cs, nil
}

// DumpServerList 显示dump 服务的详细信息
func DumpServerList(zk string) (ds []DumpServer, err error) {
	ds = []DumpServer{}

	if zk == "" {
		return ds, nil
	}

	o := orm.NewOrm()
	sql := `select ws.ip, ws.zk, ws.create_time, ws.update_time from waves ws where zk = ?`

	if _, err := o.Raw(sql, zk).QueryRows(&ds); err != nil {
		log.Error(err)
		return ds, err
	}

	return ds, nil
}

// CreateCluster 创建集群
func CreateCluster(zone, mark, zk, zkPath, slaveUuid string, slaveId int64, waves []string) error {
	// cluster sql
	cs := `insert into cluster(zone, mark, zk, path, slave_uuid, slave_id, create_time, update_time)
			values(?, ?, ?, ?, ?, ?, now(), now())`
	o := orm.NewOrm()
	o.Begin()
	defer o.Commit()

	ws := make(map[string]string)
	for _, w := range waves {
		ip := strings.TrimSpace(w)
		if len(ip) != 0 {
			ws[ip] = ip
		}
	}

	if len(ws) == 0 {
		log.Warn("no available wave nodes in cluster ")
		return errors.New("no available wave nodes in cluster ")
	}

	if _, err := o.Raw(cs, zone, mark, zk, zkPath, slaveUuid, slaveId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	// wave sql
	for _, w := range ws {
		if _, err := o.Raw("insert into waves(ip, zk, create_time, update_time)values(?, ?, now(), now())", w, zk).Exec(); err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

// UpdateCluster 更新集群信息
func UpdateCluster(zk string, ips []string) error {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	del := `delete from waves where zk = ?`

	if _, err := o.Raw(del, zk).Exec(); err != nil {
		log.Error(err)
		return err
	}

	ins := `insert into waves(ip, zk, create_time, update_time)values(?, ?, now(), now())`
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if len(ip) != 0 {
			if _, err := o.Raw(ins, ip, zk).Exec(); err != nil {
				log.Error(err)
				return err
			}
		}
	}

	return o.Commit()
}

// GetClusterByZk 获取集群信息
func GetClusterByZk(zk string) (cs []Cluster, err error) {
	o := orm.NewOrm()
	sql := `select cs.zone, cs.mark, cs.zk, cs.create_time, cs.update_time, mi.url, mi.address, ji.replica_url from cluster cs
			join mq_info mi on cs.zk = mi.zk join jed_info ji on cs.zk=ji.zk where cs.zk = ?`

	cs = []Cluster{}

	if _, err := o.Raw(sql, zk).QueryRows(&cs); err != nil {
		log.Error(err)
		return cs, err
	}

	return cs, nil
}

// GetClusterByZk 获取集群信息
func GetClusterByInstID(instId string) (cs []Cluster, err error) {
	o := orm.NewOrm()
	sql := `select cs.zone, cs.mark, cs.zk, cs.create_time, cs.update_time, mi.url, mi.address from cluster cs
			join mq_info mi on cs.zk = mi.zk where cs.zk in (select zk from instances where id = ?)`

	cs = []Cluster{}

	if _, err := o.Raw(sql, instId).QueryRows(&cs); err != nil {
		log.Error(err)
		return cs, err
	}

	return cs, nil
}
