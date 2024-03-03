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

func (e *phanpy) Retrieve(inputType reflect.Type, key string) interface{} {
	action := newInternalAction(actionRetrieve, inputType, key)
	e.channel <- action
	return <-action.output
}

func (e *phanpy) RetrieveBy(inputType reflect.Type, attribute string, input interface{}) interface{} {
	checkInitialization(e)
	action := newInternalAction(actionRetrieveBy, inputType, attribute, input)
	e.channel <- action
	return <-action.output
}

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

func (e *phanpy) Exists(inputType reflect.Type, key string) bool {
	checkInitialization(e)
	action := newInternalAction(actionExists, inputType, key)
	e.channel <- action
	return (<-action.output).(bool)
}

func (e *phanpy) ExistsBy(inputType reflect.Type, attribute string, input interface{}) bool {
	checkInitialization(e)
	action := newInternalAction(actionExistsBy, inputType, attribute, input)
	e.channel <- action
	return (<-action.output).(bool)
}

func (e *phanpy) NextID(inputType reflect.Type) string {
	checkInitialization(e)
	action := newInternalAction(actionNextID, inputType)
	e.channel <- action
	return (<-action.output).(string)
}

func (e *phanpy) BlobRetrieve(key string) *[]byte {
	checkInitialization(e)
	action := newInternalAction(actionBlobRetrieve, blobReflectType, key)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(*[]byte)
	}
	return nil
}

func (e *phanpy) BlobCreate(key string, contents *[]byte) error {
	checkInitialization(e)
	action := newInternalAction(actionBlobCreate, blobReflectType, key, contents)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

func (e *phanpy) BlobRemove(key string) error {
	checkInitialization(e)
	action := newInternalAction(actionBlobRemove, blobReflectType, key)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

func (e *phanpy) BlobUpdate(key string, contents *[]byte) error {
	checkInitialization(e)
	action := newInternalAction(actionBlobUpdate, blobReflectType, key, contents)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

func (e *phanpy) close() {
	close(e.channel)
	e.waitgroup.Wait()
}
