package elephant

import (
	"fmt"
	"reflect"
	"regexp"
	"sync"
)

// MaxStructLength defines how long can be a structure converted to JSON to be stored
const MaxStructLength = 65535 //64k

// Elephant provides db access to a concrete context
type Elephant struct {
	Context      string
	data         map[reflect.Type](map[int64]interface{})
	learntTypes  map[reflect.Type]*learntType
	channel      chan *internalAction
	waitgroup    sync.WaitGroup
	managedTypes map[reflect.Type]bool
}

// MainContext is the default context
var MainContext *Elephant

// db is the database driver in use
var db dbDriver

// Initialize requires a supported uri using one of the following supported formats
func Initialize(uri string) (err error) {
	sqlite3 := regexp.MustCompile(`sqlite3://`)
	if sqlite3split := sqlite3.Split(uri, 2); len(sqlite3split) == 2 {
		db, err = sqlite3dbConnect(sqlite3split[1])
	} else {
		err = fmt.Errorf("elephant: unsupported uri string: %s", uri)
	}
	if err == nil {
		currentElephants = make(map[string]*Elephant)
		learntTypes = make(map[reflect.Type]*learntType)
		MainContext, err = GetElephant("")
	}
	return
}

// Close should be called as a deferred method after Initialize
func Close() {
	for _, e := range currentElephants {
		close(e.channel)
		e.waitgroup.Wait()
	}
	db.dbClose()
	db = nil
	MainContext = nil
	currentElephants = nil
	learntTypes = nil
}

// GetElephant returns a valid elephant for the required context
func GetElephant(context string) (e *Elephant, err error) {
	if db == nil {
		return nil, fmt.Errorf("No database initialized")
	}
	e = currentElephants[context]
	if e == nil {
		e = new(Elephant)
		e.Context = context
		e.data = make(map[reflect.Type](map[int64]interface{}))
		e.learntTypes = make(map[reflect.Type]*learntType)
		e.channel = make(chan *internalAction)
		e.managedTypes = make(map[reflect.Type]bool)
		e.waitgroup.Add(1)
		go e.mainRoutine()
		currentElephants[context] = e
	}
	return
}

// Retrieve gets one element from a specific type filtering by key
// Returns the element if found and nil if not
func (e *Elephant) Retrieve(inputType reflect.Type, key int64) interface{} {
	action := newInternalAction(actionRetrieve, inputType, key)
	e.channel <- action
	return <-action.output
}

// RetrieveBy gets one element from a specific type filtering by other attribute
// Returns the element if found and nil if parameters are incorrect or the element is not found
func (e *Elephant) RetrieveBy(inputType reflect.Type, attribute string, input interface{}) interface{} {
	action := newInternalAction(actionRetrieveBy, inputType, attribute, input)
	e.channel <- action
	return <-action.output
}

// RetrieveAll gets all elements with a specific type
// Returns a map with all elements. It will be empty if there are no elements
func (e *Elephant) RetrieveAll(inputType reflect.Type) map[int64]interface{} {
	action := newInternalAction(actionRetrieveAll, inputType, nil)
	e.channel <- action
	return (<-action.output).(map[int64]interface{})
}

// Remove deletes one element from the database
// Returns err if the object does not exist
func (e *Elephant) Remove(input interface{}) error {
	action := newInternalAction(actionRemove, reflect.TypeOf(input), input)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

// RemoveByKey deletes one element from the database
// Returns err if the object does not exist
func (e *Elephant) RemoveByKey(inputType reflect.Type, key int64) error {
	action := newInternalAction(actionRemoveByKey, inputType, key)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

// Update modifies an element on the database
func (e *Elephant) Update(input interface{}) error {
	/*_, err := getKey(input)
	if err != nil {
		return err
	}*/
	action := newInternalAction(actionUpdate, reflect.TypeOf(input), input)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

// Create adds one element to the database
// If the key attribute value is 0, a new one will be assigned
func (e *Elephant) Create(input interface{}) (int64, error) {
	action := newInternalAction(actionCreate, reflect.TypeOf(input), input)
	e.channel <- action
	output := <-action.output
	if reflect.TypeOf(output).Kind() == reflect.Int64 {
		return output.(int64), nil
	}
	return 0, output.(error)
}

// Exists check if one key is in use in the database
func (e *Elephant) Exists(inputType reflect.Type, key int64) bool {
	action := newInternalAction(actionExists, inputType, key)
	e.channel <- action
	return (<-action.output).(bool)
}

// ExistsBy gets one element from a specific type filtering by other attribute
// Returns true if found and false if parameters are incorrect or the element is not found
func (e *Elephant) ExistsBy(inputType reflect.Type, attribute string, input interface{}) bool {
	action := newInternalAction(actionExistsBy, inputType, attribute, input)
	e.channel <- action
	return (<-action.output).(bool)
}

// NextID gives an empty id to create a new entry
func (e *Elephant) NextID(inputType reflect.Type) int64 {
	action := newInternalAction(actionNextID, inputType)
	e.channel <- action
	return (<-action.output).(int64)
}
