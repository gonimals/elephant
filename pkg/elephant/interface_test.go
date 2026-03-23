package elephant

import (
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gonimals/elephant/internal/util"
)

func TestIncorrectUri(t *testing.T) {
	err := Initialize("")
	if err == nil {
		t.Error("Initialize not giving error with invalid uri")
	}
}

func TestInterfaceSqlite3(t *testing.T) {
	uri := "sqlite3:" + sqlite3TestDB

	os.Remove(sqlite3TestDB)
	testReuseDB(uri, t)

	os.Remove(sqlite3TestDB)
	testCorrectFunctions(uri, t)

	os.Remove(sqlite3TestDB)
	testUpdate(uri, t)

	os.Remove(sqlite3TestDB)
	testUpsert(uri, t)

	os.Remove(sqlite3TestDB)
	testCorrectBlobs(uri, t)
}

func TestDriverMySQL(t *testing.T) {
	uri := "mysql:" + mysqlTestDB

	cleanMysqlTestDB()
	testReuseDB(uri, t)

	cleanMysqlTestDB()
	testCorrectFunctions(uri, t)

	cleanMysqlTestDB()
	testUpdate(uri, t)

	cleanMysqlTestDB()
	testUpsert(uri, t)

	cleanMysqlTestDB()
	testCorrectBlobs(uri, t)
}

func testReuseDB(uri string, t *testing.T) {
	if err := Initialize(uri); err != nil {
		t.Error("Initialization failed", err)
	}
	var structCheckInstance = &structCheck{
		Myint64:  3,
		Mystring: "testing3",
		Myint:    987,
		Mybool:   false}

	if _, err := Create(structCheckInstance); err != nil {
		t.Error("Creation failed:", err)
	}
	Close()

	if err := Initialize(uri); err != nil {
		t.Error("Renitialization failed", err)
	}
	retrievedStruct, err := RetrieveBy[structCheck]("Mybool", structCheckInstance.Mybool)
	if err != nil || retrievedStruct.Myint != structCheckInstance.Myint {
		t.Error("Retrieved instance's Myint and the original one should be equal")
	}
	Close()
}

func testCorrectFunctions(uri string, t *testing.T) {
	if err := Initialize(uri); err != nil {
		t.Error("Initialization failed", err)
	}
	defer Close()

	var structCheckInstance = &structCheck{
		Myint64:  3,
		Mystring: "0",
		Myint:    987,
		Mybool:   false}

	Create(&structCheck{
		Myint64:  0,
		Mystring: "1",
		Myint:    234,
		Mybool:   true})
	Create(&structCheck{
		Myint64:  1,
		Mystring: "testingUpdate",
		Myint:    234,
		Mybool:   true})
	updateTest := &structCheck{
		Myint64:  5,
		Mystring: "testingUpdate",
		Myint:    2345,
		Mybool:   false}

	if _, err := Create(*structCheckInstance); err == nil {
		t.Error("Creation of non pointer struct should fail")
	}
	if _, err := Create(structCheckInstance); err != nil {
		t.Error("Creation failed:", err)
		return
	}
	if _, err := Create(new(failingStructCheck)); err == nil {
		t.Error("Creation of incorrect struct should fail")
	}
	retrieved, err := Retrieve[structCheck](structCheckInstance.Mystring)
	if err != nil {
		t.Error("Retrieve operation failed")
	}
	if comparison, err := util.CompareInstances(retrieved, structCheckInstance); err != nil || !comparison {
		t.Error("Retrieved instance and the original one should be equal")
	}
	if retrieved, err := RetrieveBy[structCheck]("Myint", structCheckInstance.Myint); err != nil {
		t.Error("RetrieveBy produced an error")
	} else if comparison, err := util.CompareInstances(retrieved, structCheckInstance); err != nil || !comparison {
		t.Error("Retrieved instance and the original one should be equal")
	}
	if retrieved, err := Retrieve[structCheck]("unexistent"); retrieved != nil || err != nil {
		t.Error("Retrieved instance and error should be nil:", retrieved, err)
	}
	if retrieved, err := RetrieveBy[structCheck]("unexistent", nil); err == nil || retrieved != nil {
		t.Error("An error should have been produced")
	}
	if retrieved, err := RetrieveBy[structCheck]("Myint", 456); err != nil || retrieved != nil {
		t.Error("Retrieved instance should be nil")
	}
	if Update(updateTest) != nil {
		t.Error("Update without errors should be nil")
	}
	updateTest.Mystring = "testingUpdateModified"
	if Update(updateTest) == nil {
		t.Error("Update of unexistent element should not be nil")
	}

	if err := RemoveById[structCheck]("1000"); err == nil {
		t.Error("Fake deletion didn't give any error")
	}

	if allInstances, _ := RetrieveAll[structCheck](); len(allInstances) != 3 {
		t.Error("Entire instances count differs after fake deletion")
	}
	newID, err := NewID[structCheck]()
	if err != nil {
		t.Error("NewID is throwing an error:", err)
	}
	if exists, err := Exists[structCheck](newID); exists || err != nil {
		t.Error("Exists returning true or an error on non-existent entry")
	}
	if exists, err := ExistsBy[structCheck]("Mystring", "unexistent"); exists || err != nil {
		t.Errorf("Exists returning true on non-existent entry")
	}
	if err := RemoveById[structCheck](structCheckInstance.Mystring); err != nil {
		t.Error("Remove operation failed, when should be correct:", err)
	}
	if allInstances, _ := RetrieveAll[structCheck](); len(allInstances) != 2 {
		t.Error("Entire instances count differs after deletion by id")
		for id, value := range data[reflect.TypeFor[structCheck]()] {
			log.Println(id, value)
		}
	}
	updateTest.Mystring = "testingUpdate"
	if err := Remove(updateTest); err != nil {
		t.Error("Remove operation failed, when should be correct:", err)
	}
	if allInstances, _ := RetrieveAll[structCheck](); len(allInstances) != 1 {
		t.Error("Entire instances count differs after deletion")
		for id, value := range data[reflect.TypeFor[structCheck]()] {
			log.Println(id, value)
		}
	}
	if err := Remove(updateTest); err == nil {
		t.Error("Remove operation successful, when should be incorrect")
	}
}

