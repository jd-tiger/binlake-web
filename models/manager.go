package models

// 从manager 端请求有一个永远只有MetaData 一个对象
// 每次输入的对象是一个 MetaData{DbInfo, BinlogInfo, Counter, Candidate}
// 返回结果则是将对应的meta data 填充相应的值结果集合 或者是 {code == 1000, message}

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/zssky/log"
	"github.com/zssky/tc/http"

	"github.com/jd-tiger/binlake-web/jsmeta"
)

const (
	successCode = 1000
)

// Manager 连接请求后端的管理端服务
type Manager struct {
	urlPrefix   string
	token       string
	deadline    time.Duration
	dialTimeout time.Duration
}

// ManagerResp 管理端更新返回的结果集合
type ManagerResp struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// NewManager 创建一个 manager 结构体
func NewManager() *Manager {
	url := beego.AppConfig.String("manager::url")
	token := beego.AppConfig.String("manager:token")
	timeout, _ := beego.AppConfig.Int64("manager::timeout")
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return &Manager{
		urlPrefix:   url,
		token:       token,
		deadline:    time.Second * time.Duration(timeout),
		dialTimeout: time.Second * time.Duration(timeout),
	}
}

// NewTestManager create test manager
func NewTestManager(url, token string, timeout int64) *Manager {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}

	return &Manager{
		urlPrefix:   url,
		token:       token,
		deadline:    time.Second * time.Duration(timeout),
		dialTimeout: time.Second * time.Duration(timeout),
	}
}

// doWRequest do write request 调用 管理端写入数据
func (t *Manager) doWRequest(url string, req *jsmeta.MetaData) (*ManagerResp, error) {
	resp := ManagerResp{}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, errors.New("meta data to json error")
	}

	bts, err := jsmeta.GzipCompress(jsonData)
	if err != nil {
		return nil, errors.New("meta data to bytes error")
	}

	log.Debugf("begin url: %s, data:%v", url, req)
	data, code, err := http.PostHex(url, t.token, hex.EncodeToString(bts), t.deadline, t.dialTimeout)
	log.Debugf("end url:%v, data:%s, code:%d, err:%v", url, data, code, err)
	if err != nil {
		return &resp, err
	}

	return &resp, json.Unmarshal(data, &resp)
}

// doRRequest do read request 调用 管理端读数据接口
func (t *Manager) doRRequest(url string, req *jsmeta.MetaData) (*jsmeta.MetaData, error) {
	resp := jsmeta.MetaData{}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, errors.New("meta data to json error")
	}

	bts, err := jsmeta.GzipCompress(jsonData)
	if err != nil {
		return nil, errors.New("meta data to bytes error")
	}

	log.Debugf("begin url:%v", url)
	data, code, err := http.PostHex(url, t.token, hex.EncodeToString(bts), t.deadline, t.dialTimeout)
	log.Debugf("end url:%v, data:%s, code:%d, err:%v", url, data, code, err)
	if err != nil {
		return &resp, err
	}

	re := &ManagerResp{}
	if err := json.Unmarshal(data, re); err == nil {
		// 如果正常能够json 解析 返回正常 code message 则表示有异常
		log.Error("request for url error ", re)
		return nil, errors.New(re.Message)
	}

	// 16进制反解
	bts, err = hex.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		log.Error(err)
		return &resp, err
	}

	jsonData, err = jsmeta.GzipUnCompress(bts)
	if err != nil {
		log.Error(err)
		return &resp, err
	}

	return &resp, json.Unmarshal(jsonData, &resp)
}

// CreateZNodes 链接管理端 创建z-node
func (t *Manager) CreateZNodes(meta *jsmeta.MetaData) (*ManagerResp, error) {
	url := fmt.Sprintf("%screate/znodes/", t.urlPrefix)
	log.Debug("url ", url)
	resp, err := t.doWRequest(url, meta)
	if err != nil {
		log.Errorf("url:%v, error:%v", url, err)
		return nil, err
	}

	if resp.Code != successCode {
		log.Errorf("url:%v, resp:%v", url, resp)
		return nil, errors.New(resp.Message)
	}

	return resp, err
}

// SetBinlog 设置binlog 的位置
func (t *Manager) SetBinlog(req *jsmeta.MetaData) (*ManagerResp, error) {
	url := fmt.Sprintf("%sset/binlog/", t.urlPrefix)
	resp, err := t.doWRequest(url, req)

	log.Debug("url:", url, "resp:", resp)
	if err != nil {
		log.Error("url:", url, "error:", err)
		return nil, err
	}

	if resp.Code != successCode {
		log.Error(resp.Message)
		return nil, errors.New(resp.Message)
	}

	return resp, nil
}

