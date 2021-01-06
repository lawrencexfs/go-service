/*
*
 */

package mysqlservice

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	// 导入 mysql 驱动包
)

type bgSQL struct {
	sqls   []string // sql集合
	sqlLen int      // sql的长度
}

// PlayerWriteDB 写数据库玩家对象
type PlayerWriteDB struct {
	RoleID uint64

	sqlbuf    bytes.Buffer // 一次发送，不超过 MAXALLOWEDPACKET 大小
	sqlbufLen int          // 当前 sqlbuf 实际大小
	mutex     sync.Mutex
	c         chan int

	shard *MySQLShard // 数据库分片的连接信息

	SQLs       []*bgSQL // 玩家要执行的SQL集合
	bgSQLIndex int      // 当前背景存储SQL的idx记录
	bufIsFull  bool
}

// NewWriteRoleDB 创建玩家数据库对象
func NewWriteRoleDB(roleID uint64) *PlayerWriteDB {
	shard, err := GetShardObj(roleID)
	if err != nil || shard == nil {
		log.Info("GetShardObj failed, roleID:", roleID)
		return nil
	}

	playerWriteDb := &PlayerWriteDB{
		RoleID:    roleID,
		c:         make(chan int, 1),
		shard:     shard,
		bufIsFull: true,
	}

	playerWriteDb.sqlbuf.Grow(MAXALLOWEDPACKET)
	playerWriteDb.SQLs = make([]*bgSQL, 2)

	playerWriteDb.SQLs[0] = &bgSQL{}
	playerWriteDb.SQLs[0].sqls = make([]string, 0)
	playerWriteDb.SQLs[0].sqlLen = 0

	playerWriteDb.SQLs[1] = &bgSQL{}
	playerWriteDb.SQLs[1].sqls = make([]string, 0)
	playerWriteDb.SQLs[1].sqlLen = 0

	go playerWriteDb.run()

	return playerWriteDb
}

// ExecSQL 执行SQL语句 注意：拼sql的时候若参数是字符串，都要经过EscapeBytesBackslash1函数的转换，防止sql注入。
func (pwb *PlayerWriteDB) ExecSQL(sql string) error {
	sqlsize := len(sql) + 1
	if sqlsize >= MAXALLOWEDPACKET {
		log.Error("sql is too big sql len", len(sql))
		return fmt.Errorf("sql is to big")
	}

	pwb.mutex.Lock()
	defer pwb.mutex.Unlock()

	// 若buf 已经满了，
	if !pwb.bufIsFull {
		// log.Error("sql buf is full")
		return fmt.Errorf("sql buf is full")
	}

	bgSQLIndex := pwb.bgSQLIndex % 2

	// 限制下大小，以防止内存撑爆
	if pwb.SQLs[bgSQLIndex].sqlLen+sqlsize > MAXALLOWEDPACKET*MAXPACKETNUM {
		log.Error("---------------------myql buff is full")
		return fmt.Errorf("myql buff is full")
	}

	pwb.SQLs[bgSQLIndex].sqls = append(pwb.SQLs[bgSQLIndex].sqls, sql)
	pwb.SQLs[bgSQLIndex].sqlLen += sqlsize

	select {
	case pwb.c <- 1:
	default:
	}

	return nil
}

//////////////////////////////////////////// 以下内部函数调用 ///////////////
func (pwb *PlayerWriteDB) run() {
	timeout, _ := time.ParseDuration(WRITETIMEOUT)
	timeoutMs := int64(timeout) / int64(time.Millisecond)
	for {
		pwb.execRunSQL(timeoutMs)
	}
}

func (pwb *PlayerWriteDB) execRunSQL(timeoutMs int64) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("recover error:", err, "stack:", string(debug.Stack()))
			fmt.Printf("[error] %v %s", err, string(debug.Stack()))
		}
	}()

	select {
	case <-pwb.c:
		// 1. 构造 sql 文件
		pwb.mutex.Lock()
		workIndex := pwb.bgSQLIndex % 2
		pwb.bgSQLIndex++
		pwb.mutex.Unlock()

		for _, k := range pwb.SQLs[workIndex].sqls {
		LABEL2:
			if pwb.pushSQLToBuf(k) == false {
				// 若buf已满，则不接受加入的SQL
				if pwb.bufIsFull {
					pwb.mutex.Lock()
					pwb.bufIsFull = false
					pwb.mutex.Unlock()
				}

				pwb.execOnce(timeoutMs)
				goto LABEL2
			}
		}

		if pwb.sqlbufLen != 0 {
			pwb.execOnce(timeoutMs)
		}

		pwb.SQLs[workIndex] = &bgSQL{}
		pwb.SQLs[workIndex].sqls = make([]string, 0)
		pwb.SQLs[workIndex].sqlLen = 0
	}
}

func (pwb *PlayerWriteDB) pushSQLToBuf(sql string) bool {
	// 检查是否超过最大包了，超过压不进去，返回 false
	if pwb.sqlbufLen+len(sql)+1 > MAXALLOWEDPACKET {
		log.Error("sql is more than MAXALLOWEDPACKET size.")
		return false
	}
	pwb.sqlbuf.WriteString(sql)
	pwb.sqlbuf.WriteString(";")
	pwb.sqlbufLen += len(sql) + 1
	return true
}

func (pwb *PlayerWriteDB) execOnce(timeoutMs int64) {
	defer func() {
		pwb.sqlbuf.Reset()
		pwb.sqlbufLen = 0

		// 若buf是已满状态
		if !pwb.bufIsFull {
			pwb.mutex.Lock()
			pwb.bufIsFull = true
			pwb.mutex.Unlock()
		}
	}()

	sql := pwb.sqlbuf.String()
	t1 := time.Now().UnixNano() / 1e6
	tryCount := 0

LABEL:
	if _, err := pwb.shard.MysqlObj.Exec(sql); err != nil {
		log.Error("MySQL batch exec fail, err:", err)
		if /*err == mysqldriver.ErrInvalidConn && */ tryCount < 2 {
			log.Infof("conn invalidate mysql exec try again, tryCount:%d", tryCount)
			tryCount++
			goto LABEL
		} else {
			log.Error("exec sql failed, sql:", sql)
		}
	}
	t2 := time.Now().UnixNano() / 1e6
	fmt.Println("cost:", (t2 - t1))
	if (t2 - t1) >= timeoutMs {
		log.Infof("sql size:%d", len(sql), "MySQL batch exec takes %d ms", t2-t1)
	}
}