func testUpdate(uri string, t *testing.T) {
	if err := Initialize(uri); err != nil {
		t.Error("Initialization failed", err)
	}
	defer Close()

	if id, err := Create(&structCheck{
		Myint64:  0,
		Mystring: "1",
		Myint:    234,
		Mybool:   true}); id != "1" || err != nil {
		t.Error("Creation failed")
	}

	workingInstance, err := Retrieve[structCheck]("1")
	if workingInstance == nil || err != nil {
		t.Error("Retrieve operation failed")
		return //To avoid go-staticcheck SA5011
	}
	workingInstance.Mystring = strings.Repeat("A", util.MaxStructLength)
	if err := Update(workingInstance); err == nil {
		t.Error("Instance should be too long to be stored in the database")
	}
	workingInstance, err = Retrieve[structCheck]("1")
	if workingInstance == nil || err != nil {
		t.Error("Retrieve operation failed")
		return //To avoid go-staticcheck SA5011
	}
	if len(workingInstance.Mystring) == util.MaxStructLength {
		t.Error("The database runtime is not consistent with the stored data")
	}
}
func testUpsert(uri string, t *testing.T) {
	if err := Initialize(uri); err != nil {
		t.Error("Initialization failed", err)
	}
	defer Close()

	var id1, id2 string
	var err error
	if id1, err = Upsert(&structCheck{
		Myint64:  0,
		Mystring: "",
		Myint:    234,
		Mybool:   true}); id1 == "" || err != nil {
		t.Error("Creation failed")
	}

	if id2, err = Upsert(&structCheck{
		Myint64:  2,
		Mystring: "",
		Myint:    345,
		Mybool:   true}); id2 == "" || err != nil {
		t.Error("Creation failed")
	}

	workingInstance, err := Retrieve[structCheck](id1)
	if workingInstance == nil || err != nil {
		t.Error("Retrieve operation failed")
		return //To avoid go-staticcheck SA5011
	}
	workingInstance.Mystring = strings.Repeat("A", util.MaxStructLength)
	if _, err := Upsert(workingInstance); err == nil {
		t.Error("Instance should be too long to be stored in the database")
	}
	workingInstance, err = Retrieve[structCheck](id1)
	if workingInstance == nil || err != nil {
		t.Error("Retrieve operation failed")
		return //To avoid go-staticcheck SA5011
	}
	if len(workingInstance.Mystring) == util.MaxStructLength {
		t.Error("The database runtime is not consistent with the stored data")
	}
}

func testCorrectBlobs(uri string, t *testing.T) {
	if err := Initialize(uri); err != nil {
		t.Error("Initialization failed", err)
	}
	defer Close()

	if err := BlobCreate("1", &[]byte{0x00}); err != nil {
		t.Error("Creation of simple blob should not fail")
	}
	if err := BlobCreate("1", &[]byte{0x00}); err == nil {
		t.Error("Creation repeated blob should fail")
	}
	if retrieved, err := BlobRetrieve("1"); err != nil || !util.BlobsEqual(retrieved, &[]byte{0x00}) {
		t.Error("Retrieved blob and the original one should be equal")
	}
	if retrieved, err := BlobRetrieve("2"); retrieved != nil || err != nil {
		t.Error("Retrieved blob should be nil")
	}
	if exists, err := BlobExists("1"); !exists || err != nil {
		t.Error("")
	}
	if BlobUpdate("1", &[]byte{0x01}) != nil {
		t.Error("Update without errors should be nil")
	}
	if BlobUpdate("2", &[]byte{0x01}) == nil {
		t.Error("Update of unexistent blob should not be nil")
	}
	if BlobRemove("2") == nil {
		t.Error("Fake deletion didn't give any error")
	}
	if err := BlobRemove("1"); err != nil {
		t.Error("Remove operation failed, when should be correct:", err)
	}
	if err := BlobRemove("1"); err == nil {
		t.Error("Remove operation successful, when should be incorrect:", err)
	}
}
