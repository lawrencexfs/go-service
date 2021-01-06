package mysqlservice

import (
	"database/sql"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	utils "github.com/giant-tech/go-service/base/utility"
	// 导入 mysql 驱动包
)

// MAXTRYCOUNT 重试最大次数
const MAXTRYCOUNT = 5

// MAXFUNCCOUNT 缓存保存函数的个数
const MAXFUNCCOUNT = 50

// Mysqlfunc 数据库函数
type Mysqlfunc struct {
	f func(db *sql.DB, p *map[string]interface{}, t string) error // 操作数据库的执行的函数
	p *map[string]interface{}                                     // 函数执行的参数
	t string                                                      // 操作数据库的表名
}

// PlayerDB 玩家读数据库数据对象
type PlayerDB struct {
	RoleID uint64
	C      chan Mysqlfunc //func(db *sql.DB, parms int) error
	T      *time.Ticker
	shard  *MySQLShard
}

// NewRoleDB 创建玩家数据库对象
func NewRoleDB(roleID uint64) *PlayerDB {
	shard, err := GetShardObj(roleID)
	if err != nil || shard == nil {
		log.Info("GetShardObj failed, roleID:", roleID)
		return nil
	}

	db := &PlayerDB{
		RoleID: roleID,
		C:      make(chan Mysqlfunc, MAXFUNCCOUNT),
		T:      time.NewTicker(10 * time.Second), //几秒后删除
		shard:  shard,
	}

	go db.run()

	return db
}

// ExecSQL 加入最大执行次数队列 这里的函数参与，已经经过了转换，不需要再次EscapeBytesBackslash1转换
func (role *PlayerDB) ExecSQL(f Mysqlfunc) error {
	tryCount := 0
	var loopTimer *time.Ticker
LABEL:
	select {
	case role.C <- f:
		tryCount = 0
		if loopTimer != nil {
			loopTimer.Stop()
		}
		return nil
	default:
		log.Error("mysql buffer is full, roleID:", role.RoleID, ",tryCount:", tryCount)
		tryCount++
		if loopTimer == nil {
			loopTimer = time.NewTicker(time.Second)
		}

		// 若执行部正确，则一秒钟重试一次，若5次不成功，退出
		for {
			select {
			case <-loopTimer.C:
				log.Info("exec count:", tryCount)
				if tryCount < MAXTRYCOUNT {
					log.Error("can not into buff, tryCount:", tryCount, ",MAXTRYCOUNT:", MAXTRYCOUNT)
					goto LABEL
				}
				break
			}
		}

		if loopTimer != nil {
			log.Info("can not exec success, roleId:", role.RoleID)
			loopTimer.Stop()
			loopTimer = nil
		}
		log.Error("mysql buffer is full, reach max try count roleID:", role.RoleID, ",tryCount:", tryCount)
		return fmt.Errorf("mysql buffer is full, reach max try")
	}
}

// Run 更新数据库的协程
func (role *PlayerDB) run() {
	log.Info("player mysql gorutine start, roleId:", role.RoleID)

	defer func() {
		fmt.Println("..........role exit")
		log.Info("role exit..............")
		if err := recover(); err != nil {
			log.Error("recover stack:", string(debug.Stack()), ",playerId:", role.RoleID)
		}
	}()

	for {
		select {
		case f := <-role.C:
			tryCount := 0
		TRYLABEL:
			if err := f.f(role.shard.MysqlObj, f.p, f.t); err != nil {
				if /*mysqldriver.ErrInvalidConn == err && */ tryCount < 3 {
					tryCount++
					log.Info("mysql exec try again, tryCount:", tryCount)
					goto TRYLABEL
				}
			}
		case <-role.T.C:
			if len(role.C) == 0 {
				log.Info("mysql goroutine len c is 0 roleId:", role.RoleID)
				// role.T.Stop()
				// return
			}
		}
	}
}

