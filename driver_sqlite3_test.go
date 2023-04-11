package elephant

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

//t.Error("Current test")

func TestDriverSqlite3(t *testing.T) {
	os.Remove(temporaryDB)
	db, err := sqlite3dbConnect(temporaryDB)
	if err != nil {
		t.Error("failed to create a valid db")
	}
	defer db.dbClose()
	db.dbCreate("test_table", "1", "asdfasdf")
	if err != nil {
		t.Error("simple create operation fails:", err)
	}
	output, err := db.dbRetrieve("test_table", "1")
	if err != nil {
		t.Error("simple retrieve operation fails:", err)
	}
	if output != "asdfasdf" {
		t.Errorf("retrieved string is not the original")
	}
	err = db.dbUpdate("test_table", "1", "fdsafdsa")
	if err != nil {
		t.Error("simple update operation fails:", err)
	}
	err = db.dbUpdate("test_table", "1", "fdsafdsa")
	if err != nil {
		t.Error("simple update operation fails:", err)
	}
	output, err = db.dbRetrieve("test_table", "1")
	if err != nil {
		t.Error("simple retrieve operation fails:", err)
	}
	if output != "fdsafdsa" {
		t.Errorf("retrieved string is not the updated one")
	}
	err = db.dbRemove("test_table", "1")
	if err != nil {
		t.Error("simple delete operation fails:", err)
	}
	output, err = db.dbRetrieve("test_table", "1")
	if err == nil {
		t.Error("retrieve operation of deleted item doesn't give error")
	} else if output != "" {
		t.Error("retrieve operation of deleted item gives output:", output)
	}

}

// This function is just an sqlite3 usage example
func TestSqlite3(t *testing.T) {
	os.Remove(temporaryDB)
	// Connect
	db, err := sql.Open("sqlite3", temporaryDB)
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
	os.Remove(temporaryDB)
}
