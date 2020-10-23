package mysql

import (
	log "github.com/cihub/seelog"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

var (
	mysqlDB *sqlx.DB
	err     error
)

// 测试时这个函数再打开
func setConfig(configPath string) {
	viper.SetConfigFile(configPath)
	if err = viper.ReadInConfig(); err != nil {
		panic("加载配置文件失败")
	}
}

// InitDB 初始化db
func InitDB(configPath string) (err error) {
	if configPath == "" {
		log.Info("InitDB, configPath is nil, using default config")
		configPath = "./db_test.toml"
	}

	setConfig(configPath)
	dbtype := viper.GetString("MysqlDB.dbtype")
	addr := viper.GetString("MysqlDB.Addr")
	timeout := viper.GetInt64("MysqlDB.Timeout")
	if timeout == 0 {
		timeout = 1
	}

	// TODO：连接超时时间
	// dur := time.Duration(timeout) * time.Second
	log.Info("addr:", addr)

	mysqlDB, err = sqlx.Open(dbtype, addr)
	if err != nil {
		log.Error("connect failed,", err)
		return err
	}
	log.Info("connect DB success, addr = ", addr)

	maxIdleConn := viper.GetInt("MysqlDB.MaxIdleConn")
	log.Info("MaxIdleConn:", maxIdleConn)
	if maxIdleConn > 0 {
		mysqlDB.SetMaxIdleConns(maxIdleConn)
	}
	maxOpenConn := viper.GetInt("MysqlDB.MaxOpenConn")
	log.Info("maxOpenConn:", maxOpenConn)
	if maxOpenConn > 0 {
		mysqlDB.SetMaxOpenConns(maxOpenConn)
	}

	return nil
}

// GetMysqlDB 获取mysql 连接
func GetMysqlDB() *sqlx.DB {
	var errNew error
	if mysqlDB != nil {
		if errNew = mysqlDB.Ping(); errNew != nil {
			mysqlDB.Close()
			errNew = InitDB("")
		}
	} else {
		errNew = InitDB("")
	}

	if errNew != nil {
		log.Error("getMysqlDB failed,", errNew)
		return nil
	}
	log.Info("get db success")
	return mysqlDB
}
