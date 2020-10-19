package gorm

//DBQuery db查询
/*func DBQuery(db string, collection string, queryCondition interface{}, retval interface{}) error {

	if globalSession != nil {
		defer globalSession.Close() //一定要记得释放
		c := globalSession.DB(db).C(collection)

		//var users []User
		//err := c.Find(nil).Limit(5).All(&users)
		err := c.Find(queryCondition).All(&retval)
		if err != nil {
			log.Error("DBQuery failed, err: ", err.Error(), ", condition: ", queryCondition, ", ret: ", retval)
			return err
		}
	} else {
		log.Error("globalSession nil, not connect or auth")
	}
	return nil
}

//DBInsert mongodb插入
func DBInsert(db string, collection string, data interface{}) error {

	if globalSession != nil {
		defer globalSession.Close() //一定要记得释放
		// 创建
		result := globalSession.Create(data)

		if result.Error != nil {
			log.Error("insert failed", result.Error.Error())
			return result.Error
		}
	} else {
		log.Error("globalSession nil, not connect or auth")
	}
	return nil
}

//DBUpdate  db更新
func DBUpdate(db string, collection string, updateCondition interface{}, data interface{}) error {
	if globalSession != nil {
		defer globalSession.Close() //一定要记得释放
		c := globalSession.DB(db).C(collection)

		err := c.Update(updateCondition, data)

		if err != nil {
			log.Error("update failed, ", err.Error())
			return err
		}
	} else {
		log.Error("globalSession nil, not connect or auth")
	}
	return nil

}

// DBDelete  db删除
func DBDelete(db string, collection string, removeCondition interface{}) error {

	if globalSession != nil {
		defer globalSession.Close() //一定要记得释放
		c := globalSession.DB(db).C(collection)
		_, err := c.RemoveAll(removeCondition)

		if err != nil {
			log.Error("delete failed", err.Error())
			return err
		}
	} else {
		log.Error("globalSession nil, not connect or auth")
	}
	return nil
}
*/
