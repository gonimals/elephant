package phanpy

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/gonimals/elephant/internal/util"
)

type internalAction struct {
	code      int
	inputType reflect.Type
	object    []interface{}
	output    chan interface{}
}

// internalAction codes
const (
	actionRetrieve = iota
	actionRetrieveBy
	actionRetrieveAll
	actionUpdate
	actionRemove
	actionRemoveByKey
	actionCreate
	actionExists
	actionExistsBy
	actionNextID
	actionBlobRetrieve
	actionBlobCreate
	actionBlobRemove
	actionBlobUpdate
)

var /*const*/ blobReflectType = reflect.TypeOf(&[]byte{})

func (e *Phanpy) execManageType(inputType reflect.Type) error {
	if e.managedTypes[inputType] {
		return nil
	}
	// Type is not managed. Start the managing actions
	learntType, err := util.ExamineType(inputType)
	if err != nil {
		return err
	}
	e.managedTypes[inputType] = true
	e.learntTypes[inputType] = learntType

	data, err := e.dbDriver.RetrieveAll(e.getTableName(inputType))
	if err != nil {
		log.Fatalln("Error reading data from database:", err)
	}
	e.Data[inputType] = make(map[string]interface{})
	for key, value := range data {
		e.Data[inputType][key] = util.LoadObjectFromJson(inputType, []byte(value))
	}
	return nil
}

func (e *Phanpy) execRetrieve(inputType reflect.Type, key string) (output interface{}) {
	if object, exists := e.Data[inputType][key]; exists {
		return util.CopyEntireObject(object)
	}
	return nil
}

func (e *Phanpy) execRetrieveBy(inputType reflect.Type, attribute string, object interface{}) interface{} {
	//TODO: Yes, this is not the best way to search
	lt := e.learntTypes[inputType]
	filterType := lt.Fields[attribute]
	if filterType == nil || reflect.TypeOf(object) != filterType {
		//log.Println("RetrieveBy executed with invalid arguments:", filterType, reflect.TypeOf(object))
		return nil
	}
	for _, elem := range e.Data[inputType] {
		if object == reflect.ValueOf(elem).Elem().FieldByName(attribute).Interface() {
			return util.CopyEntireObject(elem)
		}
	}
	return nil
}

func (e *Phanpy) execRetrieveAll(inputType reflect.Type) (output interface{}) {
	return util.CopyMapOfObjects(e.Data[inputType])
}

func (e *Phanpy) execBlobRetrieve(key string) (output interface{}) {
	blob, _ := e.dbDriver.BlobRetrieve(key)
	// The error is ignored, as it will probably be norows
	return blob
}

func (e *Phanpy) execRemove(inputType reflect.Type, input interface{}) error {
	key, err := util.GetKey(input)
	if err != nil {
		return fmt.Errorf("elephant: cannot get id from element")
	}
	return e.execRemoveByKey(inputType, key)
}

func (e *Phanpy) execRemoveByKey(inputType reflect.Type, key string) (err error) {
	if !e.execExists(inputType, key) {
		return fmt.Errorf("elephant: there is not element with such id")
	}
	err = e.dbDriver.Remove(e.getTableName(inputType), key)
	if err == nil {
		delete(e.Data[inputType], key)
	}
	return
}

func (e *Phanpy) execBlobRemove(key string) (err error) {
	return e.dbDriver.BlobRemove(key)
}

func (e *Phanpy) execCreate(inputType reflect.Type, object interface{}) (output interface{}) {
	key, err := util.GetKey(object)
	if err != nil {
		return err
	} else if key == "" {
		key = e.execNextID(inputType)
		util.SetKey(inputType, object, key)
	} else if e.Data[inputType][key] != nil {
		return fmt.Errorf("elephant: trying to create an object with id in use")
	}
	e.Data[inputType][key] = util.CopyEntireObject(object)
	objectString, err := json.Marshal(object)
	if err != nil {
		log.Fatalln("elephant: can't convert object to json:", object)
	}
	if len(objectString) > util.MaxStructLength {
		return fmt.Errorf("elephant: serialized object too long to be stored")
	}
	err = e.dbDriver.Create(e.getTableName(inputType), key, string(objectString))
	if err != nil {
		delete(e.Data[inputType], key)
	}
	return key
}

func (e *Phanpy) execBlobCreate(key string, contents *[]byte) error {
	if len(*contents) > util.MaxBlobsLength {
		return fmt.Errorf("elephant: blob too big to be stored")
	}
	return e.dbDriver.BlobCreate(key, contents)
}

