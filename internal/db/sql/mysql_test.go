package sql

import (
	"database/sql"
	"fmt"
	"log"
	"testing"
)

// Refer to the README.md for the commands to perform these tests with this uri
const mysqlTestDB = "root:@tcp(127.0.0.1:33060)/elephant"

func TestMysql(t *testing.T) {
	db, err := ConnectMySQL(mysqlTestDB)
	if err != nil {
		t.Error("failed to create a valid db")
	}
	defer db.Close()
	testDriver(t, db)
}

// This function is just an mysql usage example
func TestDependencyMysql(t *testing.T) {
	// Connect
	db, err := sql.Open("mysql", mysqlTestDB)
	if err != nil {
		t.Error("Can't open the database", err)
		return
	}
	defer db.Close()

	// Drop table
	_, err = db.Exec(`drop table if exists foo;`)
	if err != nil {
		t.Error("Error dropping table before creation", err)
		return
	}

	// Create table
	_, err = db.Exec(`create table foo (id int not null primary key, name text);`)
	if err != nil {
		t.Error("Error creating table", err)
		return
	}

	// Insert transaction
	tx, err := db.Begin()
	if err != nil {
		t.Error("Error beginning transaction", err)
		return
	}
	stmt, err := tx.Prepare("insert into foo(id, name) values(?, ?)")
	if err != nil {
		t.Error("Error preparing statement", err)
		return
	}
	defer stmt.Close()
	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(int64(i), fmt.Sprintf("My name is %03d number", i))
		if err != nil {
			t.Error("Error inserting value", i, err)
			return
		}
	}
	tx.Commit()

	// Retrieve
	stmt, err = db.Prepare("select name from foo where id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var name string
	err = stmt.QueryRow("3").Scan(&name)
	if err != nil {
		log.Fatal(err)
	}
}
