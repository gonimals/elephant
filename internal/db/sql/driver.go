package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql" //Add support for mysql db
	"github.com/gonimals/elephant/internal/util"
	_ "github.com/ncruces/go-sqlite3/driver" //Add support for sqlite3 db
	_ "github.com/ncruces/go-sqlite3/embed"  //Do not rely on external sqlite3 libraries
)

// MaxIdLength sets the maximum string length for table ids
const MaxIdLength = 512

// maxRegexLength decides the maximum length for any string checked with alphanumericRegexp
const maxRegexLength = "40"

// Regular expression used to check that no SQL injection is produced
var /* const */ alphanumericRegexp *regexp.Regexp = regexp.MustCompile("^[A-Za-z0-9_]{1," + maxRegexLength + "}$")

// Name for the table to store byte blobs
const BlobsTableName = "blobs"

// These are the statement names
const (
	stmtDropTable = iota
	stmtCheckTable
	stmtCreateTable
	stmtExists
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
	msgErrorConnectionRefused
)

// This struct stores data needed to work with a struct in this DB
type typeHandler struct {
	name  string
	stmts map[int]*sql.Stmt
}

type driver struct {
	db            *sql.DB
	checkedTypes  map[string]*typeHandler //checkedTypes stores types that have been already handled during the execution
	blobStmts     map[int]*sql.Stmt
	baseStmts     map[int]string
	driverMsgs    map[int]*regexp.Regexp
	contextSymbol string
}

// Connect should be the first method called to initialize the db connection
func connect(driverID string, dataSourceName string, baseStmts map[int]string, driverMsgs map[int]*regexp.Regexp, contextSymbol string) (output *driver, err error) {
	output = new(driver)
	output.db, err = sql.Open(driverID, dataSourceName)
	if err != nil {
		return nil, err
	}
	output.baseStmts = baseStmts
	output.driverMsgs = driverMsgs
	output.checkedTypes = make(map[string]*typeHandler)
	output.contextSymbol = contextSymbol
	err = output.ensureBlobsTableIsHandled()
	if err != nil {
		return nil, err
	}
	return
}

func (d *driver) Close() {
	d.db.Close()
}