func (e *Phanpy) execUpdate(inputType reflect.Type, object interface{}) (err error) {
	key, err := util.GetKey(object)
	if err != nil {
		return
	}
	oldObject, existingObject := e.Data[inputType][key]
	if !existingObject {
		return fmt.Errorf("elephant: trying to update unexistent object")
	}
	objectString, err := json.Marshal(object)
	if err != nil {
		log.Fatalln("elephant: can't convert object to json:", object)
	}
	if len(objectString) > util.MaxStructLength {
		return fmt.Errorf("elephant: serialized object too long to be stored")
	}
	err = e.dbDriver.Update(e.getTableName(inputType), key, string(objectString))
	if err != nil {
		e.Data[inputType][key] = oldObject
	} else {
		e.Data[inputType][key] = util.CopyEntireObject(object)
	}
	return
}

func (e *Phanpy) execBlobUpdate(key string, contents *[]byte) (err error) {
	return e.dbDriver.BlobUpdate(key, contents)
}

func (e *Phanpy) execExists(inputType reflect.Type, key string) (output bool) {
	_, output = e.Data[inputType][key]
	return
}

func (e *Phanpy) execExistsBy(inputType reflect.Type, attribute string, object interface{}) bool {
	return e.execRetrieveBy(inputType, attribute, object) != nil
}

func (e *Phanpy) execNextID(inputType reflect.Type) string {
	//TODO: Yes, this is not the best way to search
	var outputInt int
	for outputInt = 0; e.Data[inputType][strconv.Itoa(outputInt)] != nil; outputInt++ {
	}
	return strconv.Itoa(outputInt)
}

func (e *Phanpy) mainRoutine() {
	for {
		action := <-e.channel
		if action == nil {
			//Received nil action. Shutting down mainRoutine
			break
		}
		err := e.execManageType(action.inputType)
		if err != nil {
			action.output <- err
			continue
		}
		switch action.code {
		case actionRetrieve:
			action.output <- e.execRetrieve(action.inputType, action.object[0].(string))
		case actionRetrieveAll:
			action.output <- e.execRetrieveAll(action.inputType)
		case actionRetrieveBy:
			action.output <- e.execRetrieveBy(action.inputType, action.object[0].(string), action.object[1])
		case actionRemove:
			action.output <- e.execRemove(action.inputType, action.object[0])
		case actionRemoveByKey:
			action.output <- e.execRemoveByKey(action.inputType, action.object[0].(string))
		case actionCreate:
			action.output <- e.execCreate(action.inputType, action.object[0])
		case actionUpdate:
			action.output <- e.execUpdate(action.inputType, action.object[0])
		case actionExists:
			action.output <- e.execExists(action.inputType, action.object[0].(string))
		case actionExistsBy:
			action.output <- e.execExistsBy(action.inputType, action.object[0].(string), action.object[1])
		case actionNextID:
			action.output <- e.execNextID(action.inputType)
		case actionBlobRetrieve:
			action.output <- e.execBlobRetrieve(action.object[0].(string))
		case actionBlobCreate:
			action.output <- e.execBlobCreate(action.object[0].(string), action.object[1].(*[]byte))
		case actionBlobRemove:
			action.output <- e.execBlobRemove(action.object[0].(string))
		case actionBlobUpdate:
			action.output <- e.execBlobUpdate(action.object[0].(string), action.object[1].(*[]byte))
		default:
			action.output <- nil
		}
	}
	e.waitgroup.Done()
}

func newInternalAction(code int, inputType reflect.Type, object ...interface{}) *internalAction {
	return &internalAction{
		code:      code,
		inputType: inputType,
		object:    object,
		output:    make(chan interface{})}
}

func (e *Phanpy) getTableName(inputType reflect.Type) (output string) {
	if e.Context != "" {
		output = e.Context + e.dbDriver.GetContextSymbol()
	}
	typeDescriptor, err := util.ExamineType(inputType)
	if err != nil {
		panic(err)
	}
	output += typeDescriptor.Name
	return
}

// restoreObjectFromDB is meant to be used when an update operation could not be completed
func (e *Phanpy) restoreObjectFromDB(inputType reflect.Type, key string) error {
	data, err := e.dbDriver.Retrieve(inputType.Name(), key)
	if err != nil {
		return err
	}
	e.Data[inputType][key] = util.LoadObjectFromJson(inputType, []byte(data))
	return nil
}
