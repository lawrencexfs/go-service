package entity

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	mongodbservice "github.com/giant-tech/go-service/base/mongodbservice"
	mysqlservice "github.com/giant-tech/go-service/base/mysqlservice"
	utils "github.com/giant-tech/go-service/base/utility"
	"github.com/go-sql-driver/mysql"
	"github.com/gogo/protobuf/proto"

	"database/sql"

	log "github.com/cihub/seelog"
	"github.com/globalsign/mgo/bson"
)

// DBType dbtype
var DBType string
var isAutoLoadSave bool

// PropDBName PropDBName
var PropDBName string

// PropTableName PropTableName
var PropTableName string

// SetDBType 包函数
func SetDBType(dbtype string) {
	log.Info("SetDBType type: ", dbtype)
	DBType = dbtype
}

func GetDBType() string {
	return DBType
}

func IsAutoLoadSave() bool {
	return isAutoLoadSave
}

func SetLoadSaveFlag(flag bool) {
	isAutoLoadSave = flag
}

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

	if e.GetEntityID() == 0 {
		return
	}

	var err error
	var shardObj *mysqlservice.MySQLShard
	shardObj, err = mysqlservice.GetShardObj(e.GetEntityID())
	if err != nil {
		log.Error("loadFromMysqlDB GetShardObj failed, EntityID:%v", e.GetEntityID())
	}

	type ValInfo struct {
		colname string
		val     interface{}
	}
	var valArray []ValInfo
	var props_str string
	var number = 0
	var valinfo ValInfo

	for k, v := range e.props {
		if v.def.Persistence {
			if number >= 1 {
				valinfo.colname = k
				k = "," + k
			} else {

				valinfo.colname = k
			}
			props_str = props_str + k
			valArray = append(valArray, valinfo)
			number++
		}
	}

	if props_str != "" {
		var stmt *sql.Stmt
		var prepare_str string
		prepare_str = fmt.Sprintf("select %s from %s where entity_id=?;", props_str, e.GetType())

		log.Debug(" loadFromMysqlDB prepare_str: ", prepare_str, " entityID=", e.GetEntityID(), ", entitytype=", e.GetType())
		//stmt, err = shardObj.MysqlObj.Prepare(`select ? from ? where role_id=?;`)
		stmt, err = shardObj.MysqlObj.Prepare(prepare_str)
		if stmt == nil || err != nil {

			mySQLErr, ok := err.(*mysql.MySQLError)
			if ok {
				if mySQLErr.Number == 1146 {
					log.Error("loadFromMysqlDB stmt nil, err =  ")
				} else {
					log.Error("loadFromMysqlDB stmt nil, err = ", err)
				}
			}
		} else {
			rows, err := stmt.Query( /*props_str, */ e.GetEntityID())
			if err != nil {
				log.Error("loadFromMysqlDB stmt Query failed err = ", err)
			}

			if rows.Next() {

				//反射出未导出字段
				val := reflect.ValueOf(rows).Elem().FieldByName("lastcols")
				//log.Debug("v = ", val, ", v.kind(): ", val.Kind())

				//与上面的区别是：这个是可寻址的
				val = reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem()
				//log.Debug("after reflect, val = ", val)
				if reflect.Slice == val.Kind() {
					//log.Debug("val.Kind = ", val)
					if byteSlice, ok := val.Interface().([]byte); ok {
						log.Debug("byteSlice = ", byteSlice)
					} else {
						len := val.Len()
						for i := 0; i < len; i++ {
							//log.Debug("val index i: ", i, ", index val : ", utility.ConvertReflectVal(val.Index(i).Interface()))
							valArray[i].val = val.Index(i).Interface()
						}
					}
				}
				for _, v := range valArray {

					info, ok := e.props[v.colname]
					if ok {
						info.UnPackMysqlValue(v.val)
					} else {
						log.Error("loadFromMysqlDB, prop not exist: ", v.colname)
					}
				}
			} else {
				//没有查询到entity的记录，创建角色，插入一行entity记录
				insertBuf := bytes.NewBufferString("INSERT INTO ")
				insertBuf.WriteString(e.entityType)
				insertBuf.WriteString(" (entity_id)")
				//insertBuf.WriteString(string(esc(props_str)))
				insertBuf.WriteString(" VALUES(")
				insertBuf.WriteString(utils.ConvertTypeToString(e.GetEntityID()))
				insertBuf.WriteString(" );")

				_, err := shardObj.MysqlObj.Exec(insertBuf.String())
				if err != nil {
					log.Error("loadFromMysqlDB insert sql:", insertBuf.String(), "error:", err)
				} else {
					log.Info("loadFromMysqlDB insert sql:", insertBuf.String())
				}
			}
		}
	} else {
		log.Info("loadFromMysqlDB prop_str nil:  entityID=", e.GetEntityID(), ", entitytype=", e.GetType())
	}
}

// savePropsToMysqlDB 属性保存到mysql db
func (e *Entity) savePropsToMysqlDB() {
	if e.GetEntityID() == 0 {
		return
	}

	if len(e.dirtySaveProps) == 0 {
		return
	}

	log.Debug("Entity: ", e.GetEntityID(), " savePropsToMysqlDB SyncProps len: ", len(e.dirtySaveProps))

	//拼凑属性更新语句，更新到mysql数据库
	insertBuf := bytes.NewBufferString("update ")
	insertBuf.WriteString(e.entityType)
	insertBuf.WriteString(" set ")

	esc := mysqlservice.EscapeBytesBackslash1
	var valstr string
	count := 1
	for n, p := range e.dirtySaveProps {
		insertBuf.WriteString(n)
		insertBuf.WriteString("=")
		if strings.Contains(p.def.TypeName, "protoMsg") {
			val, err := proto.Marshal(p.GetValue().(proto.Message))
			if err != nil {
				log.Error("Marshal prop val failed typename:", p.def.TypeName)
			}
			val = esc(val)
			insertBuf.WriteString("'")
			valstr = string(val)
			insertBuf.WriteString(valstr)
			insertBuf.WriteString("'")

		} else {

			valstr = utils.ConvertTypeToString(p.GetValue())
			insertBuf.WriteString(valstr)
		}

		if count < len(e.dirtySaveProps) {
			insertBuf.WriteString(",")
		}
		count++
	}

	var err error
	var shardObj *mysqlservice.MySQLShard
	shardObj, err = mysqlservice.GetShardObj(e.GetEntityID())
	if err != nil {
		log.Error("savePropsToMysqlDB GetShardObj failed, EntityID:%v", e.GetEntityID())
	}

	_, err = shardObj.MysqlObj.Exec(insertBuf.String())
	if err != nil {
		log.Error("Entity:", e.GetEntityID(), " ,savePropsToMysqlDB insertBuf= ", insertBuf.String(), " , error:", err)
	} else {
		log.Debug("Entity:", e.GetEntityID(), " ,savePropsToMysqlDB insert sql: ", insertBuf.String(), " success")
	}

	//清空数据
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
