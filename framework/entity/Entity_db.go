package entity

import (
	dbservice "github.com/giant-tech/go-service/base/mongodbservice"

	log "github.com/cihub/seelog"
	"github.com/globalsign/mgo/bson"
)

// PropDBName PropDBName
var PropDBName string

// PropTableName PropTableName
var PropTableName string

// loadFromDB 从数据库中恢复
func (e *Entity) loadFromDB() {
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

	dbservice.MongoDBQueryOneWithSelect(PropDBName, PropTableName, bson.M{"dbid": e.GetEntityID()}, selectProps, &retRawData)

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

// SavePropsToDB 属性保存到db
func (e *Entity) SavePropsToDB() {
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

	dbservice.MongoDBUpdate(PropDBName, PropTableName, bson.M{"dbid": e.GetEntityID()}, bson.M{"$set": saveMap})

	e.dirtySaveProps = make(map[string]*PropInfo)
}
