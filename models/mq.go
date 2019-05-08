package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/jsmeta"
)

// MQRule mq 规则
type MQRule struct {
	Id            string `orm:"size(16)"`
	Group         string `orm:"size(128)"`
	Topic         string `orm:"size(16)"`
	GroupId       string `orm:"size(36)"`
	FilterId      string `orm:"size(36)"`
	Name          string `orm:"size(128)"`
	WithTrx       string `orm:"size(64)"`
	ProducerClass string `orm:"size(64)"`
	OrderType     string `orm:"size(64)"`
	UpdateTime    string `orm:"size(32)"`
	CreateTime    string `orm:"size(32)"`
}

// MQInfo mq information
type MQInfo struct {
	Zk         string `orm:"size(512)"`
	Url        string `orm:"size(128)"`
	Address    string `orm:"size(128)"`
	CreateTime string `orm:"size(64)"`
	UpdateTime string `orm:"size(64)"`
}

// MQRuleList 查询当前登录用户下所有分组中包含的过滤规则 至少返回空数组
func MQRuleList(group, topic string) (mrs []MQRule, err error) {
	mrs = []MQRule{}
	if group == "" {
		return mrs, nil
	}

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	o := orm.NewOrm()

	sql := `select mr.id, gs.name 'group', mr.group_id, mr.topic, mr.name, mr.with_trx, mr.producer_class,
			mr.order_type, mfr.filter_id, mr.create_time, mr.update_time from mq_rule mr
			join groups gs on gs.id = mr.group_id
			join mq_filter_relation mfr on mfr.mq_id = mr.id
			where mr.group_id = ? and mr.topic like ? group by mr.id, mr.group_id, mfr.filter_id`

	if _, err := o.Raw(sql, groupId, "%"+topic+"%").QueryRows(&mrs); err != nil {
		log.Error(err)
		return mrs, err
	}

	return mrs, nil
}

// SearchMQRule 查询用户所有有权限分组下的topic 至少返回空数组
func SearchMQRule(ui *UserInfo, topic string) (mrs []MQRule, err error) {
	mrs = []MQRule{}

	o := orm.NewOrm()

	erp := ui.UserName
	if ui.IsSuperAdmin {
		erp = "%"
	}

	sql := `select mr.id, gs.name 'group', mr.group_id, mr.topic, mr.name, mr.with_trx, mr.producer_class,
			mr.order_type, mfr.filter_id, mr.create_time, mr.update_time from mq_rule mr
			join groups gs on gs.id = mr.group_id
			join users us on us.group_id = gs.id
			join mq_filter_relation mfr on mfr.mq_id = mr.id
			where us.erp like ? and mr.topic like ? group by mr.id, mr.group_id, mfr.filter_id`

	if _, err := o.Raw(sql, erp, "%"+topic+"%").QueryRows(&mrs); err != nil {
		log.Error(err)
		return mrs, err
	}

	return mrs, nil
}

// CreateMQ 创建mq 以及绑定相应的规则
func CreateMQ(topic, name, order, group, producer string, filterIds []string, withTrx bool) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	sqlMr := `insert into mq_rule(id, topic, group_id, name, with_trx,
			producer_class, order_type, update_time, create_time) values(?, ?, ?, ?, ?, ?, ?, now(), now())`

	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	md5sum.Reset()
	md5sum.Write([]byte(topic + groupId))
	mqId := hex.EncodeToString(md5sum.Sum(nil))

	if _, err := o.Raw(sqlMr, mqId, topic, groupId, name, strconv.FormatBool(withTrx), producer, order).Exec(); err != nil {
		log.Error(err)
		return err
	}

	sqlMfr := `insert into mq_filter_relation(id, mq_id, filter_id, create_time) values(?, ?, ?, now())`

	for _, filterId := range filterIds {
		md5sum.Reset()
		md5sum.Write([]byte(mqId + filterId))
		mqFilterRelId := hex.EncodeToString(md5sum.Sum(nil))
		if _, err := o.Raw(sqlMfr, mqFilterRelId, mqId, filterId).Exec(); err != nil {
			log.Error(err)
			return err
		}
	}

	return o.Commit()
}

