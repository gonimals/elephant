package sqlite3

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql" //Add support for mysql db
	"github.com/gonimals/elephant/internal/util"
)

// maxRegexLength decides the maximum length for any string checked with alphanumericRegexp
const maxRegexLength = "40"

// Regular expression used to check that no SQL injection is produced
var /* const */ alphanumericRegexp *regexp.Regexp = regexp.MustCompile("^[A-Za-z0-9_\\" + util.ContextSymbol + "]{1," + maxRegexLength + "}$")

// Name for the table to store byte blobs
const blobsTableName = "blobs"

// These are error strings returned by the driver
const errorSQLNoSuchTable = "no such table: "
const errorSQLNoRowsInResultSet = "sql: no rows in result set"
const errorPossibleSQLi = " could be a SQL injection attack"

// These are the sqlite table creation statements
const stmtCheckTable = "select id from %s limit 1"
const stmtCreateTable = "create table '%s' ( id text primary key, value text )"

// These are the statement names
const (
	stmtRetrieve = iota
	stmtRetrieveAll
	stmtInsert
	stmtDelete
	stmtUpdate
)

// creationMap is used just
var creationMap map[int]string

// This struct stores data needed to work with a struct in this DB
type typeHandler struct {
	name  string
	stmts map[int]*sql.Stmt
}

type driver struct {
	db           *sql.DB
	checkedTypes map[string]*typeHandler //checkedTypes stores types that have been already handled during the execution
	blobStmts    map[int]*sql.Stmt
}

func init() {
	creationMap = make(map[int]string)
	creationMap[stmtRetrieve] = "select value from '%s' where id = ?"
	creationMap[stmtRetrieveAll] = "select id, value from '%s'"
	creationMap[stmtInsert] = "insert into '%s' (id, value) values (?, ?)"
	creationMap[stmtDelete] = "delete from '%s' where id = ?"
	creationMap[stmtUpdate] = "update '%s' set value = ? where id = ?"
}

// Connect should be the first method called to initialize the db connection
func Connect(dataSourceName string) (output *driver, err error) {
	output = new(driver)
	output.db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatalln(err.Error())
	}
	output.checkedTypes = make(map[string]*typeHandler)
	output.ensureBlobsTableIsHandled()
	return
}

func (d *driver) Close() {
	d.db.Close()
}

// ensureBlobsTableIsHandled checks if the blobs table exists and creates it if not
func (d *driver) ensureBlobsTableIsHandled() {
	//Start the handling tasks
	var testID string
	err := d.db.QueryRow(fmt.Sprintf(stmtCheckTable, blobsTableName)).Scan(&testID)
	if err == nil || strings.Contains(err.Error(), errorSQLNoRowsInResultSet) {
		// Table exists and can be empty
	} else if err.Error() == errorSQLNoSuchTable+blobsTableName {
		// Table does not exist. Let's create it
		_, err := d.db.Exec(fmt.Sprintf("create table %s ( id text primary key, value longblob )", blobsTableName))
		if err != nil {
			log.Fatalln("Can't create blobs table", err)
		}
	} else {
		log.Fatalln("Unhandled error:", err)
	}
	d.blobStmts = make(map[int]*sql.Stmt)
	for i := stmtRetrieve; i <= stmtUpdate; i++ {
		d.blobStmts[i], err = d.db.Prepare(fmt.Sprintf(creationMap[i], blobsTableName))
		if err != nil {
			for _, stmt := range d.blobStmts {
				stmt.Close()
			}
			log.Fatalln("Cannot initialize blobs statements", err.Error())
		}
	}
}

// createTypeHandler just populates the struct with the required SQL statements
func (d *driver) createTypeHandler(input string) (th *typeHandler, err error) {
	th = new(typeHandler)
	th.name = input
	th.stmts = make(map[int]*sql.Stmt)
	for i := stmtRetrieve; i <= stmtUpdate; i++ {
		th.stmts[i], err = d.db.Prepare(fmt.Sprintf(creationMap[i], input))
		if err != nil {
			for _, stmt := range th.stmts {
				stmt.Close()
			}
			th.stmts = nil
			log.Println(err)
			return nil, err
		}
	}
	return
}

// ensureTableIsHandled checks if the table is already handled by the driver and handles it if not, checking for SQLi at source code
func (d *driver) ensureTableIsHandled(input string) (th *typeHandler) {
	th = d.checkedTypes[input]
	if th != nil {
		return //input is already handled
	}

	//Start the handling tasks
	var testID string
	if !alphanumericRegexp.MatchString(input) {
		log.Fatal(input + errorPossibleSQLi)
	}
	err := d.db.QueryRow(fmt.Sprintf(stmtCheckTable, input)).Scan(&testID)
	if err == nil || strings.Contains(err.Error(), errorSQLNoRowsInResultSet) {
		// Table exists and can be empty
	} else if err.Error() == errorSQLNoSuchTable+input {
		// Table does not exist. Let's create it
		_, err := d.db.Exec(fmt.Sprintf(stmtCreateTable, input))
		if err != nil {
			log.Fatalln("Can't create table for "+input, err)
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

func (d *driver) Retrieve(inputType string, key string) (output string, err error) {
	handledType := d.ensureTableIsHandled(inputType)
	err = handledType.stmts[stmtRetrieve].QueryRow(key).Scan(&output)
	return
}

func (d *driver) RetrieveAll(inputType string) (output map[string]string, err error) {
	handledType := d.ensureTableIsHandled(inputType)
	rows, err := handledType.stmts[stmtRetrieveAll].Query()
	if err != nil {
		return
	}
	output = make(map[string]string)
	for rows.Next() {
		var id string
		var value string
		err = rows.Scan(&id, &value)
		if err != nil {
			return
		}
		output[id] = value
	}
	return
}

func (d *driver) Remove(inputType string, key string) (err error) {
	handledType := d.ensureTableIsHandled(inputType)
	_, err = handledType.stmts[stmtDelete].Exec(key)
	return
}

func (d *driver) Create(inputType string, key string, input string) (err error) {
	handledType := d.ensureTableIsHandled(inputType)
	_, err = handledType.stmts[stmtInsert].Exec(key, input)
	return
}

func (d *driver) Update(inputType string, key string, input string) (err error) {
	handledType := d.ensureTableIsHandled(inputType)
	_, err = handledType.stmts[stmtUpdate].Exec(input, key)
	return
}

func (d *driver) BlobRetrieve(key string) (output *[]byte, err error) {
	err = d.blobStmts[stmtRetrieve].QueryRow(key).Scan(&output)
	return
}
func (d *driver) BlobCreate(key string, input *[]byte) (err error) {
	_, err = d.blobStmts[stmtInsert].Exec(key, input)
	return
}
func (d *driver) BlobUpdate(key string, input *[]byte) (err error) {
	result, err := d.blobStmts[stmtUpdate].Exec(input, key)
	if err != nil {
		return
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if affectedRows != 1 {
		return fmt.Errorf("sqlite3: blob update modified %d rows", affectedRows)
	}
	return
}
func (d *driver) BlobRemove(key string) (err error) {
	result, err := d.blobStmts[stmtDelete].Exec(key)
	if err != nil {
		return
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if affectedRows != 1 {
		return fmt.Errorf("sqlite3: blob delete modified %d rows", affectedRows)
	}
	return
}
