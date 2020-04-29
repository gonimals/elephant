package elephant

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3" //Add support for sqlite3 db
)

//maxRegexLength decides the maximum length for any string checked with alphanumericRegexp
const maxRegexLength = "40"

//Regular expression used to check that no SQL injection is produced
var /* const */ alphanumericRegexp *regexp.Regexp = regexp.MustCompile("^[A-Za-z0-9_]{1," + maxRegexLength + "}$")

//These are error strings returned by the driver
const errorSQLNoSuchTable = "no such table: "
const errorSQLNoRowsInResultSet = "sql: no rows in result set"
const errorPossibleSQLi = " could be a SQL injection attack"

//These are the sqlite table creation statements
const sqlite3CheckTable = "select id from %s limit 1"
const sqlite3CreateTable = "create table %s ( id sqlite3_int64 primary key, value text )"

//These are the statement names
const (
	stmtRetrieve = iota
	stmtRetrieveAll
	stmtInsert
	stmtDelete
	stmtUpdate
)

//creationMap is used just
var sqliteCreationMap map[int]string

//This struct stores data needed to work with a struct in this DB
type typeHandler struct {
	name  string
	stmts map[int]*sql.Stmt
}

type driverSqlite3 struct {
	db           *sql.DB
	checkedTypes map[string]*typeHandler //checkedTypes stores types that have been already handled during the execution
}

func init() {
	sqliteCreationMap = make(map[int]string)
	sqliteCreationMap[stmtRetrieve] = "select value from %s where id = ?"
	sqliteCreationMap[stmtRetrieveAll] = "select id, value from %s"
	sqliteCreationMap[stmtInsert] = "insert into %s (id, value) values (?, ?)"
	sqliteCreationMap[stmtDelete] = "delete from %s where id = ?"
	sqliteCreationMap[stmtUpdate] = "update %s set value = ? where id = ?"
}

// ConnectSQLdriverDatabase should be the first method called to initialize the db connection
func sqlite3dbConnect(dataSourceName string) (output *driverSqlite3, err error) {
	output = new(driverSqlite3)
	output.db, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatalln(err.Error())
	}
	output.checkedTypes = make(map[string]*typeHandler)
	return
}

func (d *driverSqlite3) dbClose() {
	d.db.Close()
}

func cancelTypeHandlerCreation(th *typeHandler) *typeHandler {
	for _, stmt := range th.stmts {
		stmt.Close()
	}
	th.stmts = nil
	return nil
}

// createTypeHandler just populates the struct with the required SQL statements and checks for SQLi at source code
func (d *driverSqlite3) createTypeHandler(input string) (th *typeHandler, err error) {
	if !alphanumericRegexp.MatchString(input) {
		log.Fatal(input + errorPossibleSQLi)
	}
	th = new(typeHandler)
	th.name = input
	th.stmts = make(map[int]*sql.Stmt)
	for i := stmtRetrieve; i <= stmtUpdate; i++ {
		th.stmts[i], err = d.db.Prepare(fmt.Sprintf(sqliteCreationMap[i], input))
		if err != nil {
			log.Fatalln(err)
			return cancelTypeHandlerCreation(th), err
		}
	}
	return
}

func (d *driverSqlite3) ensureTableIsHandled(input string) (th *typeHandler) {
	th = d.checkedTypes[input]
	if th != nil {
		return //input is already handled
	}

	//Start the handling tasks
	var testID int64
	err := d.db.QueryRow(fmt.Sprintf(sqlite3CheckTable, input)).Scan(&testID)
	if err == nil || strings.Contains(err.Error(), errorSQLNoRowsInResultSet) {
		// Table exists and can be empty
	} else if err.Error() == errorSQLNoSuchTable+input {
		// Table does not exist. Let's create it
		_, err := d.db.Exec(fmt.Sprintf(sqlite3CreateTable, input))
		if err != nil {
			log.Fatalln("Can't create table for "+th.name, err)
		}
	} else {
		log.Fatalln("Unhandled error:", err)
	}
	th, err = d.createTypeHandler(input)
	if err != nil {
		log.Fatalln(err)
	}
	d.checkedTypes[input] = th
	return
}

func (d *driverSqlite3) dbRetrieve(inputType string, key int64) (output string, err error) {
	handledType := d.ensureTableIsHandled(inputType)
	err = handledType.stmts[stmtRetrieve].QueryRow(key).Scan(&output)
	return
}

func (d *driverSqlite3) dbRetrieveAll(inputType string) (output map[int]string, err error) {
	handledType := d.ensureTableIsHandled(inputType)
	rows, err := handledType.stmts[stmtRetrieveAll].Query()
	if err != nil {
		return
	}
	output = make(map[int]string)
	for rows.Next() {
		var id int
		var value string
		err = rows.Scan(&id, &value)
		if err != nil {
			return
		}
		output[id] = value
	}
	return
}

func (d *driverSqlite3) dbRemove(inputType string, key int64) (err error) {
	handledType := d.ensureTableIsHandled(inputType)
	_, err = handledType.stmts[stmtDelete].Exec(key)
	return
}

func (d *driverSqlite3) dbCreate(inputType string, key int64, input string) (err error) {
	handledType := d.ensureTableIsHandled(inputType)
	_, err = handledType.stmts[stmtInsert].Exec(key, input)
	return
}

func (d *driverSqlite3) dbUpdate(inputType string, key int64, input string) (err error) {
	handledType := d.ensureTableIsHandled(inputType)
	_, err = handledType.stmts[stmtUpdate].Exec(input, key)
	return
}
