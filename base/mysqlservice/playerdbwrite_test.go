package mysqlservice

// import (
// 	"fmt"
// 	"runtime"
// 	"testing"
// 	"time"
// )

// func pushSQL(pwb *playerWriteDB) {
// 	idSQL := 0
// 	for {
// 		idSQL++
// 		SQLFmt := `
// 			INSERT INTO player (
// 				role_id,
// 				account,
// 				name,
// 				uuid,
// 				level
// 			)
// 			VALUES (%d, '%s', '%s', '%s', %d)
// 			ON DUPLICATE KEY UPDATE
// 				role_id=%d,
// 				account='%s',
// 				name='%s',
// 				uuid='%s',
// 				level=%d
// 		`

// 		sql := fmt.Sprintf(SQLFmt, uint64(idSQL), "liugao", "liugaotest", "u_liugao", int(1),
// 			uint64(idSQL), "liugao", "liugaotest", "u_liugao", int(1))

// 		pwb.ExecSQL(sql)
// 	}

// }

// func TestPlayerDBWrite(t *testing.T) {
// 	mysqlObj, err := InitMySQL("")
// 	if err != nil || mysqlObj == nil {
// 		t.Fatalf("unable to decode into struct, %v", err)
// 	}

// 	roleID := uint64(2000)

// 	roleObj := NewWriteRoleDB(mysqlObj, roleID)

// 	go pushSQL(roleObj)

// 	runtimes := 0
// 	c := time.NewTicker(time.Second)
// 	for {
// 		<-c.C
// 		runtimes++
// 		fmt.Println("timer runnintg... runtimes:", runtimes)
// 		fmt.Println("goroutine num:", runtime.NumGoroutine())
// 	}
// }

// func TestPlayerDBWriteLoop(t *testing.T) {
// 	mysqlObj, err := InitMySQL("")
// 	if err != nil || mysqlObj == nil {
// 		t.Fatalf("unable to decode into struct, %v", err)
// 	}

// 	roleID := uint64(2000)

// 	roleObj := NewWriteRoleDB(mysqlObj, roleID)

// 	for i := 50000; i < 101000; i++ {
// 		SQLFmt := `
// 			INSERT INTO player (
// 				role_id,
// 				account,
// 				name,
// 				uuid,
// 				level
// 			)
// 			VALUES (%d, '%s', '%s', '%s', %d)
// 			ON DUPLICATE KEY UPDATE
// 				role_id=%d,
// 				account='%s',
// 				name='%s',
// 				uuid='%s',
// 				level=%d
// 		`

// 		sql := fmt.Sprintf(SQLFmt, uint64(i), "liugao", "liugaotest", "u_liugao", int(3),
// 			uint64(i), "liugao", "liugaotest", "u_liugao", int(3))

// 		roleObj.ExecSQL(sql)
// 	}

// 	runtimes := 0
// 	c := time.NewTicker(time.Second)
// 	for {
// 		<-c.C
// 		runtimes++
// 		fmt.Println("timer runnintg... runtimes:", runtimes)
// 		fmt.Println("goroutine num:", runtime.NumGoroutine())
// 	}
// }
