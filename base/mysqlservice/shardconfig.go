package mysqlservice

import (
	"sort"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/spf13/viper"
)

// https://mholt.github.io/json-to-go/

// CellCfg 分片单元
type CellCfg struct {
	Cellid         uint64 `json:"cellid"`
	Name           string `json:"name"`
	Addr           string `json:"addr"`
	Startid        uint64 `json:"startid"`
	Endid          uint64 `json:"endid"`
	Maxidleconn    int    `json:"maxidleconn"`
	Maxopenconn    int    `json:"maxopenconn"`
	Connshardread  int    `json:"connshardread"`
	Connshardwrite int    `json:"connshardwrite"`
}

// ShardCfg hash分片结构
type ShardCfg struct {
	Shardid uint64    `json:"shardid"`
	Name    string    `json:"name"`
	Hashval uint64    `json:"hashval"`
	Cells   []CellCfg `json:"cells"`
	Minid   uint64    `json:"-"` // 不读去配置文件字段，自动算出的值
	Maxid   uint64    `json:"-"` // 不读去配置文件字段，自动算出的值
}

// GroupCfg 配置组结构
type GroupCfg struct {
	Groupid uint64 `json:"groupid"`
	Version uint32 `json:"version"`
	Name    string `json:"name"`
	// Binsert  uint32              `json:"binsert"`
	Factor   uint64               `json:"factor"`
	Shards   []ShardCfg           `json:"shards"`
	Minid    uint64               `json:"-"`
	Maxid    uint64               `json:"-"`
	ShardMap map[uint64]*ShardCfg `json:"-"` // key: Factor的余数
	// ShardCellMap map[uint64]*CellCfg  `json:"-"` // key: cellId
}

// ShardConfig 分片的go结构体
type ShardConfig struct {
	Groups []GroupCfg `json:"groups"`
}

var (
	gShardCfg  *ShardConfig
	gShardCell sync.Map
	// sm         sync.RWMutex
)

// 测试时这个函数再打开
func setConfig(configPath string) {
	viper.SetConfigFile(configPath)
	//viper.SetConfigType("json")
	if err := viper.ReadInConfig(); err != nil {
		panic("加载配置文件失败")
	}
}

func convertStructCfg(shardCfg *ShardConfig) (err error) {
	err = viper.Unmarshal(shardCfg)
	if err != nil {
		panic("解析结构体失败")
	}

	return nil
}

// initShardConfig 数据库分片的配置
func initShardConfig(configPath string) (cfg *ShardConfig, err error) {
	if configPath == "" {
		// log.Info("InitShardConfig, configPath is nil, using default config")
		configPath = "./mysqlShards.json"
	}

	setConfig(configPath)

	sharCfg := &ShardConfig{}
	err = convertStructCfg(sharCfg)
	if err != nil {
		// log.Error("InitShardConfig, convertStructCfg failed, configPath:", configPath)
		return nil, err
	}

	// log.Info("shardCfg:", gShardCfg)
	return sharCfg, nil
}

func checkShardValidate(shdCfg *ShardCfg) bool {
	var min uint64
	var max uint64
	bGetMin := false
	bGetMax := false

	var sharIDMap = make(map[uint64]int, 0)
	for _, cell := range shdCfg.Cells {
		if !bGetMin {
			bGetMin = true
			min = cell.Startid
		} else {
			if min >= cell.Startid {
				min = cell.Startid
			}
		}

		if !bGetMax {
			bGetMax = true
			max = cell.Endid
		} else {
			if max <= cell.Endid {
				max = cell.Endid
			}
		}

		if _, ok := sharIDMap[cell.Cellid]; ok {
			log.Debug("shard cell id config is invalidate cellId:", cell.Cellid)
			return false
		}
		sharIDMap[cell.Cellid] = 1
		// log.Info("cellId:", cell.Cellid)
		// 检查startid 和endid的值
		if cell.Endid <= cell.Startid {
			log.Debug("shard cell Startid and Endid config is invalidate ")
			return false
		}

		if cell.Connshardread <= 0 {
			log.Debug("shard cell connshardread is invalidate ")
			return false
		}

		if cell.Connshardwrite <= 0 {
			log.Debug("shard cell connshardwrite is invalidate ")
			return false
		}

		// log.Info("check startid and endid success, startid:", cell.Startid, ",endid:", cell.Endid, ",min:", min, ",max:", max)
	}

	// log.Info("check startid and endid success,min:", min, ",max:", max)
	// 检查值区间是否连续
	cellSize := len(shdCfg.Cells)
	if cellSize > 1 {
		for i := 0; i < cellSize-1; i++ {
			if shdCfg.Cells[i].Endid != shdCfg.Cells[i+1].Startid {
				log.Debug("shard cell Startid and Endid config does not keep increasing. Endid:",
					shdCfg.Cells[i].Endid, ",next startId:", shdCfg.Cells[i+1].Startid)
				return false
			}
		}

		// log.Info("check cell range success.")
	}

	shdCfg.Minid = min
	shdCfg.Maxid = max

	// log.Info("check cell range success. min:", min, ",max:", max)

	return true
}

