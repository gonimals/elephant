package elephant

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gonimals/elephant/internal/db/sql"
	"github.com/gonimals/elephant/internal/util"
)

// Initialize requires a supported uri using one of the following supported formats
func Initialize(uri string) (err error) {
	switch {
	case strings.HasPrefix(uri, "sqlite3:"):
		dbDriver, err = sql.ConnectSqlite3(strings.TrimPrefix(uri, "sqlite3:"))
	case strings.HasPrefix(uri, "mysql:"):
		dbDriver, err = sql.ConnectMySQL(strings.TrimPrefix(uri, "mysql:"))
	default:
		err = fmt.Errorf("elephant: unsupported uri string: %s", uri)
	}
	if err != nil {
		dbDriver = nil
		return
	}

	data = make(map[reflect.Type](map[string]interface{}))
	learntTypes = make(map[reflect.Type]*util.LearntType)
	channel = make(chan *internalAction)
	managedTypes = make(map[reflect.Type]bool)
	learntTypes[blobReflectType] = &util.LearntType{
		Name: "blob",
	}
	managedTypes[blobReflectType] = true
	waitgroup.Add(1)
	go mainRoutine()
	return nil
}

// Retrieve gets one element from a specific type filtering by key. Returns the element if found and nil if not
func Retrieve[inputType interface{}](key string) *inputType {
	checkInitialization()
	action := newInternalAction(actionRetrieve, reflect.TypeFor[*inputType](), key)
	channel <- action
	switch v := (<-action.output).(type) {
	case *inputType:
		return v
	case nil:
		return nil
	default:
		panic(v)
	}
}

// RetrieveBy gets one element from a specific type filtering by other attribute. Returns the element if found and nil if parameters are incorrect or the element is not found
func RetrieveBy[inputType any](attribute string, input interface{}) *inputType {
	checkInitialization()
	action := newInternalAction(actionRetrieveBy, reflect.TypeFor[*inputType](), attribute, input)
	channel <- action
	switch v := (<-action.output).(type) {
	case *inputType:
		return v
	case nil:
		return nil
	default:
		panic(v)
	}
}

// RetrieveAll gets all elements with a specific type. Returns a map with all elements. It will be empty if there are no elements
func RetrieveAll[inputType any]() (map[string]*inputType, error) {
	checkInitialization()
	action := newInternalAction(actionRetrieveAll, reflect.TypeFor[*inputType](), nil)
	channel <- action
	switch v := (<-action.output).(type) {
	case map[string]interface{}:
		return util.CopyMapOfObjects[*inputType](v), nil
	case error:
		return nil, v
	case nil:
		return nil, nil
	default:
		panic(v)
	}
}

// Remove deletes one element from the database. Returns err if the object does not exist
func Remove(input interface{}) error {
	checkInitialization()
	action := newInternalAction(actionRemove, reflect.TypeOf(input), input)
	channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

// RemoveByKey deletes one element from the database. Returns err if the object does not exist
func RemoveByKey[inputType any](key string) error {
	checkInitialization()
	action := newInternalAction(actionRemoveByKey, reflect.TypeFor[*inputType](), key)
	channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

// Update modifies an element on the database
func Update(input interface{}) error {
	checkInitialization()
	action := newInternalAction(actionUpdate, reflect.TypeOf(input), input)
	channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

// Create adds one element to the database. If the key attribute value is empty (""), a new one will be assigned
func Create(input interface{}) (string, error) {
	checkInitialization()
	action := newInternalAction(actionCreate, reflect.TypeOf(input), input)
	channel <- action
	output := <-action.output
	if reflect.TypeOf(output).Kind() == reflect.String {
		return output.(string), nil
	}
	return "", output.(error)
}

// Exists checks if one key is in use in the database
func Exists[inputType any](key string) bool {
	checkInitialization()
	action := newInternalAction(actionExists, reflect.TypeFor[*inputType](), key)
	channel <- action
	return (<-action.output).(bool)
}

// ExistsBy gets one element from a specific type filtering by other attribute. Returns true if found and false if parameters are incorrect or the element is not found
func ExistsBy[inputType any](attribute string, input interface{}) bool {
	checkInitialization()
	action := newInternalAction(actionExistsBy, reflect.TypeFor[*inputType](), attribute, input)
	channel <- action
	return (<-action.output).(bool)
}

// NextID gives an empty id to create a new entry
func NextID[inputType any]() string {
	checkInitialization()
	action := newInternalAction(actionNextID, reflect.TypeFor[*inputType]())
	channel <- action
	return (<-action.output).(string)
}

// Upsert updates or inserts the entry. Returns the key of the modified object or an error
func Upsert(input interface{}) (string, error) {
	checkInitialization()
	action := newInternalAction(actionUpsert, reflect.TypeOf(input), input)
	channel <- action
	output := <-action.output
	if reflect.TypeOf(output).Kind() == reflect.String {
		return output.(string), nil
	}
	return "", output.(error)
}

// BlobRetrieve returns blob contents if found. If not, returns nil
func BlobRetrieve(key string) *[]byte {
	checkInitialization()
	action := newInternalAction(actionBlobRetrieve, blobReflectType, key)
	channel <- action
	output := <-action.output
	if output != nil {
		return output.(*[]byte)
	}
	return nil
}

// BlobCreate adds one byte blob to the database
func BlobCreate(key string, contents *[]byte) error {
	checkInitialization()
	action := newInternalAction(actionBlobCreate, blobReflectType, key, contents)
	channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

// BlobRemove removes one byte blob from the database
func BlobRemove(key string) error {
	checkInitialization()
	action := newInternalAction(actionBlobRemove, blobReflectType, key)
	channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

func BlobUpdate(key string, contents *[]byte) error {
	checkInitialization()
	action := newInternalAction(actionBlobUpdate, blobReflectType, key, contents)
	channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

// Close should be called as a deferred method after Initialize
func Close() {
	channel <- nil
	close(channel)
	waitgroup.Wait()
	dbDriver.Close()
	dbDriver = nil
}