// DeleteMQ 删除 mq 规则id
func DeleteMQ(id string) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	// 删除 mq 记录
	sql := `delete from mq_rule where id = ?`
	if _, err := o.Raw(sql, id).Exec(); err != nil {
		log.Error(err)
		return err
	}

	// 删除mq过滤器关联记录
	sql = `delete from mq_filter_relation where mq_id = ?`
	if _, err := o.Raw(sql, id).Exec(); err != nil {
		log.Error(err)
		return err
	}

	// 删除实例mq关联记录
	sql = `delete from rule where rule_id = ?`
	if _, err := o.Raw(sql, id).Exec(); err != nil {
		log.Error(err)
		return err
	}

	return o.Commit()
}

// UpdateMQ 更新整个topic 对应的规则已经与 instance 关联
func UpdateMQ(topic, name, order, group, producer string, filterIds []string, withTrx bool) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer o.Rollback()

	// 获取需要变更的规则数据
	type Rule struct {
		GroupId      string `orm:"size(36)"`
		InstanceId   string `orm:"size(36)"`
		StorageType  string `orm:"size(36)"`
		ConvertClass string `orm:"size(36)"`
	}
	var rs []Rule

	sel := `select re.group_id, re.instance_id, re.storage_type, re.convert_class from rule re
			join mq_rule mr on mr.id = re.rule_id and mr.group_id = re.group_id
			where mr.topic = ?`

	if _, err := o.Raw(sel, topic).QueryRows(&rs); err != nil {
		log.Error(err)
		return err
	}

	// 删除mq topic表 在 rule 表当中记录
	md5sum := md5.New()
	md5sum.Write([]byte(group))
	groupId := hex.EncodeToString(md5sum.Sum(nil))

	md5sum.Reset()
	md5sum.Write([]byte(fmt.Sprintf("%s%s", topic, groupId)))
	mqId := hex.EncodeToString(md5sum.Sum(nil))

	// 更新 mq topic表当中的topic
	upd := `update mq_rule set name = ?, with_trx = ?, producer_class = ?,
			order_type = ? where id = ?`
	if _, err := o.Raw(upd, name, strconv.FormatBool(withTrx), producer, order, mqId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	// 删除mq_filter_relation表当中关于关联当前topic的过滤器关系
	del := `delete from mq_filter_relation where mq_id = ?`
	if _, err := o.Raw(del, mqId).Exec(); err != nil {
		log.Error(err)
		return err
	}

	// 将新的topic 与 过滤器关系 写入到mq-rule 表当中
	sql := `insert into mq_filter_relation(id, mq_id, filter_id, create_time) values(?, ?, ?, now())`

	for _, fid := range filterIds {
		md5sum.Reset()
		md5sum.Write([]byte(fmt.Sprintf("%s%s", mqId, fid)))
		mqFilterRelId := hex.EncodeToString(md5sum.Sum(nil))

		if _, err := o.Raw(sql, mqFilterRelId, mqId, fid).Exec(); err != nil {
			log.Error(err)
			return err
		}
	}

	return o.Commit()
}

func QueryMQInfoByZk(zk string) ([]MQInfo, error) {
	o := orm.NewOrm()

	var rs []MQInfo

	sql := `select zk, url, address, create_time, update_time from mq_info where zk = ?`
	if _, err := o.Raw(sql, zk).QueryRows(&rs); err != nil {
		log.Error(err)
		return nil, err
	}

	return rs, nil
}

// ClientID take client id
func ClientID(topic, host string, port int) *jsmeta.Pair {
	//jmq4必须设置client.id 并且保证同一个服务上不同producer的client.id不同
	clientId := topic + "-" + strings.TrimSuffix(host, ".mysql.jddb.com") + "-" + strconv.Itoa(port)
	return &jsmeta.Pair{
		Key:   "client.id",
		Value: clientId,
	}
}

// FixedMQParas 获取mq 生产参数
func FixedMQParas() ([]*jsmeta.Pair, error) {
	var paras []*jsmeta.Pair

	paras = append(paras, &jsmeta.Pair{
		Key:   "acks",
		Value: "1",
	})

	paras = append(paras, &jsmeta.Pair{
		Key:   "request.timeout.ms", // 如果数值越大 某一个broker发生异常 则等待时间越长
		Value: "1000",
	})

	paras = append(paras, &jsmeta.Pair{
		Key:   "batch.size",
		Value: "0",
	})

	paras = append(paras, &jsmeta.Pair{
		Key:   "key.serializer",
		Value: "org.apache.kafka.common.serialization.ByteArraySerializer",
	})

	paras = append(paras, &jsmeta.Pair{
		Key:   "value.serializer",
		Value: "org.apache.kafka.common.serialization.ByteArraySerializer",
	})

	return paras, nil
}
