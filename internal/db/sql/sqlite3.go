package sql

import "regexp"

var /*const*/ stmtsSqlite3 = map[int]string{
	stmtCheckTable:  "select id from %s limit 1",
	stmtCreateTable: "create table '%s' ( id varchar(%d) primary key, value text )",
	stmtRetrieve:    "select value from '%s' where id = ?",
	stmtRetrieveAll: "select id, value from '%s'",
	stmtInsert:      "insert into '%s' (id, value) values (?, ?)",
	stmtDelete:      "delete from '%s' where id = ?",
	stmtUpdate:      "update '%s' set value = ? where id = ?",
	stmtCreateBlobs: "create table '%s' ( id varchar(%d) primary key, value longblob )",
}

var /*const*/ msgsSqlite3 = map[int]*regexp.Regexp{
	msgErrorNoSuchTable:       regexp.MustCompile(`no such table: `),
	msgErrorNoRowsInResultSet: regexp.MustCompile(`sql: no rows in result set`),
}

// Connect should be the first method called to initialize the db connection
func ConnectSqlite3(dataSourceName string) (output *driver, err error) {
	return connect("sqlite3", dataSourceName, stmtsSqlite3, msgsSqlite3, ".")
}