// ConvertSQLString 构造SQL的语句
func (role *PlayerDB) ConvertSQLString(sqlValues *strings.Builder, v interface{}) {
	// paramType := v.(type)

	var bString bool = false
	switch v.(type) {
	case string:
		sqlValues.WriteString(" '")
		bString = true
	default:
	}

	paramStr := utils.ConvertTypeToString(v)
	if bString {
		// 若是字符串，则转换SQL中的特殊字符，防止SQL注入
		escBuf := EscapeBytesBackslash1([]byte(paramStr))
		sqlValues.WriteString(string(escBuf[:]))
	} else {
		sqlValues.WriteString(paramStr)
	}

	if bString {
		sqlValues.WriteString("' ")
	}
}

//
// InsertSQL 输入数据
func (role *PlayerDB) InsertSQL(db *sql.DB, p *map[string]interface{}, tb string) error {
	startT := time.Now()
	a := 0
	for _, value := range *p {
		// vtype := reflect.TypeOf(value)
		// log.Info("test sendInt:", key, ",value:", value, ",table name:", tb, "type:", vtype.Kind())
		// log.Info("------string:", utils.ConvertTypeToString(value))
		if value != nil {
			a++
		}
	}

	tc := time.Since(startT)                     //计算耗时
	fmt.Printf("time cost map param = %v\n", tc) //

	// if err := db.Ping(); err != nil {
	// 	log.Error("ping sql error")
	// 	return err
	// }
	startT = time.Now()
	var sqlFields strings.Builder
	var sqlValues strings.Builder
	var sqlUpdate strings.Builder
	sqlValues.WriteString("(")
	sqlFields.WriteString("(")

	valueSize := len(*p)
	idx := 0
	for key, value := range *p {

		sqlFields.WriteString("`")
		sqlFields.WriteString(key)
		sqlFields.WriteString("`")

		role.ConvertSQLString(&sqlValues, value)

		sqlUpdate.WriteString("`")
		sqlUpdate.WriteString(key)
		sqlUpdate.WriteString("`")

		sqlUpdate.WriteString(" = ")
		role.ConvertSQLString(&sqlUpdate, value)

		idx++
		if idx != valueSize {
			sqlFields.WriteString(",")
			sqlValues.WriteString(",")
			sqlUpdate.WriteString(",")
		}
	}
	sqlValues.WriteString(")")
	sqlFields.WriteString(")")

	var SQL strings.Builder
	SQL.WriteString(" INSERT INTO ")
	SQL.WriteString(tb)
	SQL.WriteString(sqlFields.String())
	SQL.WriteString(" VALUES ")
	SQL.WriteString(sqlValues.String())
	SQL.WriteString(" ON DUPLICATE KEY UPDATE ")
	SQL.WriteString(sqlUpdate.String())
	SQL.WriteString(" ;")

	tc = time.Since(startT)                //计算耗时
	fmt.Printf("time cost sql = %v\n", tc) // time cost sql = 2.55µs

	// log.Info("SQL:", SQL.String())

	/*
		startT = time.Now()

		SQLFmt := `
			INSERT INTO PLAYER (
				role_id,
				account,
				name,
				uuid,
				level
			)
			VALUES (%d, '%s', '%s', '%s', %d)
			ON DUPLICATE KEY UPDATE
				ROLEID=%d,
				ACCOUNT='%s',
				NAME='%s',
				UUID='%s',
				level=%d
		`
		fmt.Sprintf(SQLFmt, (*p)["role_id"].(uint64),
			[]byte((*p)["account"].(string)),
			[]byte((*p)["name"].(string)),
			[]byte((*p)["uuid"].(string)),
			(*p)["level"].(int),

			[]byte((*p)["account"].(string)),
			[]byte((*p)["name"].(string)),
			[]byte((*p)["uuid"].(string)),
			(*p)["level"].(int))

		tc = time.Since(startT)                 //计算耗时
		fmt.Printf("time cost sql2 = %v\n", tc) // time cost sql2 = 1.904µs

	*/

	// log.Info(SQL2)
	// log.Info("SQL:", SQL.String())

	startT = time.Now()
	_, err := db.Exec(SQL.String())
	if err != nil {
		log.Error("exec insert data into db failed. roleId:", role.RoleID, ",err:", err)
		return err
	}
	tc = time.Since(startT)                  //计算耗时
	fmt.Printf("insert cost sql = %v\n", tc) // insert time cost sql = 232.473µs

	// affRows, err := result.RowsAffected()
	// lastInsertID, err := result.LastInsertId()
	// log.Info("affectRows:", affRows, ",roleId:", lastInsertID)

	return nil
}

