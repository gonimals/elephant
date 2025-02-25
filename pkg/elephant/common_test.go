package elephant

import (
	"database/sql"
	"fmt"
	"log"
)

const mysqlTestDB = "root:@tcp(127.0.0.1:33060)/elephant"
const sqlite3TestDB = "/tmp/foo.db"

type structCheck struct {
	Mystring string `db:"key"`
	Myint    int
	Myint64  int64
	Mybool   bool
}

type failingStructCheck struct {
	Mystring string
	Myint    int `db:"key"`
	Myint64  int64
	Mybool   bool
}

func cleanMysqlTestDB() {
	// Connect
	db, err := sql.Open("mysql", mysqlTestDB)
	if err != nil {
		log.Fatalln("Cannot connect to testing mysql:", err)
	}
	defer db.Close()

	// Retrieve table names
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'elephant';")
	if err != nil {
		log.Fatalln("Cannot retrieve table names from testing mysql:", err)
	}
	var tableNames []string
	var name string
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			log.Fatalln("Cannot read table names from testing mysql:", err)
		}
		tableNames = append(tableNames, name)
	}
	rows.Close()

	// Drop all tables
	for _, table := range tableNames {
		_, err := db.Exec(fmt.Sprintf("drop table if exists `%s`", table))
		if err != nil {
			log.Fatalln("Cannot run drop on testing mysql:", err)
		}
	}
}