// ensureBlobsTableIsHandled checks if the blobs table exists and creates it if not
func (d *driver) ensureBlobsTableIsHandled() error {
	//Start the handling tasks
	var testID string
	err := d.db.QueryRow(fmt.Sprintf(d.baseStmts[stmtCheckTable], BlobsTableName)).Scan(&testID)
	if err == nil || d.driverMsgs[msgErrorNoRowsInResultSet].MatchString(err.Error()) {
		// Table exists and can be empty
	} else if d.driverMsgs[msgErrorConnectionRefused].MatchString(err.Error()) {
		return util.Errorf("cannot connect to database: %v", err)
	} else if d.driverMsgs[msgErrorNoSuchTable].MatchString(err.Error()) {
		// Table does not exist. Let's create it
		_, err := d.db.Exec(fmt.Sprintf(d.baseStmts[stmtCreateBlobs], BlobsTableName, MaxIdLength))
		if err != nil {
			return util.Errorf("cannot create blobs table: %v", err)
		}
	} else {
		return util.Errorf("unhandled error: %v", err)
	}
	d.blobStmts = make(map[int]*sql.Stmt)
	for i := stmtExists; i <= stmtUpdate; i++ {
		d.blobStmts[i], err = d.db.Prepare(fmt.Sprintf(d.baseStmts[i], BlobsTableName))
		if err != nil {
			return util.Errorf("cannot initialize blobs statements: %v", err)
		}
	}
	return nil
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
func (d *driver) ensureTableIsHandled(input string) (th *typeHandler, err error) {
	th = d.checkedTypes[input]
	if th != nil {
		return //input is already handled
	}

	//Start the handling tasks
	var testID string
	if !alphanumericRegexp.MatchString(strings.ReplaceAll(input, d.contextSymbol, "")) {
		return nil, util.Errorf("possible SQL injection: %s", input)
	}
	err = d.db.QueryRow(fmt.Sprintf(d.baseStmts[stmtCheckTable], input)).Scan(&testID)
	if err == nil || d.driverMsgs[msgErrorNoRowsInResultSet].MatchString(err.Error()) {
		// Table exists and can be empty
		err = nil
	} else if d.driverMsgs[msgErrorNoSuchTable].MatchString(err.Error()) {
		// Table does not exist. Let's create it
		_, err := d.db.Exec(fmt.Sprintf(d.baseStmts[stmtCreateTable], input, MaxIdLength))
		if err != nil {
			return nil, util.Errorf("cannot create table for %s: %v", input, err)
		}
	} else {
		return nil, util.Errorf("unhandled error with query \"%s\": %v", fmt.Sprintf(d.baseStmts[stmtCheckTable], input), err)
	}
	th, err = d.createTypeHandler(input)
	if err != nil {
		return nil, util.Errorf("cannot create type handler for %s: %v", input, err)
	}
	d.checkedTypes[input] = th
	return
}

func (d *driver) Retrieve(inputType string, id string) (output string, err error) {
	handledType, err := d.ensureTableIsHandled(inputType)
	if err != nil {
		return
	}
	err = handledType.stmts[stmtRetrieve].QueryRow(id).Scan(&output)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return
}

func (d *driver) RetrieveAll(inputType string) (output map[string]string, err error) {
	handledType, err := d.ensureTableIsHandled(inputType)
	if err != nil {
		return
	}
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

func (d *driver) Remove(inputType string, id string) (err error) {
	handledType, err := d.ensureTableIsHandled(inputType)
	if err != nil {
		return
	}
	_, err = handledType.stmts[stmtDelete].Exec(id)
	return
}

func (d *driver) Create(inputType string, id string, input string) (err error) {
	handledType, err := d.ensureTableIsHandled(inputType)
	if err != nil {
		return
	}
	if len(id) > MaxIdLength {
		return util.Errorf("sql: id too long (%d > %d)", len(id), MaxIdLength)
	}
	_, err = handledType.stmts[stmtInsert].Exec(id, input)
	return
}

func (d *driver) Update(inputType string, id string, input string) (err error) {
	handledType, err := d.ensureTableIsHandled(inputType)
	if err != nil {
		return
	}
	_, err = handledType.stmts[stmtUpdate].Exec(input, id)
	return
}

func (d *driver) BlobRetrieve(id string) (output *[]byte, err error) {
	err = d.blobStmts[stmtRetrieve].QueryRow(id).Scan(&output)
	return
}
func (d *driver) BlobCreate(id string, input *[]byte) (err error) {
	if len(id) > MaxIdLength {
		return util.Errorf("sql: id too long (%d > %d)", len(id), MaxIdLength)
	}
	_, err = d.blobStmts[stmtInsert].Exec(id, input)
	return
}
func (d *driver) BlobUpdate(id string, input *[]byte) (err error) {
	result, err := d.blobStmts[stmtUpdate].Exec(input, id)
	if err != nil {
		return
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if affectedRows != 1 {
		return util.Errorf("sql: blob update modified %d rows", affectedRows)
	}
	return
}
func (d *driver) BlobRemove(id string) (err error) {
	result, err := d.blobStmts[stmtDelete].Exec(id)
	if err != nil {
		return
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return
	}
	if affectedRows != 1 {
		return util.Errorf("sql: blob delete modified %d rows", affectedRows)
	}
	return
}
func (d *driver) BlobExists(id string) (output bool, err error) {
	var outputID string
	err = d.blobStmts[stmtExists].QueryRow(id).Scan(&outputID)
	if err == nil {
		return true, nil
	} else if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}
func (d *driver) GetContextSymbol() string {
	return d.contextSymbol
}
