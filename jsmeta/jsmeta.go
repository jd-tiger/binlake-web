package jsmeta

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

/***************************************
meta base info
****************************************/
type DbInfo struct {
	SlaveId     int64	`json:"slaveId"`	// slave id
	Host        string	`json:"host"`		// MySQL host to connect
	Port        int32	`json:"port"`		// MySQL sever port
	User        string	`json:"user"`		// MySQL replication user
	Password    string	`json:"password"`	// MySQL replication password
	CreateTime  int64	`json:"createTime"`	// node created timestamp
	State       NodeState	`json:"state"`		// node state
	Rule        []*Rule	`json:"rule"`		// rule
	MetaVersion int64	`json:"metaVersion"`	// meta version increate if update by manager-server
	SlaveUUID   string	`json:"slaveUUID"`	// slave uuid
}

/***************************************
why separate the meta info into two parts?
the base information upper is read only in most circumstance
this lower information BinlogInfo writes often so let them alone

meta dynamic info that have to update often
****************************************/
type BinlogInfo struct {
	BinlogFile       string			`json:"binlogFile"`		// MySQL replication binlog file name
	BinlogPos        int64			`json:"binlogPos"`		// MySQL replication binlog position
	ExecutedGtidSets string			`json:"executedGtidSets"`	// MySQL replication with gtid sets
	Leader           string			`json:"leader"`			// leader info
	LeaderVersion    int64			`json:"leaderVersion"`		// version for binlog info first set to 0
	BinlogWhen       int64			`json:"binlogWhen"`		// binlog dump timestamp : when
	TransactionId    int64			`json:"transactionId"`
	InstanceIp       string			`json:"instanceIp"`		// ip in relation to MySQL instance
	WithGTID         bool			`json:"withGTID"`
	PreLeader        string			`json:"preLeader"`		// previous leader
	Msgid            map[string]int64	`json:"msgid"`			// table message id
}

/**
counter is save retry times
**/
type Counter struct {
	RetryTimes int64 `json:"retryTimes"`	// times that Wave Server try to connect MySQL server
	KillTimes  int64 `json:"killTimes"`	// times that Wave Server try to connect MySQL server
}

type BindLeaders struct {
	Leaders    []*Pair  `json:"leaders"`
	RetryTimes int64    `json:"retryTimes"`
}

type Terminal struct {
	BinlogFile string	`json:"binlogFile"`
	BinlogPos  int64	`json:"binlogPos"`
	Gtid       string	`json:"gtid"`
	NewHost    []string	`json:"newHost"`	// 对应 reshard  之后新的host
}

type Candidate struct {
	Host []string	`json:"host"`
}

/***
** 规则 包括规则类型 具体的规则 mq/kv/etc.
*****/
type Rule struct {
	Storage      StorageType	`json:"storage"`	// 存储类型
	ConvertClass string		`json:"convertClass"`	// 消息格式转换器类名
	Any          []byte		`json:"any"`		// 存储类型对应的具体参数， 规则， 类型使用方式 etc..
}

/**
topic rules map :

one topic in relation to many rules

one rule in relation to many topic

**/
type MQRule struct {
	Topic            string		`json:"topic"`			// mq topic
	WithTransaction	 bool		`json:"withTransaction"`	// 是否携带事务信息 begin or commit message
	WithUpdateBefore bool		`json:"withUpdateBefore"`	// update事件信息是否携带变更前数据
	ProducerClass    string		`json:"producerClass"`		// producer class name with constructor parameter List<Pair> paras
	Order            OrderType	`json:"order"`			// topic类型：顺序主题或是非顺序主题
	Para             []*Pair	`json:"para"`			// MQ 链接参数
	White            []*Filter	`json:"white"`			// 白名单
	Black            []*Filter	`json:"black"`			// 黑名单
}

// 过滤器 说明: 列过滤只在白名单当中生效 并且列过滤可以使用黑名单过滤 或者白名单过滤
type Filter struct {
	Table      string      `json:"table"`		// 正则表达式 表名全称
	EventType  []EventType `json:"eventType"`	// 事件类型
	White      []*Column   `json:"white"`		// 过滤列白名单字段
	Black      []*Column   `json:"black"`		// 过滤列黑名单字段
	FakeColumn []*Pair     `json:"fakeColumn"`	// 伪列字段
	HashKey    []string    `json:"hashKey"`		// 业务主键
}

type Column struct {
	Name  string   `json:"name"`	// 列名
	Value []string `json:"value"`
}

