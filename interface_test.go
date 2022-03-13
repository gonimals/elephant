package elephant

import (
	"log"
	"os"
	"reflect"
	"testing"
)

func TestCorrectFunctions(t *testing.T) {
	os.Remove(temporaryDB)
	err := Initialize("sqlite3://" + temporaryDB)
	if err != nil {
		t.Error("Initialization failed", err)
	}
	defer Close()
	structCheckType := reflect.TypeOf((*structCheck)(nil))

	var structCheckInstance = &structCheck{
		Myint64:  3,
		Mystring: "testing3",
		Myint:    987,
		Mybool:   false}

	MainContext.Create(structCheckType, &structCheck{
		Myint64:  0,
		Mystring: "testing",
		Myint:    234,
		Mybool:   true})
	MainContext.Create(structCheckType, &structCheck{
		Myint64:  1,
		Mystring: "testing",
		Myint:    234,
		Mybool:   true})

	if _, err = MainContext.Create(structCheckType, structCheckInstance); err != nil {
		t.Error("Creation failed")
	}
	if MainContext.Retrieve(structCheckType, structCheckInstance.Myint64) != structCheckInstance {
		t.Error("Retrieved instance and the original one should be equal")
	}
	if MainContext.RetrieveBy(structCheckType, "Mystring", structCheckInstance.Mystring) != structCheckInstance {
		t.Error("Retrieved instance and the original one should be equal")
	}
	if MainContext.RetrieveBy(structCheckType, "unexistent", nil) != nil {
		t.Error("Retrieved instance should be nil")
	}
	if MainContext.RetrieveBy(structCheckType, "Mystring", "unexistent") != nil {
		t.Error("Retrieved instance should be nil")
	}
	err = MainContext.Remove(structCheckType, 1000)
	if err == nil {
		t.Error("Fake deletion didn't give any error")
	}
	if len(MainContext.RetrieveAll(structCheckType)) != 3 {
		t.Error("Entire instances count differs after fake deletion")
	}
	if MainContext.NextID(structCheckType) != 2 {
		t.Error("NextID is not giving the first empty ID")
	}
	if MainContext.Exists(structCheckType, 2) {
		t.Error("Exists returning true on non-existent entry")
	}
	err = MainContext.Remove(structCheckType, structCheckInstance.Myint64)
	if err != nil {
		t.Error("Remove operation failed, when should be correct:", err)
	}
	if len(MainContext.RetrieveAll(structCheckType)) != 2 {
		t.Error("Entire instances count differs after deletion")
		for key, value := range MainContext.data[structCheckType] {
			log.Println(key, value)
		}
	}
	customContext, err := GetElephant("customContext")
	if err != nil {
		t.Error("Error getting custom context", err)
	}
	if _, err = customContext.Create(structCheckType, &structCheck{
		Myint64:  0,
		Mystring: `[ {name: 'item', sort: [0 ,0] } ]`,
	}); err != nil {
		t.Error("Creation failed with custom context")
	}
}

func TestIncorrectUri(t *testing.T) {
	err := Initialize("")
	if err == nil {
		t.Error("Initialize not giving error with invalid uri")
	}
	nilObject, err := GetElephant("")
	if nilObject != nil {
		t.Error("Uninitialized library is giving instances")
	}
	if err == nil {
		t.Error("Uninitialized library is not giving error when asking for an instance")
	}
}
