create database IF NOT EXISTS tower;

use tower;

-- filter table
DROP TABLE IF EXISTS filter;
create table if not exists filter(
id varchar(36) COMMENT '主键 md5(${group_id}, ${type}, ${table_name})',
name varchar(128) comment '过滤器名称 用于显示',
group_id varchar(36) comment '分组id',
type enum('WHITE','BLACK') CHARACTER SET utf8 NOT NULL DEFAULT 'WHITE' COMMENT '过滤规则类型',
table_name varchar(255) not null default '.*\\..*' COMMENT '表名',
events varchar(128) not null default 'all' COMMENT 'binlog事件类型',
white_columns varchar(1024) default null COMMENT '保留列信息',
black_columns varchar(1024) default null COMMENT '过滤列信息',
fake_cols varchar(128) COMMENT '伪列信息',
business_keys varchar(128) COMMENT '业务主键',
create_time timestamp NOT NULL DEFAULT '2016-12-31 16:00:00' COMMENT '创建时间',
update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
primary key(id),
UNIQUE INDEX(group_id, type, table_name)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='过滤器';

-- mq rule table
DROP TABLE IF EXISTS mq_rule;
create table if not exists mq_rule(
id varchar(36) COMMENT '主键 md5(${topic}, ${group_id}, ${filter_id})',
name varchar(36) comment 'mq 规则名称 用于显示',
topic varchar(36) COMMENT '消息队列主题',
group_id VARCHAR(36) comment '组id',
with_trx enum('true', 'false') CHARACTER SET utf8 NOT NULL DEFAULT 'false' COMMENT '是否发送 begin/commit消息体',
producer_class varchar(254) COMMENT '生产者类类名',
paras varchar(3000) COMMENT '消息队列参数 包括用户名,密码etc.',
order_type enum('NO_ORDER', 'BUSINESS_KEY_ORDER', 'TABLE_ORDER', 'DB_ORDER', 'INSTANCE_ORDER') default 'INSTANCE_ORDER' COMMENT '消息遵循顺序',
create_time timestamp NOT NULL DEFAULT '2016-12-31 16:00:00' COMMENT '创建时间',
update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
primary key (id),
UNIQUE INDEX(topic, group_id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='mq规则';

-- mysql instance one instance only belongs to one group
DROP TABLE IF EXISTS instances;
create table if not exists instances(
id varchar(36) COMMENT '主键 md5(${host:port})',
group_id VARCHAR(36) comment '组id',
slave_id bigint(20) COMMENT 'mysql dump slave id',
host varchar(64) COMMENT 'mysql host',
port int COMMENT '端口',
zk varchar(254) COMMENT 'ZK servers 地址',
status  enum('init', 'wait', 'unauthorized', 'agree', 'oppose', 'dump') CHARACTER SET utf8 NOT NULL DEFAULT 'init' COMMENT '审批状态 同意或者驳回',
user varchar(32) COMMENT 'mysql dump 用户名',
password varchar(64) COMMENT 'mysql dump 密码',
request_id varchar(64) COMMENT '请求MySQL编号',
url VARCHAR(128) COMMENT '授权服务流程查看url',
create_time timestamp NOT NULL DEFAULT '2016-12-31 16:00:00' COMMENT '创建时间',
update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
unique(group_id, host, port),
primary key (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='MySQL 实例';

-- rule table
DROP TABLE IF EXISTS rule;
create table if not exists rule(
group_id VARCHAR(36) comment '组id',
instance_id varchar(36) COMMENT 'mysql instance id',
storage_type enum('MQ_STORAGE', 'KV_STORAGE') COMMENT '存储类型',
convert_class varchar(254) COMMENT '消息转换类类名',
rule_id varchar(36) COMMENT '规则id 不同存储类型对应不同表里的规则',
create_time timestamp NOT NULL DEFAULT '2016-12-31 16:00:00' COMMENT '创建时间',
update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
unique(instance_id, rule_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='实例规则关系表';

-- group information 
create table if not exists groups(
id varchar(36) COMMENT '主键',
name varchar(36) COMMENT '分组名',
mark varchar(64) COMMENT '备注',
create_time timestamp NOT NULL DEFAULT '2016-12-31 16:00:00' COMMENT '创建时间',
update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
unique(name),
primary key(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='分组信息表';

-- one erp only belongs to one group
create table if not exists users(
id VARCHAR(36) COMMENT '主键 md5(${erp}, ${group_id})',
erp varchar(32) COMMENT '用户erp',
email VARCHAR(128) COMMENT '用户邮箱 用做系统升级统一邮件',
org_name VARCHAR(128) COMMENT '组织机构名 用来排查组织机构详情',
group_id varchar(36) COMMENT '分组的id',
role enum('user', 'admin', 'creator') default 'user' COMMENT '用户角色 {普通用户,管理员,创建者}',
create_time timestamp NOT NULL DEFAULT '2016-12-31 16:00:00' COMMENT '创建时间',
update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
UNIQUE(erp, group_id),
primary key(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户信息表';

-- 超级管理员表
create table if not EXISTS super_admin(
erp VARCHAR(33) COMMENT '管理员erp',
create_time timestamp NOT NULL DEFAULT '2016-12-31 16:00:00' COMMENT '创建时间',
update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
primary key(erp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='超级管理员表';


insert into super_admin(erp, create_time, update_time)values('pengan3', now(), now()),('qianxi23', now(), now());

CREATE TABLE `mq_filter_relation` (
  `id` varchar(36) NOT NULL COMMENT '主键 md5(${mq_id}, ${filter_id})',
  `mq_id` varchar(36) NOT NULL COMMENT 'mq id',
  `filter_id` varchar(36) NOT NULL COMMENT 'filter id',
  `create_time` timestamp NOT NULL DEFAULT '2016-12-31 16:00:00' COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `relation` (`mq_id`,`filter_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='mq过滤器关系表';

-- config instances
DROP TABLE IF EXISTS config;
CREATE TABLE `config` (
  `keyword` varchar(64) NOT NULL COMMENT '关键字',
  `mark`  varchar(64) not null COMMENT '备注',
  `zk` varchar(255) NOT NULL COMMENT 'zk地址',
  `dns` varchar(255) NOT NULL COMMENT 'wave 集群对应域名',
  `path` varchar(64) COMMENT 'ZK 监听的根路径',
  `slave_id` bigint(64) COMMENT 'MySQL dump slave id',
  `slave_uuid` VARCHAR (36) COMMENT 'MySQL dump slave uuid',
  `dbs_api` varchar(255) NOT NULL COMMENT 'dbs 请求DBS接口地址 前缀 ',
  `dbs_token` varchar(255) NOT NULL COMMENT 'dbs 请求DBS接口地址 前缀 ',
  `mq_addr` varchar(255) NOT NULL COMMENT 'jmq 生产者发送地址',
  `create_time` timestamp NOT NULL DEFAULT '2019-12-31 16:00:00' COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`keyword`),
  UNIQUE (`zk`),
  UNIQUE (`dns`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='保存配置信息';

-- wave instances
DROP TABLE IF EXISTS waves;
create table if not EXISTS waves (
ip varchar(32) COMMENT 'wave 服务的ip',
zk varchar(254) COMMENT 'ZK servers 地址',
create_time timestamp NOT NULL DEFAULT '2016-12-31 16:00:00' COMMENT '创建时间',
update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
unique(ip, zk)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='dump 服务wave地址记录表';


-- 后续 /manage/api/destination.do 将会去除
insert into config select 'zh', '中国国内','10.10.10.10:2181', '/zk/wave', 10010, '54c7c6f0-70ad-11e9-a07d-8c1645350bc8','http://dbs.jd.com:8080', '{uuid-gen}','mq-cluster.jd.local:50088', now(), now();

