package sql

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"

	_ "github.com/go-sql-driver/mysql" //Add support for mysql db
	"github.com/gonimals/elephant/internal/util"
	_ "github.com/mattn/go-sqlite3" //Add support for sqlite3 db
)

// MaxKeyLength sets the maximum string length for table keys
const MaxKeyLength = 512

// maxRegexLength decides the maximum length for any string checked with alphanumericRegexp
const maxRegexLength = "40"

// Regular expression used to check that no SQL injection is produced
var /* const */ alphanumericRegexp *regexp.Regexp = regexp.MustCompile("^[A-Za-z0-9_\\" + util.ContextSymbol + "]{1," + maxRegexLength + "}$")

// Name for the table to store byte blobs
const BlobsTableName = "blobs"

const errorPossibleSQLi = " could be a SQL injection attack"

// These are the statement names
const (
	stmtCheckTable = iota
	stmtCreateTable
	stmtRetrieve
	stmtRetrieveAll
	stmtInsert
	stmtDelete
	stmtUpdate
	stmtCreateBlobs
)

const (
	msgErrorNoSuchTable = iota
	msgErrorNoRowsInResultSet
)

// This struct stores data needed to work with a struct in this DB
type typeHandler struct {
	name  string
	stmts map[int]*sql.Stmt
}

type driver struct {
	db           *sql.DB
	checkedTypes map[string]*typeHandler //checkedTypes stores types that have been already handled during the execution
	blobStmts    map[int]*sql.Stmt
	baseStmts    map[int]string
	driverMsgs   map[int]*regexp.Regexp
}

// Connect should be the first method called to initialize the db connection
func connect(driverID string, dataSourceName string, baseStmts map[int]string, driverMsgs map[int]*regexp.Regexp) (output *driver, err error) {
	output = new(driver)
	output.db, err = sql.Open(driverID, dataSourceName)
	if err != nil {
		log.Fatalln(err.Error())
	}
	output.baseStmts = baseStmts
	output.driverMsgs = driverMsgs
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
	err := d.db.QueryRow(fmt.Sprintf(d.baseStmts[stmtCheckTable], BlobsTableName)).Scan(&testID)
	if err == nil || d.driverMsgs[msgErrorNoRowsInResultSet].MatchString(err.Error()) {
		// Table exists and can be empty
	} else if d.driverMsgs[msgErrorNoSuchTable].MatchString(err.Error()) {
		// Table does not exist. Let's create it
		_, err := d.db.Exec(fmt.Sprintf(d.baseStmts[stmtCreateBlobs], BlobsTableName, MaxKeyLength))
		if err != nil {
			log.Fatalln("Can't create blobs table", err)
		}
	} else {
		log.Fatalln("Unhandled error:", err)
	}
	d.blobStmts = make(map[int]*sql.Stmt)
	for i := stmtRetrieve; i <= stmtUpdate; i++ {
		d.blobStmts[i], err = d.db.Prepare(fmt.Sprintf(d.baseStmts[i], BlobsTableName))
		if err != nil {
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
		th.stmts[i], err = d.db.Prepare(fmt.Sprintf(d.baseStmts[i], input))
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
	err := d.db.QueryRow(fmt.Sprintf(d.baseStmts[stmtCheckTable], input)).Scan(&testID)
	if err == nil || d.driverMsgs[msgErrorNoRowsInResultSet].MatchString(err.Error()) {
		// Table exists and can be empty
	} else if d.driverMsgs[msgErrorNoSuchTable].MatchString(err.Error()) {
		// Table does not exist. Let's create it
		_, err := d.db.Exec(fmt.Sprintf(d.baseStmts[stmtCreateTable], input, MaxKeyLength))
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
	if len(key) > MaxKeyLength {
		return fmt.Errorf("sql: key too long (%d > %d)", len(key), MaxKeyLength)
	}
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
	if len(key) > MaxKeyLength {
		return fmt.Errorf("sql: key too long (%d > %d)", len(key), MaxKeyLength)
	}
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
		return fmt.Errorf("sql: blob update modified %d rows", affectedRows)
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
		return fmt.Errorf("sql: blob delete modified %d rows", affectedRows)
	}
	return
}
