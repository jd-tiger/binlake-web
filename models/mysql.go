package models

import (
	"fmt"
	"strconv"
	"strings"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/jsmeta"
)

/**
+------------------+----------+--------------+------------------+---------------------------------------------+
| File             | Position | Binlog_Do_DB | Binlog_Ignore_DB | Executed_Gtid_Set                           |
+------------------+----------+--------------+------------------+---------------------------------------------+
| mysql-bin.000029 |      191 |              |                  | 154bb9d3-2beb-11e8-bbe8-507b9d578e91:1-2872 |
+------------------+----------+--------------+------------------+---------------------------------------------+
*/

// GetMasterStatus 获取当前binlog 位置
func GetMasterStatus(user, password, host string, port int) (*jsmeta.BinlogInfo, error) {
	// 生成数据源
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		user, password, host, port, "mysql", "utf8")
	log.Debug("url ", url)

	db, err := sql.Open("mysql", url)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer db.Close()

	// 判断是否 打开gtid
	hasGTID := hasOpenGTID(db)
	log.Debug("gtid is open ", hasGTID)

	return getBinlogOffset(db, hasGTID)
}

// getBinlogOffset 获取binlog 当前的位置
func getBinlogOffset(db *sql.DB, hasGTID bool) (*jsmeta.BinlogInfo, error) {
	// 执行 master status 查询
	rows, err := db.Query(`show master status`)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()

	var file = ""
	var position = ""
	var logDb = ""
	var ignoreDb = ""
	var gtids = ""

	slave := &jsmeta.BinlogInfo{
		WithGTID: hasGTID, // 设置是否开启gtid dump
	}

	for rows.Next() {
		if len(cols) == 5 {
			// 如果有gtid 则需要获取gtid
			err := rows.Scan(&file, &position, &logDb, &ignoreDb, &gtids)
			if err != nil {
				return nil, err
			}

			slave.BinlogFile = file
			pos, _ := strconv.Atoi(position)
			slave.BinlogPos = int64(pos)
			slave.ExecutedGtidSets = gtids
		} else {
			err := rows.Scan(&file, &position, &logDb, &ignoreDb)
			if err != nil {
				return nil, err
			}

			slave.BinlogFile = file
			pos, _ := strconv.Atoi(position)
			slave.BinlogPos = int64(pos)
		}
	}
	return slave, nil
}

// GetBinlogFiles show binary logs 倒序
func GetBinlogFiles(user, password, host string, port int) ([]string, error) {
	// 生成数据源
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		user, password, host, port, "mysql", "utf8")
	log.Debug("url ", url)

	db, err := sql.Open("mysql", url)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer db.Close()

	return showBinaryLogs(db)
}

// showBinaryLogs 执行sql show binary logs 倒序
func showBinaryLogs(db *sql.DB) ([]string, error) {
	// 执行 master status 查询
	rows, err := db.Query(`show binary logs`)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rows.Close()

	var bfs []string

	var file = ""
	var size = 0
	for rows.Next() {
		// 获取当前的binlog文件 以及binlog 位置
		err := rows.Scan(&file, &size)
		if err != nil {
			return nil, err
		}
		bfs = append(bfs, fmt.Sprintf("%s:%d", file, size))
	}

	log.Debug("binary logs files ", bfs)
	return reverse(bfs), nil
}

// reverse 倒转数组 最新的binlog 在最后
func reverse(para []string) []string {
	for i, j := 0, len(para)-1; i < j; i, j = i+1, j-1 {
		para[i], para[j] = para[j], para[i]
	}
	return para
}

// hasOpenGTID 判断是否打开 gtid
func hasOpenGTID(db *sql.DB) bool {
	// 查看是否打开gtid
	rows, err := db.Query(`show variables like 'gtid_mode'`)
	if err != nil {
		log.Error(err)
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var name = ""
		var value = ""
		rows.Scan(&name, &value)
		if strings.ToUpper(value) == "ON" {
			return true
		}
	}

	return false
}