type Pair struct {
	Key   string   `json:"key"`
	Value string   `json:"value"`
}

type NodeState string

const (
	NodeState_ONLINE NodeState = "ONLINE"
	NodeState_OFFLINE NodeState = "OFFLINE"
)

var NodeState_value = map[string]NodeState{
	"ONLINE": NodeState_ONLINE,
	"OFFLINE": NodeState_OFFLINE,
}

// 存储类型
type StorageType string

const (
	// 消息队列 规则
	StorageType_MQ_STORAGE StorageType = "MQ_STORAGE"
	// KV storage
	StorageType_KV_STORAGE StorageType = "KV_STORAGE"
)

var StorageType_value = map[string]StorageType{
	"MQ_STORAGE": StorageType_MQ_STORAGE,
	"KV_STORAGE": StorageType_KV_STORAGE,
}

// 消息顺序类型
type OrderType string

const (
	// 完全乱序 无规则
	OrderType_NO_ORDER OrderType = "NO_ORDER"
	// 业务主键级别消息顺序 partition 对应多个
	OrderType_BUSINESS_KEY_ORDER OrderType = "BUSINESS_KEY_ORDER"
	// 表级别消息顺序 partition 对应多个
	OrderType_TABLE_ORDER OrderType = "TABLE_ORDER"
	// 库级别消息 顺序
	OrderType_DB_ORDER OrderType = "DB_ORDER"
	// 实例级别消息顺序 broker 对应一个
	OrderType_INSTANCE_ORDER OrderType = "INSTANCE_ORDER"
)

var OrderType_value = map[string]OrderType{
	"NO_ORDER": OrderType_NO_ORDER,
	"BUSINESS_KEY_ORDER": OrderType_BUSINESS_KEY_ORDER,
	"TABLE_ORDER": OrderType_TABLE_ORDER,
	"DB_ORDER": OrderType_DB_ORDER,
	"INSTANCE_ORDER": OrderType_INSTANCE_ORDER,
}

/** 事件类型 **/
type EventType string

const (
	EventType_OTHER EventType = "OTHER"
	EventType_INSERT EventType = "INSERT"
	EventType_UPDATE EventType = "UPDATE"
	EventType_DELETE EventType = "DELETE"
	EventType_CREATE EventType = "CREATE"
	EventType_ALTER EventType = "ALTER"
	EventType_ERASE EventType = "ERASE"
	EventType_QUERY EventType = "QUERY"
	EventType_TRUNCATE EventType = "TRUNCATE"
	EventType_RENAME EventType = "RENAME"
	EventType_CINDEX EventType = "CINDEX"
	EventType_DINDEX EventType = "DINDEX"
)

var EventType_value = map[string]EventType{
	"OTHER": EventType_OTHER,
	"INSERT": EventType_INSERT,
	"UPDATE": EventType_UPDATE,
	"DELETE": EventType_DELETE,
	"CREATE": EventType_CREATE,
	"ALTER": EventType_ALTER,
	"ERASE": EventType_ERASE,
	"QUERY": EventType_QUERY,
	"TRUNCATE": EventType_TRUNCATE,
	"RENAME": EventType_RENAME,
	"CINDEX": EventType_CINDEX,
	"DINDEX": EventType_DINDEX,
}

// zookeeper 信息结构体
type ZK struct {
	Servers string   `json:"servers"`	// zk servers地址
	Path    string   `json:"path"`		// 监听的根路径
}

// MetaData 页面与管理端交互用的元数据信息
type MetaData struct {
	DbInfo    *DbInfo     `json:"dbInfo"`
	Slave     *BinlogInfo `json:"slave"`
	Master    *BinlogInfo `json:"master"`
	Counter   *Counter    `json:"counter"`
	Terminal  *Terminal   `json:"terminal"`
	Candidate []string    `json:"candidate"`
	Zk        *ZK         `json:"zk"`
}

func GzipCompress(data []byte) ([]byte, error) {
	var (
		buffer bytes.Buffer
		out    []byte
	)
	writer := gzip.NewWriter(&buffer)
	_, err := writer.Write(data)
	if err != nil {
		writer.Close()
		return out, err
	}
	err = writer.Close()
	if err != nil {
		return out, err
	}

	return buffer.Bytes(), nil
}

func GzipUnCompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		var out []byte
		return out, err
	}
	defer reader.Close()

	return ioutil.ReadAll(reader)
}
