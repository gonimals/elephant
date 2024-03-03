package elephant

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/gonimals/elephant/internal/db"
	"github.com/gonimals/elephant/internal/db/sqlite3"
	"github.com/gonimals/elephant/internal/phanpy"
)

// MainContext is the default context
var MainContext Elephant

// db is the database driver in use
var dbDriver db.Driver

var currentElephants map[string]Elephant

// Elephant provides db access to a concrete context
//
// # Retrieve gets one element from a specific type filtering by key
// Returns the element if found and nil if not
//
// # RetrieveBy gets one element from a specific type filtering by other attribute
// Returns the element if found and nil if parameters are incorrect or the element is not found
//
// # RetrieveAll gets all elements with a specific type
// Returns a map with all elements. It will be empty if there are no elements
//
// # Remove deletes one element from the database
// Returns err if the object does not exist
//
// # RemoveByKey deletes one element from the database
// Returns err if the object does not exist
//
// # Update modifies an element on the database
//
// # Create adds one element to the database
// If the key attribute value is empty (""), a new one will be assigned
//
// # Exists checks if one key is in use in the database
//
// # ExistsBy gets one element from a specific type filtering by other attribute
// Returns true if found and false if parameters are incorrect or the element is not found
//
// # NextID gives an empty id to create a new entry
//
// # BlobCreate adds one byte blob to the database
//
// # BlobRemove removes one byte blob from the database
//
// # BlobRetrieve returns blob contents if found. If not, returns nil
//
// # Close should be called only from elephant.Close()
type Elephant interface {
	Retrieve(inputType reflect.Type, key string) interface{}
	RetrieveBy(inputType reflect.Type, attribute string, input interface{}) interface{}
	RetrieveAll(inputType reflect.Type) (map[string]interface{}, error)
	Remove(input interface{}) error
	RemoveByKey(inputType reflect.Type, key string) error
	Update(input interface{}) error
	Create(input interface{}) (string, error)
	Exists(inputType reflect.Type, key string) bool
	ExistsBy(inputType reflect.Type, attribute string, input interface{}) bool
	NextID(inputType reflect.Type) string
	BlobRetrieve(key string) *[]byte
	BlobCreate(key string, contents *[]byte) error
	BlobRemove(key string) error
	BlobUpdate(key string, contents *[]byte) error
	Close()
}

// Initialize requires a supported uri using one of the following supported formats
func Initialize(uri string) (err error) {
	sqlite3Regexp := regexp.MustCompile(`sqlite3://`)
	if sqlite3split := sqlite3Regexp.Split(uri, 2); len(sqlite3split) == 2 {
		dbDriver, err = sqlite3.Connect(sqlite3split[1])
	} else {
		err = fmt.Errorf("elephant: unsupported uri string: %s", uri)
	}
	if err == nil {
		currentElephants = make(map[string]Elephant)
		MainContext, err = GetElephant("")
	}
	return
}

// Close should be called as a deferred method after Initialize
func Close() {
	for _, e := range currentElephants {
		e.Close()
	}
	dbDriver.Close()
	dbDriver = nil
	MainContext = nil
	currentElephants = nil
}

// GetElephant returns a valid elephant for the required context
func GetElephant(context string) (e Elephant, err error) {
	if dbDriver == nil {
		return nil, fmt.Errorf("no database initialized")
	}
	e = currentElephants[context]
	if e == nil {
		p := phanpy.CreatePhanpy(context, dbDriver)
		currentElephants[context] = p
		e = p
	}
	return
}
