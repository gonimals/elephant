package elephant

import (
	"reflect"
	"strings"

	"github.com/gonimals/elephant/internal/db/sql"
	"github.com/gonimals/elephant/internal/util"
)

// Initialize requires a supported uri using one of the following supported formats:
// - sqlite3:path/to/file.db
// - mysql:user:password@tcp(hostname:port)/database
func Initialize(uri string) (err error) {
	switch {
	case strings.HasPrefix(uri, "sqlite3:"):
		dbDriver, err = sql.ConnectSqlite3(strings.TrimPrefix(uri, "sqlite3:"))
	case strings.HasPrefix(uri, "mysql:"):
		dbDriver, err = sql.ConnectMySQL(strings.TrimPrefix(uri, "mysql:"))
	default:
		err = util.Errorf("unsupported uri string: %s", uri)
	}
	if err != nil {
		dbDriver = nil
		return
	}

	data = make(map[reflect.Type](map[string]any))
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

// Retrieve gets one element from a specific type filtering by id. Returns the element if found and nil if not
func Retrieve[inputType any](id string) (*inputType, error) {
	checkInitialization()
	action := newInternalAction(actionRetrieve, reflect.TypeFor[*inputType](), id)
	channel <- action
	return handleOutputType[*inputType](<-action.output, true)
}

// RetrieveBy gets one element from a specific type filtering by other attribute. Returns the element if found and nil if parameters are incorrect or the element is not found
func RetrieveBy[inputType any](attribute string, input any) (*inputType, error) {
	checkInitialization()
	action := newInternalAction(actionRetrieveBy, reflect.TypeFor[*inputType](), attribute, input)
	channel <- action
	return handleOutputType[*inputType](<-action.output, true)
}

// RetrieveAll gets all elements with a specific type. Returns a map with all elements. It will be empty if there are no elements
func RetrieveAll[inputType any]() (map[string]*inputType, error) {
	checkInitialization()
	action := newInternalAction(actionRetrieveAll, reflect.TypeFor[*inputType](), nil)
	channel <- action
	return handleOutputMap[map[string]*inputType](<-action.output)
}

// Remove deletes one element from the database. Returns err if the object does not exist
func Remove(input any) error {
	checkInitialization()
	action := newInternalAction(actionRemove, reflect.TypeOf(input), input)
	channel <- action
	return (<-action.output).err
}

// RemoveById deletes one element from the database. Returns err if the object does not exist
func RemoveById[inputType any](id string) error {
	checkInitialization()
	action := newInternalAction(actionRemoveById, reflect.TypeFor[*inputType](), id)
	channel <- action
	return (<-action.output).err
}

// Update modifies an element on the database
func Update(input any) error {
	checkInitialization()
	action := newInternalAction(actionUpdate, reflect.TypeOf(input), input)
	channel <- action
	return (<-action.output).err
}

// Create adds one element to the database. If the id attribute value is empty (""), a new one will be assigned
func Create(input any) (string, error) {
	checkInitialization()
	action := newInternalAction(actionCreate, reflect.TypeOf(input), input)
	channel <- action
	return handleOutputType[string](<-action.output, false)
}

// Exists checks if one id is in use in the database
func Exists[inputType any](id string) (bool, error) {
	checkInitialization()
	action := newInternalAction(actionExists, reflect.TypeFor[*inputType](), id)
	channel <- action
	return handleOutputType[bool](<-action.output, false)
}

// ExistsBy gets one element from a specific type filtering by other attribute. Returns true if found and false if parameters are incorrect or the element is not found
func ExistsBy[inputType any](attribute string, input any) (bool, error) {
	checkInitialization()
	action := newInternalAction(actionExistsBy, reflect.TypeFor[*inputType](), attribute, input)
	channel <- action
	return handleOutputType[bool](<-action.output, false)
}

// NewID gives an empty id to create a new entry
//
// Deprecated: It is better to leave the object without ID so elephant can assign the ID in that moment
func NewID[inputType any]() (string, error) {
	checkInitialization()
	action := newInternalAction(actionNextID, reflect.TypeFor[*inputType]())
	channel <- action
	return handleOutputType[string](<-action.output, false)
}

// Upsert updates or inserts the entry. Returns the id of the modified object or an error
func Upsert(input any) (string, error) {
	checkInitialization()
	action := newInternalAction(actionUpsert, reflect.TypeOf(input), input)
	channel <- action
	return handleOutputType[string](<-action.output, false)
}

// BlobRetrieve returns blob contents if found. If not, returns nil
func BlobRetrieve(id string) (*[]byte, error) {
	checkInitialization()
	action := newInternalAction(actionBlobRetrieve, blobReflectType, id)
	channel <- action
	return handleOutputType[*[]byte](<-action.output, false)
}

// BlobCreate adds one byte blob to the database
func BlobCreate(id string, contents *[]byte) error {
	checkInitialization()
	action := newInternalAction(actionBlobCreate, blobReflectType, id, contents)
	channel <- action
	return (<-action.output).err
}

// BlobRemove removes one byte blob from the database
func BlobRemove(id string) error {
	checkInitialization()
	action := newInternalAction(actionBlobRemove, blobReflectType, id)
	channel <- action
	return (<-action.output).err
}

func BlobUpdate(id string, contents *[]byte) error {
	checkInitialization()
	action := newInternalAction(actionBlobUpdate, blobReflectType, id, contents)
	channel <- action
	return (<-action.output).err
}

// BlobExists checks if one id is in use in the blobs table
func BlobExists(id string) (bool, error) {
	checkInitialization()
	action := newInternalAction(actionBlobExists, blobReflectType, id)
	channel <- action
	return handleOutputType[bool](<-action.output, false)
}

// Close should be called as a deferred method after Initialize
func Close() {
	channel <- nil
	close(channel)
	waitgroup.Wait()
	dbDriver.Close()
	dbDriver = nil
}

func handleOutputType[inputType any](output actionOutput, copy bool) (inputType, error) {
	var zero inputType
	if output.err != nil {
		return zero, output.err
	}
	if util.IsNilable[inputType]() && output.data == nil {
		return zero, nil
	}
	if reflect.TypeFor[inputType]().Kind() == reflect.Map {
		return zero, util.Errorf("error handling map instance")
	}
	v, ok := (output.data).(inputType)
	if !ok {
		return zero, util.Errorf("error casting output to defined type")
	}
	if !copy {
		return v, nil
	}
	objectCopy, err := util.CopyEntireObject(v)
	if err != nil {
		return zero, err
	}
	return objectCopy.(inputType), nil
}

func handleOutputMap[inputType map[string]K, K any](output actionOutput) (inputType, error) {
	var zero inputType
	originalMap, ok := (output.data).(map[string]any)
	if !ok {
		return zero, util.Errorf("error casting output map")
	}
	mapCopy, err := util.CopyMapOfObjects[inputType](originalMap)
	if err != nil {
		return zero, err
	}
	if outputMap, ok := any(mapCopy).(inputType); ok {
		return outputMap, nil
	}
	return zero, util.Errorf("error copying map instance")
}
