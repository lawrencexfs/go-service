package mysqlservice

import (
	"database/sql"
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"

	log "github.com/cihub/seelog"
)

type Player struct {
	ROLEID     uint64 `mysql:"role_id"`
	SeqID      uint64 `mysql:"seqId"`
	ACCOUNT    string `mysql:"account"`
	NAME       string `mysql:"name"`
	UUID       string `mysql:"uuid"`
	MOBILETYPE uint8  `mysql:"mobiletype"`
	LEVEL      uint8  `mysql:"level"`
	CREATEIP   string `mysql:"create_ip"`
	LASTIP     string `mysql:"last_ip"`
	EXP        uint64 `mysql:"exp"`
	LOGINTYPE  uint8  `mysql:"logintype"`

	AllItem []byte `mysql:"all_item"`
}

func TestPlayerDBInsert(t *testing.T) {
	mysqlObj, err := InitMySQL("")
	if err != nil || mysqlObj == nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	roleID := uint64(2000)

	// shardObj, err := mysqlObj.GetShardObj(roleID)
	// if err != nil {
	// 	t.Fatalf("GetShardObj failed, roleId:%v", roleID)
	// }

	// log.Info("mysql db addr:", shardObj.addr, ",cellid:", shardObj.cellid)
	roleObj := NewRoleDB(roleID)
	// sendInt := 5

	for i := 1; i < 50; i++ {
		params := make(map[string]interface{}, 0)
		params["role_id"] = uint64(2000 + i)
		params["account"] = "aaa" + strconv.Itoa(i)
		params["name"] = "bbbb" + strconv.Itoa(i)
		params["uuid"] = "cccc" + strconv.Itoa(i)
		params["mobiletype"] = 1
		params["level"] = int(100 + i)
		params["create_ip"] = "127.0.0.1"
		params["last_ip"] = "127.0.0.1"
		params["exp"] = 500 + i
		params["logintype"] = 1

		testFunc := Mysqlfunc{
			f: roleObj.InsertSQL,
			p: &params,
			t: "player",
		}
		err := roleObj.ExecSQL(testFunc)
		if err != nil {
			log.Error("exec sql failed.")
		}
	}

	runtimes := 0
	c := time.NewTicker(time.Second)
	for {
		<-c.C
		runtimes++
		fmt.Println("timer runnintg... runtimes:", runtimes)
		fmt.Println("goroutine num:", runtime.NumGoroutine())
	}
}

func selectPlayerInfo(db *sql.DB, p *map[string]interface{}, tb string) error {
	start, ok := (*p)["role_id_start"].(uint64)
	if !ok {
		fmt.Println("params is invalidate")
		return nil
	}
	end, ok := (*p)["role_id_end"].(uint64)
	if !ok {
		fmt.Println("params is invalidate")
		return nil
	}

	fmt.Println("selectPlayerInfo start:", start, ",end:", end, ",table:", tb)

	// fmtSQL := fmt.Sprintf(
	// 	` select * from %s where role_id >= %d and role_id < %d;`, tb, start, end)

	// rows, err := db.Query(fmtSQL)
	// 这里的查询字段数量和类型一定要和rows.Scan中的类型和数量一致(或者指定查询的字段名，推荐用字段名查询)
	stmt, err := db.Prepare(`select * from player where role_id >= ? and role_id < ?;`)

	if err != nil {
		fmt.Println("======prepare failed")
		return fmt.Errorf("prepare failed")
	}

	fmt.Println("======exec=========1")

	defer func() {
		if stmt != nil {
			stmt.Close()
		}
	}()

	rows, err := stmt.Query(start, end)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	if err != nil {
		log.Error("exec select data from db failed. start:", start, ",end:", end, ",err:", err)
		return err
	}
	fmt.Println("======exec=========2")
	var roles = make(map[uint64]*Player, 0)

	for rows.Next() {
		var newPlayer Player
		rows.Scan(&newPlayer.ROLEID, &newPlayer.SeqID, &newPlayer.ACCOUNT, &newPlayer.NAME, &newPlayer.UUID, &newPlayer.MOBILETYPE, &newPlayer.LEVEL, &newPlayer.CREATEIP, &newPlayer.LASTIP, &newPlayer.EXP, &newPlayer.LOGINTYPE, &newPlayer.AllItem)

		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		if newPlayer.ROLEID == uint64(0) {
			fmt.Println(err.Error())
			return err
		}

		roles[newPlayer.ROLEID] = &newPlayer
	}
	fmt.Println("======exec=========3")
	err = rows.Err()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Println("============= players len:", len(roles))
	fmt.Println("======exec=========4")

	return nil
}

func getSelectPlayers(roleID uint64, startNum uint64, endNum uint64) {
	// roleID := uint64(id)
	shardObj, err := GetShardObj(roleID)
	if err != nil {
		fmt.Printf("GetShardObj failed, roleId:%v", roleID)
	}

	log.Info("mysql db addr:", shardObj.addr, ",cellid:", shardObj.cellid)
	roleObj := NewRoleDB(roleID)
	// sendInt := 5

	params := make(map[string]interface{}, 0)
	params["role_id_start"] = startNum
	params["role_id_end"] = endNum

	testFunc := Mysqlfunc{
		f: selectPlayerInfo,
		p: &params,
		t: "player",
	}

	err2 := roleObj.ExecSQL(testFunc)
	if err2 != nil {
		log.Error("exec sql failed.")
	}
}

// 查询数据的测试用例
func TestSelectDataFromDB(t *testing.T) {
	InitMySQL("")

	for i := 1; i < 100; i++ {
		startNum := uint64(i * 10)
		endNum := uint64((i + 1) * 10)
		/*go */ getSelectPlayers(uint64(i), startNum, endNum)
	}

	runtimes := 0
	c := time.NewTicker(time.Second)
	for {
		<-c.C
		runtimes++
		fmt.Println("timer runnintg... runtimes:", runtimes)
		fmt.Println("goroutine num:", runtime.NumGoroutine())
	}
}
