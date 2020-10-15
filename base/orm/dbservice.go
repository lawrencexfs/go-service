package orm

import (
	log "github.com/cihub/seelog"
	"github.com/spf13/viper"

	//"gopkg.in/mgo.v2"
	//_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	//_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	globalSession *gorm.DB /* = nil*/
)

func setConfig(configPath string) {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		panic("加载配置文件失败")
	}
}

//InitDB 初始化db
//todo: param: 传入数据库类型
func InitDB() bool {

	var err error
	if globalSession == nil {

		configPath := "./db_test.toml"
		setConfig(configPath)
		dbtype := viper.GetString("OrmDB.dbtype")
		addr := viper.GetString("OrmDB.Addr")
		timeout := viper.GetInt64("OrmDB.Timeout")
		if timeout == 0 {
			timeout = 1
		}

		globalSession, err = gorm.Open(dbtype, addr)

		if err != nil {
			log.Debug("connect failed,", err.Error())
			return false
			//panic(err)
		}
		log.Info("connect DB success, addr = ", addr)
	}

	return true
}
