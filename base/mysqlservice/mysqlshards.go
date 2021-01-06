package mysqlservice

import (
	"database/sql"
	"fmt"

	log "github.com/cihub/seelog"

	assert "github.com/arl/assertgo"
	_ "github.com/go-sql-driver/mysql" // 导入 mysql 驱动包
)

// MySQLShard 分片信息
type MySQLShard struct {
	MysqlObj *sql.DB
	addr     string
	cellid   uint64
	// roles  map[uint64]*DBRole
	// mutext sync.RWMutex
}

// DBRole role
// type DBRole struct {
// 	RoleID uint64
// 	C      chan func(db *sql.DB) error
// 	T      *time.Ticker
// 	owner  *MySQLShard
// }

func openMysql(addr string, isRead bool) (db *sql.DB, err error) {
	if isRead {
		db, err = sql.Open("mysql", fmt.Sprintf("%s&readTimeout=%s&maxAllowedPacket=%d&interpolateParams=true",
			addr, READTIMEOUT, MAXALLOWEDPACKET))
	} else {
		db, err = sql.Open("mysql", fmt.Sprintf("%s&writeTimeout=%s&maxAllowedPacket=%d&interpolateParams=true",
			addr, WRITETIMEOUT, MAXALLOWEDPACKET))
	}

	return db, err
}

func initMySQLShard(celCfg *CellCfg, isDatabaseForRead bool) (*MySQLShard, error) {
	assert.True(celCfg != nil)

	var err error
	var db *sql.DB
	if db, err = openMysql(celCfg.Addr, isDatabaseForRead); err != nil {
		log.Error(" InitMySQLShard open mysql error")
		return nil, err
	}
	db.SetMaxIdleConns(celCfg.Maxidleconn)
	db.SetMaxOpenConns(celCfg.Maxopenconn)

	log.Info("InitMySQLShard open mysql success, address:", celCfg.Addr, ",cellId:", celCfg.Cellid, ",isDatabaseForRead:", isDatabaseForRead, "Maxidleconn:", celCfg.Maxidleconn, ",Maxopenconn:", celCfg.Maxopenconn)

	if err := db.Ping(); err != nil {
		log.Error("InitMySQLShard ping mysql error")
		return nil, err
	}

	result := &MySQLShard{
		MysqlObj: db,
		addr:     celCfg.Addr,
		cellid:   celCfg.Cellid,
	}
	return result, nil
}
