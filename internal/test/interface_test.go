package test

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/gonimals/elephant"
	"github.com/gonimals/elephant/internal/phanpy"
	"github.com/gonimals/elephant/internal/util"
)

func TestReuseDB(t *testing.T) {
	os.Remove(temporaryDB)
	err := elephant.Initialize("sqlite3://" + temporaryDB)
	if err != nil {
		t.Error("Initialization failed", err)
	}
	structCheckType := reflect.TypeOf((*structCheck)(nil))
	var structCheckInstance = &structCheck{
		Myint64:  3,
		Mystring: "testing3",
		Myint:    987,
		Mybool:   false}

	if _, err = elephant.MainContext.Create(structCheckInstance); err != nil {
		t.Error("Creation failed")
	}
	elephant.Close()
	err = elephant.Initialize("sqlite3://" + temporaryDB)
	if err != nil {
		t.Error("Renitialization failed", err)
	}
	retrievedStruct := elephant.MainContext.RetrieveBy(structCheckType, "Mybool", structCheckInstance.Mybool)
	if retrievedStruct.(*structCheck).Myint != structCheckInstance.Myint {
		t.Error("Retrieved instance's Myint and the original one should be equal")
	}
	elephant.Close()
}

/*
func TestStringKey(t *testing.T) {
	os.Remove(temporaryDB)
	err := elephant.Initialize("sqlite3://" + temporaryDB)
	if err != nil {
		t.Error("Initialization failed", err)
	}
	defer elephant.Close()

	key, err := elephant.MainContext.Create(&stringStructCheck{
		Mystring: "Testing",
		Mydate:   0,
	})
	if err != nil || key.(string) != "Testing" {
		t.Error("Creation of struct with string key should be allowed")
	}
}
*/

func TestCorrectFunctions(t *testing.T) {
	os.Remove(temporaryDB)
	err := elephant.Initialize("sqlite3://" + temporaryDB)
	if err != nil {
		t.Error("Initialization failed", err)
	}
	defer elephant.Close()
	structCheckType := reflect.TypeOf((*structCheck)(nil))

	var structCheckInstance = &structCheck{
		Myint64:  3,
		Mystring: "0",
		Myint:    987,
		Mybool:   false}

	elephant.MainContext.Create(&structCheck{
		Myint64:  0,
		Mystring: "1",
		Myint:    234,
		Mybool:   true})
	elephant.MainContext.Create(&structCheck{
		Myint64:  1,
		Mystring: "testingUpdate",
		Myint:    234,
		Mybool:   true})
	updateTest := &structCheck{
		Myint64:  5,
		Mystring: "testingUpdate",
		Myint:    2345,
		Mybool:   false}

	if _, err = elephant.MainContext.Create(*structCheckInstance); err == nil {
		t.Error("Creation of non pointer struct should fail")
	}
	if _, err = elephant.MainContext.Create(structCheckInstance); err != nil {
		t.Error("Creation failed")
	}
	if _, err = elephant.MainContext.Create(new(failingStructCheck)); err == nil {
		t.Error("Creation of incorrect struct should fail")
	}
	if !util.CompareInstances(elephant.MainContext.Retrieve(structCheckType, structCheckInstance.Mystring), structCheckInstance) {
		t.Error("Retrieved instance and the original one should be equal")
	}
	if !util.CompareInstances(elephant.MainContext.RetrieveBy(structCheckType, "Myint", structCheckInstance.Myint), structCheckInstance) {
		t.Error("Retrieved instance and the original one should be equal")
	}
	if elephant.MainContext.RetrieveBy(structCheckType, "unexistent", nil) != nil {
		t.Error("Retrieved instance should be nil")
	}
	if elephant.MainContext.RetrieveBy(structCheckType, "Myint", 456) != nil {
		t.Error("Retrieved instance should be nil")
	}
	if elephant.MainContext.Update(updateTest) != nil {
		t.Error("Update without errors should be nil")
	}
	updateTest.Mystring = "testingUpdateModified"
	if elephant.MainContext.Update(updateTest) == nil {
		t.Error("Update of unexistent element should not be nil")
	}
	err = elephant.MainContext.RemoveByKey(structCheckType, "1000")
	if err == nil {
		t.Error("Fake deletion didn't give any error")
	}

	if allInstances, _ := elephant.MainContext.RetrieveAll(structCheckType); len(allInstances) != 3 {
		t.Error("Entire instances count differs after fake deletion")
	}
	if elephant.MainContext.NextID(structCheckType) != "2" {
		t.Error("NextID is not giving the first empty ID:", elephant.MainContext.NextID(structCheckType))
	}
	if elephant.MainContext.Exists(structCheckType, "2") {
		t.Error("Exists returning true on non-existent entry")
	}
	if elephant.MainContext.ExistsBy(structCheckType, "Mystring", "unexistent") {
		t.Errorf("Exists returning true on non-existent entry")
	}
	err = elephant.MainContext.RemoveByKey(structCheckType, structCheckInstance.Mystring)
	if err != nil {
		t.Error("Remove operation failed, when should be correct:", err)
	}
	if allInstances, _ := elephant.MainContext.RetrieveAll(structCheckType); len(allInstances) != 2 {
		t.Error("Entire instances count differs after deletion by key")
		for key, value := range elephant.MainContext.(*phanpy.Phanpy).Data[structCheckType] {
			log.Println(key, value)
		}
	}
	updateTest.Mystring = "testingUpdate"
	if elephant.MainContext.Remove(updateTest) != nil {
		t.Error("Remove operation failed, when should be correct:", err)
	}
	if allInstances, _ := elephant.MainContext.RetrieveAll(structCheckType); len(allInstances) != 1 {
		t.Error("Entire instances count differs after deletion")
		for key, value := range elephant.MainContext.(*phanpy.Phanpy).Data[structCheckType] {
			log.Println(key, value)
		}
	}
	if elephant.MainContext.Remove(updateTest) == nil {
		t.Error("Remove operation successful, when should be incorrect:", err)
	}
	customContext, err := elephant.GetElephant("customContext")
	if err != nil {
		t.Error("Error getting custom context", err)
	}
	if _, err = customContext.Create(&structCheck{
		Myint64:  0,
		Mystring: `[ {name: 'item', sort: [0 ,0] } ]`,
	}); err != nil {
		t.Error("Creation failed with custom context")
	}
}