func checkGroupValidate(gpcfg *GroupCfg) bool {
	// 第一步:检查版本号
	if gpcfg.Version != 200000 {
		log.Debug("gpcfg version is not 200000")
		return false
	}
	// log.Debug("check version success")

	// 检查factor的分片组是否都配置了
	shardsLen := len(gpcfg.Shards)
	if uint64(shardsLen) != gpcfg.Factor {
		log.Debug("Factor is not the same with shardsLen ")
		return false
	}
	// log.Debug("Shards lenght is the same with shardsLen success.")
	var exist = false
	for i := uint64(0); i < gpcfg.Factor; i++ {
		exist = false
		for _, shard := range gpcfg.Shards {
			if i == shard.Hashval {
				exist = true
			}
		}

		if exist == false {
			log.Debug("shard Factor config is invalidate ")
			return false
		}
	}
	// log.Debug("check hasval success")

	// 检查分片组id是否重复
	gpcfg.ShardMap = make(map[uint64]*ShardCfg, 0)
	// gpcfg.ShardCellMap = make(map[uint64]*CellCfg, 0)

	var sharIDMap = make(map[uint64]int, 0)

	bGetMin := false
	bGetMax := false

	for i := 0; i < len(gpcfg.Shards); i++ {
		if _, ok := sharIDMap[gpcfg.Shards[i].Shardid]; ok {
			log.Debug("shard Factor config is invalidate ")
			return false
		}

		sharIDMap[gpcfg.Shards[i].Shardid] = 1
		ret := checkShardValidate(&gpcfg.Shards[i])
		if !ret {
			log.Debug("checkShardValidate failed ")
			return false
		}

		gpcfg.ShardMap[gpcfg.Shards[i].Hashval] = &gpcfg.Shards[i]

		if !bGetMin {
			bGetMin = true
			gpcfg.Minid = gpcfg.Shards[i].Minid
		} else {
			if gpcfg.Minid > gpcfg.Shards[i].Minid {
				gpcfg.Minid = gpcfg.Shards[i].Minid
			}
		}

		if !bGetMax {
			bGetMax = true
			gpcfg.Maxid = gpcfg.Shards[i].Maxid
		} else {
			if gpcfg.Maxid < gpcfg.Shards[i].Maxid {
				gpcfg.Maxid = gpcfg.Shards[i].Maxid
			}
		}
	}

	// log.Info("====group cfg maxId:", gpcfg.Maxid, ",minId:", gpcfg.Minid)
	// log.Debug("====group cfg maxId:", gpcfg.Maxid, ",minId:", gpcfg.Minid)
	return true
}

func checkFactorValidate(factors []int) bool {
	sizeFactor := len(factors)
	if sizeFactor <= 1 {
		// log.Error("factor is invalidate, key len is invalidate")
		// 若factor 是 1, 直接返回正确
		if sizeFactor == 1 {
			if factors[0] == 1 {
				return true
			}
		}
		return false
	}

	for i := 0; i < sizeFactor-1; i++ {
		if factors[i+1] != factors[i]*2 {
			// log.Error("factor is invalidate, key factors is invalidate")
			return false
		}
	}

	return true
}

