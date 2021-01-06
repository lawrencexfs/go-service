package mysqlservice

import (
	"fmt"
	"sync"

	"go.uber.org/atomic"

	log "github.com/cihub/seelog"
)

// MySQL hash分片集合
type MySQL struct {
	shards map[uint64][]*MySQLShard // read 读数据库的连接列表 key: 分片单元 value: 数据库分片列表
	sm     sync.RWMutex

	shardsWrite map[uint64][]*MySQLShard // read 读数据库的连接列表 key: 分片单元 value: 数据库分片列表
}

var globalMysqlObj *MySQL

var isMysqlDBInited = atomic.NewBool(false)

// InitMySQL 初始化数据文件
func InitMySQL(cfgPath string) (*MySQL, error) {
	if isMysqlDBInited.Load() == true {
		log.Info("InitMySQL mysql had inited before.")
		return globalMysqlObj, nil
	}

	shardCfg, err := GetShardConfig(cfgPath)
	if shardCfg == nil || err != nil {
		log.Error("InitMySQL mysql conf error")
		return nil, fmt.Errorf("mysql conf error")
	}

	allCells := GetAllCellCfg(shardCfg)
	cellCount := len(allCells)
	if cellCount == 0 {
		log.Error("InitMySQL mysql conf error")
		return nil, fmt.Errorf("mysql conf error")
	}

	// 初始化 读数据库连接
	shards, err := initMysqlConn(allCells, cellCount, true)
	if err != nil {
		return nil, err
	}

	// 初始化 写数据库连接
	shardsWrite, err := initMysqlConn(allCells, cellCount, false)
	if err != nil {
		return nil, err
	}

	log.Info("init mysql database  success")

	isMysqlDBInited.Store(true)
	globalMysqlObj = &MySQL{shards: *shards, shardsWrite: *shardsWrite}
	return globalMysqlObj, nil
}

func initMysqlConn(allCells []*CellCfg, cellCount int, isDatabaseForRead bool) (mysqlshards *map[uint64][]*MySQLShard, err error) {
	shards := make(map[uint64][]*MySQLShard, cellCount)

	connNum := 0
	for i := 0; i < cellCount; i++ {
		var err error
		cell := allCells[i]

		// 读取配置文件中 读写数据库的分片连接数量
		if isDatabaseForRead {
			connNum = cell.Connshardread
		} else {
			connNum = cell.Connshardwrite
		}
		log.Info("Init mysql connection isDatabaseForRead=", isDatabaseForRead, " ,cell count=", cellCount, ", start to init mysql ,connNum:", connNum)

		shards[cell.Cellid] = make([]*MySQLShard, connNum)
		for i := 0; i < connNum; i++ {
			shards[cell.Cellid][i], err = initMySQLShard(cell, isDatabaseForRead)
			if err != nil {
				log.Error("initMysqlConn initMySQLShard err:", err)
				return nil, err
			}
		}
	}

	return &shards, nil
}

// GetShardObj 根据entityid获取mysql分片连接数据
func GetShardObj(entityID uint64) (*MySQLShard, error) {
	if isMysqlDBInited.Load() == false {
		return nil, fmt.Errorf("get db failed, isMysqlDBInited is false, entityID:%d", entityID)
	}
	globalMysqlObj.sm.RLock()
	defer globalMysqlObj.sm.RUnlock()

	cell, err := GetCellCfgByUID(entityID)
	if err != nil {
		log.Error("Get log error, entityID:", entityID)
		return nil, fmt.Errorf("get db cellid failed, entityID:%d", entityID)
	}

	shardobjs := globalMysqlObj.shards[cell.Cellid]
	if shardobjs == nil {
		log.Error("GetShardObj shardobj, entityID:", entityID)
		return nil, fmt.Errorf("GetShardObj get db shard failed, entityID:%d", entityID)
	}

	idx := entityID % uint64(cell.Connshardread)

	shard := shardobjs[idx]
	if shard == nil {
		log.Error("GetShardObj shardobj, entityID:", entityID)
		return nil, fmt.Errorf("get db shard failed, entityID:%d", entityID)
	}

	log.Info("GetShardObj entityID:", entityID, ",idx:", idx, ",Connshardread:", cell.Connshardread, ",cellId:", cell.Cellid)
	return shard, nil
}

// GetWriteShardObj 获取写数据库分片连接数据
func GetWriteShardObj(entityID uint64) (*MySQLShard, error) {
	globalMysqlObj.sm.RLock()
	defer globalMysqlObj.sm.RUnlock()

	cell, err := GetCellCfgByUID(entityID)
	if err != nil {
		log.Error("GetWriteShardObj Get log error, entityID:", entityID)
		return nil, fmt.Errorf("GetWriteShardObj get db cellid failed, entityID:%d", entityID)
	}

	shardobjs := globalMysqlObj.shardsWrite[cell.Cellid]
	if shardobjs == nil {
		log.Error("GetWriteShardObj shardobj, entityID:", entityID)
		return nil, fmt.Errorf("GetWriteShardObj get db shard failed, entityID:%d", entityID)
	}

	idx := entityID % uint64(cell.Connshardwrite)

	shard := shardobjs[idx]
	if shard == nil {
		log.Error("GetWriteShardObj shardobj, entityID:", entityID)
		return nil, fmt.Errorf("GetWriteShardObj get db shard failed, entityID:%d", entityID)
	}

	log.Info("GetWriteShardObj roleId:", entityID, ",idx:", idx, ",Connshardwrite:", cell.Connshardwrite, ",cellId:", cell.Cellid)
	return shard, nil
}

// ReInitMysql 重新加载分片配置
func ReInitMysql(cfgPath string) error {
	globalMysqlObj.sm.Lock()
	defer func() {
		log.Debug("end ReInitMysql")
		globalMysqlObj.sm.Unlock()
	}()

	if isMysqlDBInited.Load() == false {
		return fmt.Errorf("mysql had not inited before.")
	}
	log.Debug("ReInitMysql success")

	shardCfg, err := ReloadShardConfig(cfgPath)
	if err != nil {
		log.Error("mysql conf error")
		return fmt.Errorf("mysql conf error")
	}

	allCells := GetAllCellCfg(shardCfg)
	shardCount := len(allCells)
	if shardCount == 0 {
		log.Error("mysql conf error")
		return fmt.Errorf("mysql conf error")
	}

	// 初始化 读数据库连接
	shards, err := initMysqlConn(allCells, shardCount, true)
	if err != nil {
		return err
	}

	// 初始化 写数据库连接
	shardsWrite, err := initMysqlConn(allCells, shardCount, false)
	if err != nil {
		return err
	}

	// 设置全局配置文件为新的配置文件
	SetGlobalShardConfig(shardCfg)

	log.Info("ReInitMysql init mysql database  success")

	globalMysqlObj.shards = *shards
	globalMysqlObj.shardsWrite = *shardsWrite

	return nil
}

// GetShardReadCount :
func GetShardReadCount() int {
	globalMysqlObj.sm.RLock()
	defer func() {
		fmt.Println("end GetShardCount")
		globalMysqlObj.sm.RUnlock()
	}()

	return len(globalMysqlObj.shards)
}
