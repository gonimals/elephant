package phanpy

import (
	"reflect"
)

func (e *Phanpy) Retrieve(inputType reflect.Type, key string) interface{} {
	action := newInternalAction(actionRetrieve, inputType, key)
	e.channel <- action
	return <-action.output
}

func (e *Phanpy) RetrieveBy(inputType reflect.Type, attribute string, input interface{}) interface{} {
	checkInitialization(e)
	action := newInternalAction(actionRetrieveBy, inputType, attribute, input)
	e.channel <- action
	return <-action.output
}

func (e *Phanpy) RetrieveAll(inputType reflect.Type) (map[string]interface{}, error) {
	checkInitialization(e)
	action := newInternalAction(actionRetrieveAll, inputType, nil)
	e.channel <- action
	outputInterface := <-action.output
	if output, ok := outputInterface.(map[string]interface{}); ok {
		return output, nil
	}
	return nil, outputInterface.(error)
}

func (e *Phanpy) Remove(input interface{}) error {
	checkInitialization(e)
	action := newInternalAction(actionRemove, reflect.TypeOf(input), input)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

func (e *Phanpy) RemoveByKey(inputType reflect.Type, key string) error {
	checkInitialization(e)
	action := newInternalAction(actionRemoveByKey, inputType, key)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

func (e *Phanpy) Update(input interface{}) error {
	checkInitialization(e)
	action := newInternalAction(actionUpdate, reflect.TypeOf(input), input)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

func (e *Phanpy) Create(input interface{}) (string, error) {
	checkInitialization(e)
	action := newInternalAction(actionCreate, reflect.TypeOf(input), input)
	e.channel <- action
	output := <-action.output
	if reflect.TypeOf(output).Kind() == reflect.String {
		return output.(string), nil
	}
	return "", output.(error)
}

func (e *Phanpy) Exists(inputType reflect.Type, key string) bool {
	checkInitialization(e)
	action := newInternalAction(actionExists, inputType, key)
	e.channel <- action
	return (<-action.output).(bool)
}

func (e *Phanpy) ExistsBy(inputType reflect.Type, attribute string, input interface{}) bool {
	checkInitialization(e)
	action := newInternalAction(actionExistsBy, inputType, attribute, input)
	e.channel <- action
	return (<-action.output).(bool)
}

func (e *Phanpy) NextID(inputType reflect.Type) string {
	checkInitialization(e)
	action := newInternalAction(actionNextID, inputType)
	e.channel <- action
	return (<-action.output).(string)
}

func (e *Phanpy) BlobRetrieve(key string) *[]byte {
	checkInitialization(e)
	action := newInternalAction(actionBlobRetrieve, blobReflectType, key)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(*[]byte)
	}
	return nil
}

func (e *Phanpy) BlobCreate(key string, contents *[]byte) error {
	checkInitialization(e)
	action := newInternalAction(actionBlobCreate, blobReflectType, key, contents)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

func (e *Phanpy) BlobRemove(key string) error {
	checkInitialization(e)
	action := newInternalAction(actionBlobRemove, blobReflectType, key)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

func (e *Phanpy) BlobUpdate(key string, contents *[]byte) error {
	checkInitialization(e)
	action := newInternalAction(actionBlobUpdate, blobReflectType, key, contents)
	e.channel <- action
	output := <-action.output
	if output != nil {
		return output.(error)
	}
	return nil
}

func (e *Phanpy) Close() {
	close(e.channel)
	e.waitgroup.Wait()
}
