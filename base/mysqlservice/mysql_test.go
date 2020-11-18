package mysqlservice

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	log "github.com/cihub/seelog"

	"github.com/giant-tech/go-service/base/utility"
)

func TestMysqlShard(t *testing.T) {
	_, err := InitMySQL("")
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}
}

func TestSelectFromDB(t *testing.T) {

	_, err := InitMySQL("")
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	var ID uint64
	ID = 12
	shard, err := GetShardObj(ID)
	if err != nil {
		fmt.Printf("GetShardObj failed, Id:%v", ID)
	}
	stmt, err := shard.MysqlObj.Prepare(`select * from player where role_id >= ? and role_id < ?;`)

	if err != nil {
		fmt.Println("======prepare failed")
	}

	defer func() {
		if stmt != nil {
			stmt.Close()
		}
	}()

	rows, err := stmt.Query(10, 20)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	if err != nil {
		log.Error("select data from db failed. ID:", ID, ",err:", err)
	}
	var roles = make(map[uint64]*Player, 0)

	for rows.Next() {

		//反射出未导出字段
		val := reflect.ValueOf(rows).Elem().FieldByName("lastcols")

		log.Debug("v = ", val)
		fmt.Println("v: ", val, ", v.kind()", val.Kind())

		/*for i := range v {

			fmt.Println("i = ", i)
		}
		*/

		//与上面的区别是：这个是可寻址的
		val = reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem()
		if reflect.Slice == val.Kind() {
			if byteSlice, ok := val.Interface().([]byte); ok {
				fmt.Println("byteSlice = ", byteSlice)
			} else {
				len := val.Len()

				for i := 0; i < len; i++ {
					/*if i < 2 {
						fmt.Println("val.Index(i) = ", val.Index(i), ", kind=", val.Index(i).Kind(), ", interface=", val.Index(i).Interface().(int64))
					} else {
						fmt.Println("val.Index(i) = ", val.Index(i), ", kind=", val.Index(i).Kind(), ", interface=", val.Index(i).Interface().([]uint8))
					}*/
					fmt.Println(utility.ConvertReflectVal(val.Index(i).Interface()))
				}
			}
		}

		var newPlayer Player
		rows.Scan(&newPlayer.ROLEID, &newPlayer.SeqID, &newPlayer.ACCOUNT, &newPlayer.NAME, &newPlayer.UUID, &newPlayer.MOBILETYPE, &newPlayer.LEVEL, &newPlayer.CREATEIP, &newPlayer.LASTIP, &newPlayer.EXP, &newPlayer.LOGINTYPE, &newPlayer.AllItem)

		if err != nil {
			fmt.Println(err.Error())
		}

		if newPlayer.ROLEID == uint64(0) {
			fmt.Println(err.Error())
		}

		roles[newPlayer.ROLEID] = &newPlayer
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestSelectFromDB_reflect(t *testing.T) {

	_, err := InitMySQL("")
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	var ID uint64
	ID = 12
	shard, err := GetShardObj(ID)
	if err != nil {
		fmt.Printf("GetShardObj failed, Id:%v", ID)
	}
	stmt, err := shard.MysqlObj.Prepare(`select * from player where role_id >= ? and role_id < ?;`)

	if err != nil {
		fmt.Println("======prepare failed")
	}

	defer func() {
		if stmt != nil {
			stmt.Close()
		}
	}()

	rows, err := stmt.Query(10, 20)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	if err != nil {
		log.Error("select data from db failed. ID:", ID, ",err:", err)
	}

	if rows.Next() {

		//反射出未导出字段
		val := reflect.ValueOf(rows).Elem().FieldByName("lastcols")

		log.Debug("v = ", val)
		fmt.Println("v: ", val, ", v.kind()", val.Kind())

		//与上面的区别是：这个是可寻址的
		val = reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem()
		if reflect.Slice == val.Kind() {
			if byteSlice, ok := val.Interface().([]byte); ok {
				fmt.Println("byteSlice = ", byteSlice)
			} else {
				len := val.Len()

				for i := 0; i < len; i++ {
					/*if i < 2 {
						fmt.Println("val.Index(i) = ", val.Index(i), ", kind=", val.Index(i).Kind(), ", interface=", val.Index(i).Interface().(int64))
					} else {
						fmt.Println("val.Index(i) = ", val.Index(i), ", kind=", val.Index(i).Kind(), ", interface=", val.Index(i).Interface().([]uint8))
					}*/
					fmt.Println(utility.ConvertReflectVal(val.Index(i).Interface()))
				}
			}
		}

	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err.Error())
	}
}