// // Player 玩家数据库对象
// type Player struct {
// 	ROLEID     uint64 `mysql:"role_id"`
// 	SeqID      uint64 `mysql:"seqId"`
// 	ACCOUNT    string `mysql:"account"`
// 	NAME       string `mysql:"name"`
// 	UUID       string `mysql:"uuid"`
// 	MOBILETYPE uint8  `mysql:"mobiletype"`
// 	LEVEL      uint8  `mysql:"level"`
// 	CREATEIP   string `mysql:"create_ip"`
// 	LASTIP     string `mysql:"last_ip"`
// 	EXP        uint64 `mysql:"exp"`
// 	LOGINTYPE  uint8  `mysql:"logintype"`

// 	AllItem []byte `mysql:"all_item"`
// }

// func (role *PlayerDB) selectPlayerInfo(db *sql.DB, p *map[string]interface{}, tb string) error {
// 	start, ok := (*p)["role_id_start"].(uint64)
// 	if !ok {
// 		log.Error("params is invalidate")
// 		return nil
// 	}
// 	end, ok := (*p)["role_id_end"].(uint64)
// 	if !ok {
// 		log.Error("params is invalidate")
// 		return nil
// 	}

// 	fmt.Println("start:", start, ",end:", end, ",table:", tb)

// 	// fmtSQL := fmt.Sprintf(
// 	// 	` select * from %s where role_id >= %d and role_id < %d;`, tb, start, end)

// 	// result, err := db.Exec(fmtSQL.String())
// 	// 这里的查询字段数量和类型一定要和rows.Scan中的类型和数量一致
// 	stmt, err := db.Prepare(`select role_id, seqId, account, name, uuid, mobiletype, level, create_ip, last_ip, exp,logintype,all_item
// 						 from player where role_id >= ? and role_id < ?;`)
// 	// db.QueryDataPre()
// 	if err != nil {
// 		log.Error("prepare failed")
// 	}

// 	defer stmt.Close()

// 	rows, err := stmt.Query(start, end)
// 	defer rows.Close()
// 	if err != nil {
// 		log.Error("exec select data from db failed. start:", start, ",end:", end, ",err:", err)
// 		return err
// 	}

// 	var roles = make(map[uint64]*Player, 0)

// 	for rows.Next() {
// 		var newPlayer Player
// 		// rows.Scan(&newPlayer.ROLEID, &newPlayer.SeqID, &newPlayer.ACCOUNT,
// 		// 	&newPlayer.NAME, &newPlayer.UUID, &newPlayer.MOBILETYPE, &newPlayer.LEVEL)
// 		rows.Scan(&newPlayer.ROLEID, &newPlayer.SeqID, &newPlayer.ACCOUNT, &newPlayer.NAME, &newPlayer.UUID,
// 			&newPlayer.MOBILETYPE, &newPlayer.LEVEL, &newPlayer.CREATEIP, &newPlayer.LASTIP,
// 			&newPlayer.EXP, &newPlayer.LOGINTYPE, &newPlayer.AllItem)

// 		if err != nil {
// 			fmt.Println(err.Error())
// 			continue
// 		}
// 		fmt.Println("roleId:", newPlayer.ROLEID)
// 		if newPlayer.ROLEID == uint64(0) {
// 			fmt.Println(err.Error())
// 			break
// 		}

// 		roles[newPlayer.ROLEID] = &newPlayer
// 	}

// 	err = rows.Err()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return err
// 	}

// 	fmt.Println("players len:", len(roles))

// 	return nil
// }