// CheckConfigValidate 检查配置文件是否正确
func CheckConfigValidate(cfg *ShardConfig) bool {
	// log.Info("start to check group validate")
	// log.Debug("start to check group validate")
	if cfg == nil {
		log.Debug("CheckConfigValidate failed. cfg is nil")
		return false
	}
	// 第一步：检查groupid是否重复
	var shardGrops = make(map[uint64]interface{}, 0)
	var shardFactor = make(map[uint64]int, 0)
	for i := 0; i < len(cfg.Groups); i++ {
		cfgG := &cfg.Groups[i]
		if _, ok := shardGrops[cfgG.Groupid]; ok {
			log.Debug("数据库配置文件错误,groupid 重复")
			return false
		}

		// log.Debug("start to check group validate")
		retCode := checkGroupValidate(cfgG)
		if !retCode {
			log.Debug("数据库配置文件错误, checkGroupValidate failed")
			return false
		}

		shardGrops[cfgG.Groupid] = cfgG

		if _, ok := shardFactor[cfgG.Factor]; !ok {
			shardFactor[cfgG.Factor] = 1
		} else {
			shardFactor[cfgG.Factor]++
		}
	}

	// 检查 factor 数值
	sizeFactor := len(shardFactor)
	if sizeFactor < 1 {
		log.Debug("数据库配置文件错误, shardFactor count is 0")
		return false
	}

	keys := make([]int, 0)
	for key, num := range shardFactor {
		if num != 1 {
			log.Debug("数据库配置文件错误, factor count is invalidate, key:%d, num:%d", key, num)
			return false
		}

		keys = append(keys, int(key))
	}
	// log.Info("shard factor keys:", keys)
	sort.Ints(keys)
	// log.Info("after shard factor keys:", keys)

	ok := checkFactorValidate(keys)
	if !ok {
		log.Debug("数据库配置文件错误, checkFactorValidate failed")
		return false
	}
	// log.Debug("CheckConfigValidate success")
	return true
}

// GetShardConfig 第一次初始化获取分片
func GetShardConfig(configPath string) (cfg *ShardConfig, err error) {
	if gShardCfg == nil {
		// sharCfg := &ShardConfig{}
		sharCfg, err := initShardConfig(configPath)
		if err != nil {
			// log.Error("shard config is nil")
			return nil, log.Error("Init shard config failed")
		}

		// log.Debug("===============initShardConfig end")
		ret := CheckConfigValidate(sharCfg)
		if !ret {
			// log.Error("CheckConfigValidate failed")
			return nil, log.Error("CheckConfigValidate failed")
		}

		gShardCfg = sharCfg
	}

	return gShardCfg, nil
}

// ReloadShardConfig 重载配置文件
func ReloadShardConfig(configPath string) (cfg *ShardConfig, err error) {
	if configPath == "" {
		return nil, log.Error("configPath failed")
	}
	// sharCfg := &ShardConfig{}
	sharCfg, err := initShardConfig(configPath)
	if err != nil {
		return nil, log.Error("Init shard config failed")
	}

	ret := CheckConfigValidate(cfg)
	if !ret {
		// log.Error("CheckConfigValidate failed")
		return nil, log.Error("CheckConfigValidate failed")
	}

	// gShardCfg = scfg

	return sharCfg, nil
}

// SetGlobalShardConfig 设置全局变量
func SetGlobalShardConfig(cfg *ShardConfig) error {
	if cfg == nil {
		return log.Error("SetGlobalShardConfig failed, cfg is nil")
	}

	gShardCfg = cfg

	return nil
}

// GetCellCfgByUID 通过ID获取分片的单元配置
func GetCellCfgByUID(id uint64) (cell *CellCfg, err error) {
	// log.Debug("GetCellCfgByUID")
	for _, cfgGrup := range gShardCfg.Groups {
		// log.Debug("cfgGrup", cfgGrup.Minid, cfgGrup.Maxid)
		if id >= cfgGrup.Minid && id < cfgGrup.Maxid {
			ivalue := id % cfgGrup.Factor
			shard, ok := cfgGrup.ShardMap[ivalue]
			if !ok {
				// log.Errorf("id is not validate id:", id)
				return nil, log.Error("id is not validate, id:%d", id)
			}

			for _, cell := range shard.Cells {
				if id >= cell.Startid && id < cell.Endid {
					return &cell, nil
				}
			}
		}
	}
	// log.Debug("id is not validate id:", id)
	return nil, log.Error("cell is not exist, id:%d", id)
}

// GetAllCellCfg 获取所有的分片数据的配置
func GetAllCellCfg(cfg *ShardConfig) []*CellCfg {
	var scfg *ShardConfig
	if cfg == nil {
		scfg = gShardCfg
	} else {
		scfg = cfg
	}

	cellCfgs := make([]*CellCfg, 0)
	for i := 0; i < len(gShardCfg.Groups); i++ {
		group := &scfg.Groups[i]

		for j := 0; j < len(group.Shards); j++ {
			shard := &group.Shards[j]

			for k := 0; k < len(shard.Cells); k++ {
				cell := &(shard.Cells[k])

				//log.Debug("==get==cellId:", cell.Cellid)
				cellCfgs = append(cellCfgs, cell)
			}
		}
	}

	return cellCfgs
}
