package entity

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime/debug"

	mysqlservice "github.com/giant-tech/go-service/base/mysqlservice"
	"github.com/giant-tech/go-service/base/stream"

	log "github.com/cihub/seelog"

	"github.com/giant-tech/go-service/framework/idata"
)

// InitData 初始化data
func (e *Entity) CreateEntityTable() {
	if GetDBType() == "mysql" && e.isCreateTable {
		//创建entity数据库表
		var err error
		var shardObj *mysqlservice.MySQLShard
		shardObj, err = mysqlservice.GetShardObj(e.GetEntityID())
		if err != nil {
			log.Error("OnEntityInit GetShardObj failed, EntityID:%v", e.GetEntityID())
		}

		sqlBuf := bytes.NewBufferString("CREATE TABLE IF NOT EXISTS`")
		sqlBuf.WriteString(e.entityType)
		sqlBuf.WriteString("`")
		sqlBuf.WriteString("(")
		sqlBuf.WriteString(" \n")

		sqlBuf.WriteString("	`ID` bigint(0) UNSIGNED NOT NULL AUTO_INCREMENT,")
		sqlBuf.WriteString(" \n")
		sqlBuf.WriteString("	`entity_id` bigint(0) NOT NULL,")
		sqlBuf.WriteString(" \n")

		//add prop
		for name, prop := range e.props {
			if prop.def.Persistence {
				st := prop.def.TypeName
				switch st {
				case "bool", "int8":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" tinyint(0) NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				case "int16":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" smallint(0) NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				case "int32":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" int(0) NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				case "int64":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" bigint(0) NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				case "uint8":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" tinyint(0) UNSIGNED NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				case "uint16":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" smallint(0) UNSIGNED NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				case "uint32":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" int(0) UNSIGNED NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				case "uint64":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" bigint(0) UNSIGNED NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				case "string":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				case "float32":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" float NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				case "float64":
					sqlBuf.WriteString("	`")
					sqlBuf.WriteString(name)
					sqlBuf.WriteString("`")
					sqlBuf.WriteString(" double NULL DEFAULT NULL,")
					sqlBuf.WriteString(" \n")
				}
			}
		}

		sqlBuf.WriteString("	PRIMARY KEY (`ID`) USING BTREE")
		sqlBuf.WriteString("\n")
		sqlBuf.WriteString(") ENGINE = InnoDB AUTO_INCREMENT = 200 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;")

		/*sql := "IF NOT EXISTS `Player`;
		CREATE TABLE `Player`  (
		  `ID` bigint(0) UNSIGNED NOT NULL AUTO_INCREMENT,
		  `ENTITY_TYPE` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,
		  PRIMARY KEY (`mailID`, `receiverID`) USING BTREE
		) ENGINE = InnoDB AUTO_INCREMENT = 200 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;
		*/
		result, err := shardObj.MysqlObj.Exec(sqlBuf.String())
		if err != nil {
			log.Error("result= ", result, ", err=", err, ", sql str=", sqlBuf.String(), ", stack=", string(debug.Stack()))
		}

	}
	//todo:mongodb不需要事先创建表
}

// InitProp 初始化属性列表
func (e *Entity) InitProp(def *Def) {
	if def == nil {
		return
	}

	e.isCreateTable = false
	e.def = def
	for _, p := range e.def.Props {
		for _, sType := range p.Sync {
			if uint32(e.GetIEntities().GetLocalService().GetSType()) == sType {
				e.addProp(p)
				//只要有一个属性需要持久化，就创建entity数据库表
				if p.Persistence {
					e.isCreateTable = true
				}
				break
			}
		}
	}
}

// addProp 添加属性
func (e *Entity) addProp(prop *PropDef) {
	e.props[prop.Name] = newPropInfo(prop)
}

// SetProp 设置一个属性的值 上层开发者调用
func (e *Entity) SetProp(name string, v interface{}) {

	p := e.props[name]

	//log.Debug("SetProp name begin: ", name, " p.value: ", p.value)
	if e.def == nil {
		panic(fmt.Errorf("no def file exist %s", name))
	}

	if !p.def.IsValidValue(v) {
		panic(fmt.Errorf("prop type error %s", name))
	}

	if reflect.DeepEqual(p.value, v) {
		//log.Debug("SetProp DeepEqual return: ", name)
		return
	}

	p.value = v

	//log.Debug("SetProp name end: ", name)
	e.PropDirty(name)
}

// GetProp 获取一个属性的值
func (e *Entity) GetProp(name string) interface{} {
	return e.props[name].value
}

// PropDirty 设置某属性为Dirty
func (e *Entity) PropDirty(name string) {
	p := e.props[name]

	if p.syncFlag == false {
		e.dirtyPropList = append(e.dirtyPropList, p)
	}
	p.syncFlag = true

	if p.def.Persistence {
		if p.dbFlag == false {
			e.dirtySaveProps[name] = p
		}
		p.dbFlag = true
	}
}

