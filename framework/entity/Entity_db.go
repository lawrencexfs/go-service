package entity

import (
	mongodbservice "github.com/giant-tech/go-service/base/mongodbservice"
	mysqlservice "github.com/giant-tech/go-service/base/mysqlservice"

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

	if e.GetEntityID() == 0 {
		return
	}

	/*	rows, err := mysqlDB.Queryx("select * from team ")

		for rows.Next() {
			err := rows.StructScan(&team)
			if err != nil {
				log.Critical(err)
			}
			fmt.Printf("%#v\n", team)
		}

		selectProps := bson.M{}
		for k, v := range e.props {
			if v.def.Persistence {
				selectProps[k] = 1
			}
		}
	*/
	var selectProps []string
	var props_str string
	for k, v := range e.props {
		if v.def.Persistence {
			selectProps = append(selectProps, k)

			props_str = props_str + k
		}
	}
	db := mysqlservice.GetMysqlDB()
	db.Queryx("select $1 from $2 where dbid=$3", props_str, PropTableName, e.GetEntityID())

	//var val_struct
	/*for rows.Next() {
		err := rows.StructScan(&val_struct)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("%#v\n", place)
	}*/

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
		e.savePropsToMongoDB()
	} else if DBType == "mongodb" {
		e.savePropsToMysqlDB()
	}
}

// LoadFromDB 从db中加载属性
func (e *Entity) LoadFromDB() {

	//判断用哪个db
	if DBType == "mysql" {
		e.loadFromMysqlDB()
	} else if DBType == "mongodb" {
		e.loadFromMongoDB()
	}
}
