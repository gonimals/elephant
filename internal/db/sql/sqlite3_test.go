package sql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
)

const sqlite3TestDB = "/tmp/foo.db"

func TestSqlite3(t *testing.T) {
	db, err := ConnectSqlite3(sqlite3TestDB)
	if err != nil {
		t.Error("failed to create a valid db")
	}
	defer db.Close()
	testDriver(t, db)
}

// This function is just an sqlite3 usage example
func TestDependencySqlite3(t *testing.T) {
	os.Remove(sqlite3TestDB)
	// Connect
	db, err := sql.Open("sqlite3", sqlite3TestDB)
	if err != nil {
		t.Error("Can't open the database", err)
		return
	}
	defer db.Close()

	// Create table
	sqlStmt := `
	create table foo (id sqlite3_int64 not null primary key, name text);
	delete from foo;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		t.Error("Error creating table", err, "\n", sqlStmt)
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

	// Reset
	os.Remove(sqlite3TestDB)
}