func TestIncorrectUri(t *testing.T) {
	err := elephant.Initialize("")
	if err == nil {
		t.Error("elephant.Initialize not giving error with invalid uri")
	}
	nilObject, err := elephant.GetElephant("")
	if nilObject != nil {
		t.Error("Unelephant.Initialized library is giving instances")
	}
	if err == nil {
		t.Error("Unelephant.Initialized library is not giving error when asking for an instance")
	}
}

func TestCorrectBlobs(t *testing.T) {
	os.Remove(temporaryDB)
	err := elephant.Initialize("sqlite3://" + temporaryDB)
	if err != nil {
		t.Error("Initialization failed", err)
	}
	defer elephant.Close()

	if err = elephant.MainContext.BlobCreate("1", &[]byte{0x00}); err != nil {
		t.Error("Creation of simple blob should not fail")
	}
	if err = elephant.MainContext.BlobCreate("1", &[]byte{0x00}); err == nil {
		t.Error("Creation repeated blob should fail")
	}
	if !util.BlobsEqual(elephant.MainContext.BlobRetrieve("1"), &[]byte{0x00}) {
		t.Error("Retrieved blob and the original one should be equal")
	}
	if elephant.MainContext.BlobRetrieve("2") != nil {
		t.Error("Retrieved blob should be nil")
	}
	if elephant.MainContext.BlobUpdate("1", &[]byte{0x01}) != nil {
		t.Error("Update without errors should be nil")
	}
	if elephant.MainContext.BlobUpdate("2", &[]byte{0x01}) == nil {
		t.Error("Update of unexistent blob should not be nil")
	}
	if elephant.MainContext.BlobRemove("2") == nil {
		t.Error("Fake deletion didn't give any error")
	}
	if elephant.MainContext.BlobRemove("1") != nil {
		t.Error("Remove operation failed, when should be correct:", err)
	}
	if elephant.MainContext.BlobRemove("1") == nil {
		t.Error("Remove operation successful, when should be incorrect:", err)
	}
}
