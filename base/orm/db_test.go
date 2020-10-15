package orm

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

// orm使用文档 https://gorm.io/zh_CN/docs/query.html

func TestDBInsert(t *testing.T) {

	InitDB()
	if globalSession != nil {
		defer globalSession.Close() //一定要记得释放
		// 创建
		result := globalSession.Create(&Product{Code: "L1212", Price: 1000})
		if result.Error != nil {
			fmt.Println("err:", result.Error.Error())
		} else {
			fmt.Println("insert recode success,result.RowsAffected=", result.RowsAffected)
		}
	}
}

func TestDBQuery(t *testing.T) {

	InitDB()
	if globalSession != nil {
		defer globalSession.Close() //一定要记得释放
		// 创建
		var product Product
		result := globalSession.First(&product, 1)                  // 查询id为1的product
		result = globalSession.First(&product, "code = ?", "L1212") // 查询code为l1212的product
		if result.Error != nil {
			fmt.Println("err:", result.Error.Error())
		} else {
			fmt.Println("query success, result.RowsAffected=", result.RowsAffected)
		}
	}
}

func TestDBUpdate(t *testing.T) {

	InitDB()
	if globalSession != nil {
		defer globalSession.Close() //一定要记得释放

		// 更新
		var product Product
		product.Code = "L1212"
		product.Price = 3000
		result := globalSession.Model(&product).Update("Price", 33333)
		if result.Error != nil {
			fmt.Println("err:", result.Error.Error())
		} else {
			fmt.Println("update success, result.RowsAffected=", result.RowsAffected)
		}
	}
}

func TestDBDelete(t *testing.T) {

	InitDB()
	if globalSession != nil {
		defer globalSession.Close() //一定要记得释放

		var product Product
		//	product.ID = 1
		//	result := globalSession.Delete(&product) // 删除 - 删除product

		// 带额外条件的删除
		result := globalSession.Where("price = ?", 3000).Delete(&product)

		// 根据主键删除
		result = globalSession.Delete(&Product{}, 13)

		if result.Error != nil {
			fmt.Println("err:", result.Error.Error())
		} else {
			fmt.Println("delete success, result.RowsAffected=", result.RowsAffected)
		}
	}
}
