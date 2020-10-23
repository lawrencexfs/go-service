package mysql

import (
	"database/sql"
	"fmt"
	"testing"

	log "github.com/cihub/seelog"
)

type Member struct {
	UID   uint64
	LEVEL uint32
}

// orm使用文档 https://github.com/jmoiron/sqlx

func TestDBExec(t *testing.T) {
	mysqlDB := GetMysqlDB()
	if mysqlDB == nil {
		log.Error("createTable failed db is nil")
		return
	}

	_, err := mysqlDB.Exec("CREATE TABLE IF NOT EXISTS `team`(name2 text,desc2 text);")

	if err != nil {
		fmt.Printf("create table faied, error:[%v]", err.Error())
		return
	}

	_, err = mysqlDB.Exec("insert into team  values(?,?)", "aaaa", "bbbbbbbb")

	if err != nil {
		fmt.Printf("data insert faied, error:[%v]", err.Error())
		return
	}
}

func TestDBQuery(t *testing.T) {
	mysqlDB := GetMysqlDB()
	if mysqlDB == nil {
		log.Error("createTable failed db is nil")
		return
	}

	type Team struct {
		tname string `mysqlDB:"name2"`
		tdesc string `mysqlDB:"desc2"`
	}

	team := Team{}

	rows, err := mysqlDB.Queryx("select * from team ")

	for rows.Next() {
		err := rows.StructScan(&team)
		if err != nil {
			log.Critical(err)
		}
		fmt.Printf("%#v\n", team)
	}

	if err != nil {
		fmt.Printf("create table faied, error:[%v]", err.Error())
		return
	}

	//fmt.Printf("rows = ", rows)
}

func TestDBSelect(t *testing.T) {

	mysqlDB := GetMysqlDB()
	if mysqlDB == nil {
		log.Error("createTable failed db is nil")
		return
	}

	type Team struct {
		F string `db:"name2"`
		L string `db:"desc2"`
	}

	aaa := []Team{}

	mysqlDB.Select(&aaa, "SELECT * FROM team ")

	a := aaa[0]

	fmt.Printf("%#v", a)
}

func TestSqlx(t *testing.T) {

	/*	var schema1 = `
		CREATE TABLE IF NOT EXISTS person (
		    first_name text,
		    last_name text,
		    email text
		);`

			var schema2 = `
		CREATE TABLE IF NOT EXISTS place (
		    country text,
		    city text NULL,
		    telcode integer
		);`*/
	type Person struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string
	}

	type Place struct {
		Country string
		City    sql.NullString
		TelCode int
	}

	mysqlDB := GetMysqlDB()
	if mysqlDB == nil {
		log.Error("createTable failed db is nil")
		return
	}
	// exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers;  pq will exec them all, sqlite3 won't, ymmv
	//mysqlDB.MustExec("CREATE TABLE person (first_name text,last_name text,email text);")
	/*	mysqlDB.MustExec(schema1)
		mysqlDB.MustExec(schema2)

		tx := mysqlDB.MustBegin()
		tx.MustExec("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)", "Jason", "Moiron", "jmoiron@jmoiron.net")
		tx.MustExec("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)", "John", "Doe", "johndoeDNE@gmail.net")
		tx.MustExec("INSERT INTO place (country, city, telcode) VALUES (?, ?, ?)", "United States", "New York", "1")
		tx.MustExec("INSERT INTO place (country, telcode) VALUES (?, ?)", "Hong Kong", "852")
		tx.MustExec("INSERT INTO place (country, telcode) VALUES (?, ?)", "Singapore", "65")
		// Named queries can use structs, so if you have an existing struct (i.e. person := &Person{}) that you have populated, you can pass it in as &person
		tx.NamedExec("INSERT INTO person (first_name, last_name, email) VALUES (:first_name, :last_name, :email)", &Person{"Jane", "Citizen", "jane.citzen@example.com"})
		tx.Commit()
	*/
	// Query the database, storing results in a []Person (wrapped in []interface{})
	people := []Person{}
	mysqlDB.Select(&people, "SELECT * FROM person ORDER BY first_name ASC")
	jason, john := people[0], people[1]

	fmt.Printf("%#v\n%#v", jason, john)
	// Person{FirstName:"Jason", LastName:"Moiron", Email:"jmoiron@jmoiron.net"}
	// Person{FirstName:"John", LastName:"Doe", Email:"johndoeDNE@gmail.net"}

	// You can also get a single result, a la QueryRow
	jason = Person{}
	err = mysqlDB.Get(&jason, "SELECT * FROM person WHERE first_name=$1", "Jason")
	fmt.Printf("%#v\n", jason)
	// Person{FirstName:"Jason", LastName:"Moiron", Email:"jmoiron@jmoiron.net"}

	// if you have null fields and use SELECT *, you must use sql.Null* in your struct
	places := []Place{}
	err = mysqlDB.Select(&places, "SELECT * FROM place ORDER BY telcode ASC")
	if err != nil {
		fmt.Println(err)
		return
	}
	usa, singsing, honkers := places[0], places[1], places[2]

	fmt.Printf("%#v\n%#v\n%#v\n", usa, singsing, honkers)
	// Place{Country:"United States", City:sql.NullString{String:"New York", Valid:true}, TelCode:1}
	// Place{Country:"Singapore", City:sql.NullString{String:"", Valid:false}, TelCode:65}
	// Place{Country:"Hong Kong", City:sql.NullString{String:"", Valid:false}, TelCode:852}

	// Loop through rows using only one struct
	place := Place{}
	rows, err := mysqlDB.Queryx("SELECT * FROM place")
	for rows.Next() {
		err = rows.StructScan(&place)
		if err != nil {
			log.Critical(err)
		}
		fmt.Printf("%#v\n", place)
	}
	// Place{Country:"United States", City:sql.NullString{String:"New York", Valid:true}, TelCode:1}
	// Place{Country:"Hong Kong", City:sql.NullString{String:"", Valid:false}, TelCode:852}
	// Place{Country:"Singapore", City:sql.NullString{String:"", Valid:false}, TelCode:65}

	// Named queries, using `:name` as the bindvar.  Automatic bindvar support
	// which takes into account the dbtype based on the driverName on sqlx.Open/Connect
	_, err = mysqlDB.NamedExec(`INSERT INTO person (first_name,last_name,email) VALUES (:first,:last,:email)`,
		map[string]interface{}{
			"first": "Bin",
			"last":  "Smuth",
			"email": "bensmith@allblacks.nz",
		})

	// Selects Mr. Smith from the database
	rows, err = mysqlDB.NamedQuery(`SELECT * FROM person WHERE first_name=:fn`, map[string]interface{}{"fn": "Bin"})

	// Named queries can also use structs.  Their bind names follow the same rules
	// as the name -> db mapping, so struct fields are lowercased and the `db` tag
	// is taken into consideration.
	rows, err = mysqlDB.NamedQuery(`SELECT * FROM person WHERE first_name=:first_name`, jason)

}
