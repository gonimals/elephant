package sql

import (
	"regexp"
)

var /*const*/ stmtsMysql = map[int]string{
	stmtCheckTable:  "select id from %s limit 1",
	stmtCreateTable: "create table %s ( id varchar(%d) primary key, value longtext )",
	stmtRetrieve:    "select value from %s where id = ?",
	stmtRetrieveAll: "select id, value from %s",
	stmtInsert:      "insert into %s (id, value) values (?, ?)",
	stmtDelete:      "delete from %s where id = ?",
	stmtUpdate:      "update %s set value = ? where id = ?",
	stmtCreateBlobs: "create table %s ( id varchar(%d) primary key, value longblob )",
}

var /*const*/ msgsMysql = map[int]*regexp.Regexp{
	msgErrorNoSuchTable:       regexp.MustCompile(`Table '[^']+' doesn't exist`),
	msgErrorNoRowsInResultSet: regexp.MustCompile(`sql: no rows in result set`),
}

// Connect should be the first method called to initialize the db connection
func ConnectMySQL(dataSourceName string) (output *driver, err error) {
	return connect("mysql", dataSourceName, stmtsMysql, msgsMysql)
}