// RemoveNode 重置重复次数
func (t *Manager) RemoveNode(req *jsmeta.MetaData) (*ManagerResp, error) {
	url := fmt.Sprintf("%sremove/node/", t.urlPrefix)

	resp, err := t.doWRequest(url, req)

	log.Debug("url:", url, "resp:", resp)

	if err != nil {
		log.Error("url:", url, "error:", err)
		return nil, err
	}

	if resp.Code != successCode {
		log.Error(resp.Message)
		return nil, errors.New(resp.Message)
	}

	return resp, nil
}

// ExistNode 检查节点是否存在 存在 返回true， 否则返回 false
func (t *Manager) ExistNode(req *jsmeta.MetaData) (*ManagerResp, error) {
	url := fmt.Sprintf("%sexist/node/", t.urlPrefix)

	resp, err := t.doWRequest(url, req)

	log.Debug("url:", url, "resp:", resp)

	if err != nil {
		log.Error("url:", url, "error:", err)
		return nil, err
	}

	if resp.Code != successCode {
		log.Error(resp.Message)
		return nil, errors.New(resp.Message)
	}

	return resp, nil
}

// ResetTryTimes 重置重复次数
func (t *Manager) ResetTryTimes(req *jsmeta.MetaData) (*ManagerResp, error) {
	url := fmt.Sprintf("%sreset/counter/", t.urlPrefix)

	resp, err := t.doWRequest(url, req)

	log.Debug("url:", url, "resp:", resp)

	if err != nil {
		log.Error("url:", url, "error:", err)
		return nil, err
	}

	if resp.Code != successCode {
		log.Error(resp.Message)
		return nil, errors.New(resp.Message)
	}

	return resp, nil
}

// SetOnline 设置dump节点状态为 online
func (t *Manager) SetOnline(req *jsmeta.MetaData) (*ManagerResp, error) {
	url := fmt.Sprintf("%sset/online/", t.urlPrefix)

	resp, err := t.doWRequest(url, req)

	log.Debug("url:", url, "resp:", resp)
	if err != nil {
		log.Error("url:", url, "error:", err)
		return nil, err
	}

	if resp.Code != successCode {
		log.Error(resp.Message)
		return nil, errors.New(resp.Message)
	}

	return resp, nil
}

// SetOffline 设置dump 节点状态为offline
func (t *Manager) SetOffline(req *jsmeta.MetaData) (*ManagerResp, error) {
	url := fmt.Sprintf("%sset/offline/", t.urlPrefix)

	resp, err := t.doWRequest(url, req)
	log.Debug("url:", url, "resp:", resp)
	if err != nil {
		log.Error("url:", url, "error:", err)
		return nil, err
	}

	if resp.Code != successCode {
		log.Error(resp.Message)
		return nil, errors.New(resp.Message)
	}

	return resp, nil
}

// SetLeader 设置binlog dump 节点leader
func (t *Manager) SetLeader(req *jsmeta.MetaData) (*ManagerResp, error) {
	url := fmt.Sprintf("%sset/leader/", t.urlPrefix)

	resp, err := t.doWRequest(url, req)

	log.Debug("url:", url, "resp:", resp)
	if err != nil {
		log.Error("url:", url, "error:", err)
		return nil, err
	}

	if resp.Code != successCode {
		log.Error(resp.Message)
		return nil, errors.New(resp.Message)
	}

	return resp, nil
}

// SetCandidate 设置 候选节点信息
func (t *Manager) SetCandidate(req *jsmeta.MetaData) (*ManagerResp, error) {
	url := fmt.Sprintf("%sset/candidate/", t.urlPrefix)

	resp, err := t.doWRequest(url, req)

	log.Debug("url:", url, "resp:", resp)
	if err != nil {
		log.Error("url:", url, "error:", err)
		return nil, err
	}

	if resp.Code != successCode {
		log.Error(resp.Message)
		return nil, errors.New(resp.Message)
	}

	return resp, nil
}

// SetTerminal 设置终止节点
func (t *Manager) SetTerminal(req *jsmeta.MetaData) (*ManagerResp, error) {
	url := fmt.Sprintf("%sset/terminal/", t.urlPrefix)

	resp, err := t.doWRequest(url, req)

	log.Debug("url:", url, "resp:", resp)
	if err != nil {
		log.Error("url:", url, "error:", err)
		return nil, err
	}

	if resp.Code != successCode {
		log.Error(resp.Message)
		return nil, errors.New(resp.Message)
	}

	return resp, nil
}

// GetSlaveStatus 获取订阅消费的binlog offset
func (t *Manager) GetSlaveStatus(req *jsmeta.MetaData) (*jsmeta.MetaData, error) {
	log.Debug("request ", req)
	url := fmt.Sprintf("%sslave/status/", t.urlPrefix)

	req, err := t.doRRequest(url, req)

	log.Debug("url:", url, "resp:", req)
	if err != nil {
		log.Error("url:", url, "error:", err)
		return nil, err
	}

	return req, nil
}
