package entity

import (
	"fmt"
	"reflect"
	"unsafe"

	mongodbservice "github.com/giant-tech/go-service/base/mongodbservice"
	mysqlservice "github.com/giant-tech/go-service/base/mysqlservice"
	"github.com/giant-tech/go-service/base/utility"

	"database/sql"

	log "github.com/cihub/seelog"
	"github.com/globalsign/mgo/bson"
)

// DBType dbtype
var DBType string

// PropDBName PropDBName
var PropDBName string

// PropTableName PropTableName
var PropTableName string

// loadFromMongoDB 从数据库中恢复
func (e *Entity) loadFromMongoDB() {
	if e.GetEntityID() == 0 {
		return
	}

	selectProps := bson.M{}
	for k, v := range e.props {
		if v.def.Persistence {
			selectProps[k] = 1
		}
	}

	retRawData := bson.Raw{}
	retM := bson.M{}
	var tempDBElems []bson.RawDocElem

	log.Debug("loadFromDB MongoDBQueryOneWithSelect: ", " , PropDBName: ", PropDBName, " , PropTableName: ", PropTableName, " , selectProps", selectProps)

	mongodbservice.MongoDBQueryOneWithSelect(PropDBName, PropTableName, bson.M{"dbid": e.GetEntityID()}, selectProps, &retRawData)

	bson.Unmarshal(retRawData.Data, &retM)
	bson.Unmarshal(retRawData.Data, &tempDBElems)

	//log.Info("select props: ", selectProps, "props return: ", ret)

	for k, v := range retM {
		if k == "_id" {
			continue
		}

		info, ok := e.props[k]
		if ok {
			log.Debug("loadFromDB, key is : ", k)
			info.UnPackMongoValue(v, tempDBElems)
		} else {
			log.Error("loadFromDB, prop not exist: ", k)
		}
	}

	/*for _, elem := range tempDBElems {

		if elem.Name == "_id" {
			continue
		}

		info, ok := e.props[elem.Name]
		if ok {
			info.UnPackMongoValue(elem.Value.Data)

		} else {

		}
	}

	*/
}

// savePropsToMongoDB 属性保存到db
func (e *Entity) savePropsToMongoDB() {
	if e.GetEntityID() == 0 {
		return
	}

	if len(e.dirtySaveProps) == 0 {
		return
	}

	log.Debug("SyncProps len: ", len(e.dirtySaveProps))

	saveMap := bson.M{}
	for n, p := range e.dirtySaveProps {
		saveMap[n] = p.GetValue()

		p.dbFlag = false
	}

	log.Info("prop saveMap: ", saveMap)

	mongodbservice.MongoDBUpdate(PropDBName, PropTableName, bson.M{"dbid": e.GetEntityID()}, bson.M{"$set": saveMap})

	e.dirtySaveProps = make(map[string]*PropInfo)
}

// loadFromMysqlDB 从mysql加载数据
func (e *Entity) loadFromMysqlDB() {

	log.Debug("loadFromMysqlDB1.....")
	if e.GetEntityID() == 0 {
		return
	}

	var err error
	var shardObj *mysqlservice.MySQLShard
	shardObj, err = mysqlservice.GetShardObj(e.GetEntityID())
	if err != nil {
		log.Debug("GetShardObj failed, EntityID:%v", e.GetEntityID())
	}

	var selectProps []string
	var props_str string
	for k, v := range e.props {
		if v.def.Persistence {
			selectProps = append(selectProps, k)

			props_str = props_str + k
		}
	}

	var stmt *sql.Stmt
	stmt, err = shardObj.MysqlObj.Prepare(`select ? from ? where dbid=?;`)

	defer func() {
		if stmt != nil {
			stmt.Close()
		}
	}()

	rows, err := stmt.Query(props_str, PropTableName, e.GetEntityID())
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	if rows.Next() {

		//反射出未导出字段
		val := reflect.ValueOf(rows).Elem().FieldByName("lastcols")
		log.Debug("v = ", val, ", v.kind()", val.Kind())

		//与上面的区别是：这个是可寻址的
		val = reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem()
		if reflect.Slice == val.Kind() {
			if byteSlice, ok := val.Interface().([]byte); ok {
				fmt.Println("byteSlice = ", byteSlice)
			} else {
				len := val.Len()
				for i := 0; i < len; i++ {
					fmt.Println(utility.ConvertReflectVal(val.Index(i).Interface()))
				}
			}
		}
	}

	if err != nil {
		log.Error("query failed, Entity ID= ", e.GetEntityID(), ",err:", err)
	}

	/*log.Debug("loadFromMysqlDB3.....")
	db := mysqlservice.GetMysqlDB()
	if db != nil {
		db.Queryx("select $1 from $2 where dbid=$3", props_str, PropTableName, e.GetEntityID())
	*/

}

// savePropsToMysqlDB 属性保存到mysql db
func (e *Entity) savePropsToMysqlDB() {
	if e.GetEntityID() == 0 {
		return
	}

	if len(e.dirtySaveProps) == 0 {
		return
	}

	log.Debug("SyncProps len: ", len(e.dirtySaveProps))

	saveMap := bson.M{}
	for n, p := range e.dirtySaveProps {
		saveMap[n] = p.GetValue()

		p.dbFlag = false
	}

	log.Info("prop saveMap: ", saveMap)

	mongodbservice.MongoDBUpdate(PropDBName, PropTableName, bson.M{"dbid": e.GetEntityID()}, bson.M{"$set": saveMap})

	e.dirtySaveProps = make(map[string]*PropInfo)
}

// SavePropsToDB 属性保存到db
func (e *Entity) SavePropsToDB() {

	//判断用哪个db存
	if DBType == "mysql" {
		e.savePropsToMysqlDB()
	} else if DBType == "mongodb" {
		e.savePropsToMongoDB()
	}
}

// LoadFromDB 从db中加载属性
func (e *Entity) LoadFromDB() {

	//panic("load from db")
	//判断用哪个db
	if DBType == "mysql" {
		e.loadFromMysqlDB()
	} else if DBType == "mongodb" {
		e.loadFromMongoDB()
	}
}
