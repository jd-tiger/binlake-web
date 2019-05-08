package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/zssky/log"
	"github.com/zssky/tc/http"
)

const (
	// HttpOkResponse http 请求返回成功
	HttpOkResponse = 200
)

// DBS dbs相关操作结构提封装
type DBS struct {
	timeout     time.Duration
}

// grant 授权结构体
type Grant struct {
	Host         string `orm:"size(128)"`
	Port         string `orm:"size(32)"`
	Type         string `orm:"size(32)"`
	TableName    string `orm:"size(128)"`
	WhiteColumns string `orm:"size(1024)"`
	BlackColumns string `orm:"size(1024)"`
	Events       string `orm:"size(254)"`
	CreateTime   string `orm:"size(64)"`
	UpdateTime   string `orm:"size(64)"`
}

// CallBack 回调结构体
type CallBack struct {
	ApproveStatus int    `json:"approveStatus"`
	ApproveUser   string `json:"approveUser"`
	Comment       string `json:"comment"`
	TowerOrderId  int    `json:"towerId"`
	DbsOrderId    int    `json:"dbsOrderId"`
}

// GrantList 获取需要授权的列表
func GrantList(instanceIds []string) (grants []Grant, err error) {
	o := orm.NewOrm()

	sel := `select it.host, it.port, fe.type, fe.table_name, fe.white_columns, fe.black_columns, fe.events
			from instances it join rule re
			on re.instance_id = it.id join mq_rule mr
			on mr.id = re.rule_id join mq_filter_relation mfr
			on mr.id = mfr.mq_id join filter fe
			on fe.id = mfr.filter_id where it.id in ('` +
		strings.Join(instanceIds, "','") + `') group by it.host, it.port
				order by it.host, it.port`
	grants = []Grant{}
	if _, err := o.Raw(sel).QueryRows(&grants); err != nil {
		log.Error(err)
		return grants, err
	}

	return grants, err
}

// NewDBS 初始化dbs
func NewDBS() *DBS {
	dbs := &DBS{
	}
	dbs.timeout = time.Second * time.Duration(120)
	return dbs
}


// CheckSlave 从库检查 是否为从库 refer to doc:{ https://cf.jd.com/pages/viewpage.action?pageId=80085879}
func (d *DBS) CheckSlave(dbsUrl, token string, ips map[string]int) (bool, error) {
	url := fmt.Sprintf("%s/grant/verifymaster/", dbsUrl)

	// HostDetail dbs 请求从库检查需要的结构
	type DetailReq struct {
		Index int    `json:"dbidx"`
		Host  string `json:"host"`
		Port  int    `json:"port"`
	}

	// Hosts dbs 请求检查从库检查需要的结构
	type HostsReq struct {
		Hosts []DetailReq `json:"hosts"`
	}

	req := &HostsReq{}
	index := 0
	for k, v := range ips {
		req.Hosts = append(req.Hosts, DetailReq{
			Index: index,
			Host:  k,
			Port:  v,
		})
		index += 1
	}

	// {"Status":400,"Message":"need Token in Header"}, code:400, err:<nil>
	data, code, err := http.PostJSON(url, token, req, d.timeout, d.timeout)
	log.Debugf("end url:%v, data:%s, code:%d, err:%v", url, data, code, err)
	if err != nil || code != HttpOkResponse {
		return false, fmt.Errorf("%v:%s", err, string(data))
	}

	// 成功返回 0
	success := 0

	// 从库检查返回结构体
	type DetailResp struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		IsMaster bool   `json:"isMaster"`
	}

	type CheckSlaveResp struct {
		Code    int          `json:"code"`
		Message string       `json:"message"`
		Hosts   []DetailResp `json:"hosts"`
	}

	resp := &CheckSlaveResp{
		Code: -1, // 默认初始值
	}
	if err := json.Unmarshal(data, resp); err != nil {
		log.Error(err)
		return false, err
	}

	if resp.Code != success {
		log.Error(resp.Message)
		return false, errors.New(resp.Message)
	}

	for _, h := range resp.Hosts {
		if h.IsMaster {
			// 如果是master
			return false, fmt.Errorf("host %s:%d is on the master role", h.Host, h.Port)
		}
	}

	return true, nil
}

// Grant 请求接口授权 参考文档 {https://cf.jd.com/pages/viewpage.action?pageId=80085879}
func (d *DBS) Grant(dbsUrl, token string, ips map[string]int, authHosts []string) (bool, error) {
	url := fmt.Sprintf("%s/grant/binlake", dbsUrl)

	type DetailReq struct {
		DbsID int64  `json:"dbsid"`
		Host  string `json:"host"`
		Port  int    `json:"port"`
	}

	type GrantReq struct {
		RequestID int64       `json:"requestId"`
		AuthHosts []string    `json:"authHosts"`
		Hosts     []DetailReq `json:"hosts"`
	}

	// 初始化grant 结构体
	req := &GrantReq{}
	req.RequestID = time.Now().Unix()
	req.AuthHosts = authHosts

	// 初始化host
	index := 0
	for k, v := range ips {
		req.Hosts = append(req.Hosts, DetailReq{
			DbsID: req.RequestID,
			Host:  k,
			Port:  v,
		})
		index += 1
	}

	// {"Status":400,"Message":"need Token in Header"}, code:400, err:<nil>
	data, code, err := http.PostJSON(url, token, req, d.timeout, d.timeout)
	log.Debugf("end url:%v, data:%s, code:%d, err:%v", url, data, code, err)
	if err != nil || code != HttpOkResponse {
		return false, fmt.Errorf("%v:%s", err, string(data))
	}

	// 成功返回 0
	success := 0

	// 请求授权接口返回的结构
	type GrantResp struct {
		Error   bool   `json:"error"`
		Status  int    `json:"Status"`
		Message string `json:"message"`
	}

	resp := &GrantResp{
		Status: -1, // 默认初始值
	}
	if err := json.Unmarshal(data, resp); err != nil {
		log.Error(err)
		return false, err
	}

	if resp.Status != success {
		log.Error(resp.Message)
		return false, errors.New(resp.Message)
	}

	return true, nil
}

// FixBinlogFormat 修正binlog 格式为 row格式 参考文档 {https://cf.jd.com/pages/viewpage.action?pageId=80085879}
func (d *DBS) FixBinlogFormat(dbsUrl, token string, host string, port int) (bool, error) {
	url := fmt.Sprintf("%s/grant/binlogformat", dbsUrl)

	type FixReq struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	req := &FixReq{
		Host: host,
		Port: port,
	}

	// {"Status":400,"Message":"need Token in Header"}, code:400, err:<nil>
	data, code, err := http.PostJSON(url, token, req, d.timeout, d.timeout)
	log.Debugf("end url:%v, data:%s, code:%d, err:%v", url, data, code, err)
	if err != nil || code != HttpOkResponse {
		return false, fmt.Errorf("%v:%s", err, string(data))
	}

	// 成功返回 0
	success := 0

	// 请求授权接口返回的结构
	type FixResp struct {
		Error   bool   `json:"error"`
		Status  int    `json:"Status"`
		Message string `json:"message"`
	}

	resp := &FixResp{
		Status: -1, // 默认初始值
	}
	if err := json.Unmarshal(data, resp); err != nil {
		log.Error(err)
		return false, err
	}

	if resp.Status != success {
		log.Error(resp.Message)
		return false, errors.New(resp.Message)
	}

	return true, nil
}
