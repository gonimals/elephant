package elephant

import (
	"fmt"
	"reflect"
	"regexp"
)

// MaxStructLength defines how long can be a structure converted to JSON to be stored
const MaxStructLength = 65535 //64k

// MainContext is the default context
var MainContext Elephant

// db is the database driver in use
var db dbDriver

// Elephant provides db access to a concrete context
//
// Retrieve gets one element from a specific type filtering by key
// Returns the element if found and nil if not
//
// RetrieveBy gets one element from a specific type filtering by other attribute
// Returns the element if found and nil if parameters are incorrect or the element is not found
//
// RetrieveAll gets all elements with a specific type
// Returns a map with all elements. It will be empty if there are no elements
//
// Remove deletes one element from the database
// Returns err if the object does not exist
//
// RemoveByKey deletes one element from the database
// Returns err if the object does not exist
//
// # Update modifies an element on the database
//
// Create adds one element to the database
// If the key attribute value is empty (""), a new one will be assigned
//
// # Exists check if one key is in use in the database
//
// ExistsBy gets one element from a specific type filtering by other attribute
// Returns true if found and false if parameters are incorrect or the element is not found
//
// # NextID gives an empty id to create a new entry
//
// close should be called to stop Elephant
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

	close()
}

// Initialize requires a supported uri using one of the following supported formats
func Initialize(uri string) (err error) {
	sqlite3 := regexp.MustCompile(`sqlite3://`)
	if sqlite3split := sqlite3.Split(uri, 2); len(sqlite3split) == 2 {
		db, err = sqlite3dbConnect(sqlite3split[1])
	} else {
		err = fmt.Errorf("elephant: unsupported uri string: %s", uri)
	}
	if err == nil {
		currentElephants = make(map[string]Elephant)
		learntTypes = make(map[reflect.Type]*learntType)
		MainContext, err = GetElephant("")
	}
	return
}

// Close should be called as a deferred method after Initialize
func Close() {
	for _, e := range currentElephants {
		e.close()
	}
	db.dbClose()
	db = nil
	MainContext = nil
	currentElephants = nil
	learntTypes = nil
}

// GetElephant returns a valid elephant for the required context
func GetElephant(context string) (e Elephant, err error) {
	if db == nil {
		return nil, fmt.Errorf("no database initialized")
	}
	e = currentElephants[context]
	if e == nil {
		p := new(phanpy)
		p.Context = context
		p.data = make(map[reflect.Type](map[string]interface{}))
		p.learntTypes = make(map[reflect.Type]*learntType)
		p.channel = make(chan *internalAction)
		p.managedTypes = make(map[reflect.Type]bool)
		p.waitgroup.Add(1)
		go p.mainRoutine()
		currentElephants[context] = p
		e = p
	}
	return
}
