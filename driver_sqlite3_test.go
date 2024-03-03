package elephant

import (
	"bytes"
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
	err = db.dbCreate("test_table", "1", "asdfasdf")
	if err != nil {
		t.Error("simple create operation fails:", err)
	}
	output, err := db.dbRetrieve("test_table", "1")
	if err != nil {
		t.Error("simple retrieve operation fails:", err)
	}
	if output != "asdfasdf" {
		t.Error("retrieved string is not the original")
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
	err = db.dbBlobCreate("1", &[]byte{0x00})
	if err != nil {
		t.Error("blob create operation fails:", err)
	}
	blob, err := db.dbBlobRetrieve("1")
	if err != nil {
		t.Error("blob retrieve operation fails:", err)
	}
	if !blobsEqual(blob, &[]byte{0x00}) {
		t.Error("retrieved blob is not the original")
	}
	err = db.dbBlobUpdate("1", &[]byte{0x01})
	if err != nil {
		t.Error("blob update operation fails:", err)
	}
	blob, err = db.dbBlobRetrieve("1")
	if err != nil {
		t.Error("blob retrieve operation fails:", err)
	}
	if !blobsEqual(blob, &[]byte{0x01}) {
		t.Error("retrieved blob is not the updated one")
	}
	err = db.dbBlobRemove("1")
	if err != nil {
		t.Error("blob delete operation fails:", err)
	}
	err = db.dbBlobRemove("1")
	if err == nil {
		t.Error("blob delete operation should fail")
	}
	blob, err = db.dbBlobRetrieve("1")
	if err == nil {
		t.Error("retrieve operation of deleted blob doesn't give error")
	} else if blobsEqual(blob, &[]byte{0x00}) {
		t.Error("retrieve operation of deleted blob gives output:", blob)
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

func blobsEqual(input1, input2 *[]byte) bool {
	if input1 == input2 {
		return true
	}
	if input1 == nil || input2 == nil {
		return false
	}
	return bytes.Equal(*input1, *input2)
}