// FlushDirtyProp 处理脏属性
func (e *Entity) FlushDirtyProp() {
	e.SyncProps()
	e.SavePropsToDB()
}

func (e *Entity) getDirtyProps(syncType uint32) []*PropInfo {
	m, ok := e.dirtySyncProps[syncType]
	if !ok {
		m = make([]*PropInfo, 0, 10)
		e.dirtySyncProps[syncType] = m
	}
	return m
}

// SendPropsSyncMsg 发送属性同步消息
func (e *Entity) SendPropsSyncMsg(propMap map[uint32][]*PropInfo) error {
	//msg := &msgdef.SyncProps{}
	//msg.EntityID = e.entityID

	for syncType, v := range propMap {
		//msg.PropNum = uint32(len(v))
		//msg.Data = PackPropsToBytes(v)

		servicebase := e.GetIEntities().GetLocalService()
		stype := servicebase.GetSType()
		if syncType != uint32(idata.ServiceClient) && syncType != uint32(stype) {

			log.Debug("SendPropsSyncMsg, syncType: ", syncType)
			e.AsyncCall(idata.ServiceType(syncType), "SyncProps", uint32(len(v)), PackPropsToBytes(v))

			/*s := iserver.GetServiceProxyMgr().GetServiceByID(sid)
			if !s.IsValid() {
				log.Error("AsyncCall service not exist: ", sid, ",  sType: ", sType, " ,methodName: ", methodName)
				return fmt.Errorf("service not exist: %d", sid)
			}



			sid, _, err := e.GetEntitySrvID(uint8(syncType))
			if err != nil {
				log.Error("SendPropsSyncMsg GetEntitySrvID error: ", err, ", syncType: ", syncType)
				return err
			}

			s := iserver.GetServiceProxyMgr().GetServiceByID(sid)

			if !s.IsValid() {
				log.Error("SendPropsSyncMsg service not exist: ", sid, ",  syncType: ", syncType)
				return fmt.Errorf("service not exist: %d", sid)
			}

			if s.IsLocal() {
				//直接发送
				is := iserver.GetLocalServiceMgr().GetLocalService(sid)
				if is == nil {
					log.Error("SendPropsSyncMsg error, service is local, but not found, sid: ", sid)
					return fmt.Errorf("SendPropsSyncMsg error, service is local, but not found")
				}

				return is.PostCallMsg(msg)
			}

			return s.SendMsg(msg)
			*/
		}
	}

	return nil
}

// SyncProps 同步属性给其他人
func (e *Entity) SyncProps() {
	//log.Debug("SyncProps len1: ", len(e.dirtyPropList))
	if len(e.dirtyPropList) == 0 {
		return
	}

	log.Debug("SyncProps len: ", len(e.dirtyPropList))

	for _, p := range e.dirtyPropList {
		for _, s := range p.def.Sync {
			m := e.getDirtyProps(s)
			e.dirtySyncProps[s] = append(m, p)
		}

		p.syncFlag = false
	}

	e.dirtyPropList = e.dirtyPropList[0:0]

	if len(e.dirtySyncProps) > 0 {
		e.SendPropsSyncMsg(e.dirtySyncProps)
	}
}

// UpdateFromMsg 从消息中更新属性
func (e *Entity) UpdateFromMsg(num int, data []byte) {
	bs := stream.NewByteStream(data)
	for i := 0; i < num; i++ {
		name, err := bs.ReadStr()
		if err != nil {
			//e.Error("read prop name fail ", err)
			return
		}

		prop, ok := e.props[name]
		if !ok {
			//e.Error("target entity not own prop ", name)
			return
		}

		err = prop.ReadValueFromStream(bs)
		if err != nil {
			//e.Error("read prop from stream failed ", name, err)
			return
		}
	}
}

// PackProps 打包属性
func (e *Entity) PackProps(syncType uint32) (uint32, []byte) {

	props := make([]*PropInfo, 0, 10)

	for _, prop := range e.props {
		for _, sync := range prop.def.Sync {
			if sync == syncType {
				props = append(props, prop)
				break
			}
		}
	}

	return uint32(len(props)), PackPropsToBytes(props)
}

// PackPropsToBytes 把属列表打包成bytes
func PackPropsToBytes(props []*PropInfo) []byte {
	size := 0

	for _, prop := range props {
		size = size + len(prop.def.Name) + 2
		size = size + prop.GetValueStreamSize()
	}

	if size == 0 {
		return nil
	}

	buf := make([]byte, size)
	bs := stream.NewByteStream(buf)

	for _, prop := range props {
		if err := bs.WriteStr(prop.def.Name); err != nil {
			//e.Error("Pack props failed ", err, prop.def.Name)
		}
		if err := prop.WriteValueToStream(bs); err != nil {
			//e.Error("Pack props failed ", err, prop.def.Name)
		}
	}

	return buf
}
