package elephant

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
)

type internalAction struct {
	code      int
	inputType reflect.Type
	object    []interface{}
	output    chan interface{}
}

//internalAction codes
const (
	actionRetrieve = iota
	actionRetrieveBy
	actionRetrieveAll
	actionUpdate
	actionRemove
	actionCreate
	actionExists
	actionExistsBy
	actionNextID
)

func (e *Elephant) execManageType(inputType reflect.Type) {
	if e.managedTypes[inputType] {
		return
	}
	// Type is not managed. Start the managing actions
	e.managedTypes[inputType] = true
	e.learntTypes[inputType] = examineType(inputType)

	data, err := db.dbRetrieveAll(e.getTableName(inputType))
	if err != nil {
		log.Fatalln("Error reading data from database:", err)
	}
	e.data[inputType] = make(map[int64]interface{})
	for id, value := range data {
		instance := reflect.New(inputType).Interface()
		err := json.Unmarshal([]byte(value), instance)
		if err != nil {
			log.Println("Can't unmarshall this value:", value)
			log.Fatalln(err)
		}
		e.data[inputType][int64(id)] = instance
	}
}

func (e *Elephant) execRetrieve(inputType reflect.Type, key int64) (output interface{}) {
	return e.data[inputType][key]
}

func (e *Elephant) execRetrieveBy(inputType reflect.Type, attribute string, object interface{}) interface{} {
	//TODO: Yes, this is not the best way to search
	lt := e.learntTypes[inputType]
	filterType := lt.fields[attribute]
	if filterType == nil || reflect.TypeOf(object) != filterType {
		//log.Println("RetrieveBy executed with invalid arguments:", filterType, reflect.TypeOf(object))
		return nil
	}
	for _, elem := range e.data[inputType] {
		if object == reflect.ValueOf(elem).Elem().FieldByName(attribute).Interface() {
			return elem
		}
	}
	return nil
}

func (e *Elephant) execRetrieveAll(inputType reflect.Type) (output interface{}) {
	return e.data[inputType]
}

func (e *Elephant) execRemove(inputType reflect.Type, key int64) (err error) {
	if !e.execExists(inputType, key) {
		return fmt.Errorf("elephant: there is not element with such id")
	}
	err = db.dbRemove(e.getTableName(inputType), key)
	if err == nil {
		delete(e.data[inputType], key)
	}
	return
}

func (e *Elephant) execCreate(inputType reflect.Type, object interface{}) (output interface{}) {
	key, err := getKey(inputType, object)
	if err != nil {
		return
	} else if key == 0 {
		key = e.execNextID(inputType)
		setKey(inputType, object, key)
	} else if e.data[inputType][key] != nil {
		return fmt.Errorf("elephant: trying to create an object with id in use")
	}
	e.data[inputType][key] = object
	objectString, err := json.Marshal(object)
	if err != nil {
		log.Fatalln("elephant: can't convert object to json:", object)
	}
	err = db.dbCreate(e.getTableName(inputType), key, string(objectString))
	if err != nil {
		delete(e.data[inputType], key)
	}
	return key
}

func (e *Elephant) execUpdate(inputType reflect.Type, object interface{}) (err error) {
	key, err := getKey(inputType, object)
	if err != nil {
		return
	}
	oldObject := e.data[inputType][key]
	// e.data[inputType][key] = object
	objectString, err := json.Marshal(object)
	if err != nil {
		log.Fatalln("Can't convert object to json:", object)
	}
	err = db.dbUpdate(e.getTableName(inputType), key, string(objectString))
	if err != nil {
		e.data[inputType][key] = oldObject
	} else {
		err := copyInstance(object, oldObject)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return
}

func (e *Elephant) execExists(inputType reflect.Type, key int64) (output bool) {
	_, output = e.data[inputType][key]
	return
}

func (e *Elephant) execExistsBy(inputType reflect.Type, attribute string, object interface{}) bool {
	return e.execRetrieveBy(inputType, attribute, object) == nil
}

func (e *Elephant) execNextID(inputType reflect.Type) (output int64) {
	//TODO: Yes, this is not the best way to search
	for output = 0; e.data[inputType][output] != nil; output++ {
	}
	return
}

func (e *Elephant) mainRoutine() {
	for true {
		action := <-e.channel
		if action == nil {
			//Received nil action. Shutting down mainRoutine
			break
		}
		e.execManageType(action.inputType)
		switch action.code {
		case actionRetrieve:
			action.output <- e.execRetrieve(action.inputType, action.object[0].(int64))
		case actionRetrieveAll:
			action.output <- e.execRetrieveAll(action.inputType)
		case actionRetrieveBy:
			action.output <- e.execRetrieveBy(action.inputType, action.object[0].(string), action.object[1])
		case actionRemove:
			action.output <- e.execRemove(action.inputType, action.object[0].(int64))
		case actionCreate:
			action.output <- e.execCreate(action.inputType, action.object[0])
		case actionUpdate:
			action.output <- e.execUpdate(action.inputType, action.object[0])
		case actionExists:
			action.output <- e.execExists(action.inputType, action.object[0].(int64))
		case actionExistsBy:
			action.output <- e.execExistsBy(action.inputType, action.object[0].(string), action.object[1])
		case actionNextID:
			action.output <- e.execNextID(action.inputType)
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
