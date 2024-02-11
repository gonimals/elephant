package elephant

import (
	"reflect"
	"sync"
)

// phanpy is the initial Elephant implementation
type phanpy struct {
	Context      string
	data         map[reflect.Type](map[string]interface{})
	learntTypes  map[reflect.Type]*learntType
	channel      chan *internalAction
	waitgroup    sync.WaitGroup
	managedTypes map[reflect.Type]bool
}

// Retrieve gets one element from a specific type filtering by key
// Returns the element if found and nil if not
func (e *phanpy) Retrieve(inputType reflect.Type, key string) interface{} {
	action := newInternalAction(actionRetrieve, inputType, key)
	e.channel <- action
	return <-action.output
}

// RetrieveBy gets one element from a specific type filtering by other attribute
// Returns the element if found and nil if parameters are incorrect or the element is not found
func (e *phanpy) RetrieveBy(inputType reflect.Type, attribute string, input interface{}) interface{} {
	checkInitialization(e)
	action := newInternalAction(actionRetrieveBy, inputType, attribute, input)
	e.channel <- action
	return <-action.output
}

// RetrieveAll gets all elements with a specific type
// Returns a map with all elements. It will be empty if there are no elements
func (e *phanpy) RetrieveAll(inputType reflect.Type) (map[string]interface{}, error) {
	checkInitialization(e)
	action := newInternalAction(actionRetrieveAll, inputType, nil)
	e.channel <- action
	outputInterface := <-action.output
	if output, ok := outputInterface.(map[string]interface{}); ok {
		return output, nil
	}
	return nil, outputInterface.(error)
}

// Remove deletes one element from the database
// Returns err if the object does not exist
func (e *phanpy) Remove(input interface{}) error {
	checkInitialization(e)
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
func (e *phanpy) RemoveByKey(inputType reflect.Type, key string) error {
	checkInitialization(e)
	action := newInternalAction(actionRemoveByKey, inputType, key)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

// Update modifies an element on the database
func (e *phanpy) Update(input interface{}) error {
	checkInitialization(e)
	action := newInternalAction(actionUpdate, reflect.TypeOf(input), input)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

// Create adds one element to the database
// If the key attribute value is empty (""), a new one will be assigned
func (e *phanpy) Create(input interface{}) (string, error) {
	checkInitialization(e)
	action := newInternalAction(actionCreate, reflect.TypeOf(input), input)
	e.channel <- action
	output := <-action.output
	if reflect.TypeOf(output).Kind() == reflect.String {
		return output.(string), nil
	}
	return "", output.(error)
}

// Exists check if one key is in use in the database
func (e *phanpy) Exists(inputType reflect.Type, key string) bool {
	checkInitialization(e)
	action := newInternalAction(actionExists, inputType, key)
	e.channel <- action
	return (<-action.output).(bool)
}

// ExistsBy gets one element from a specific type filtering by other attribute
// Returns true if found and false if parameters are incorrect or the element is not found
func (e *phanpy) ExistsBy(inputType reflect.Type, attribute string, input interface{}) bool {
	checkInitialization(e)
	action := newInternalAction(actionExistsBy, inputType, attribute, input)
	e.channel <- action
	return (<-action.output).(bool)
}

// NextID gives an empty id to create a new entry
func (e *phanpy) NextID(inputType reflect.Type) string {
	checkInitialization(e)
	action := newInternalAction(actionNextID, inputType)
	e.channel <- action
	return (<-action.output).(string)
}

// close should be called to stop Elephant
func (e *phanpy) close() {
	close(e.channel)
	e.waitgroup.Wait()
}
