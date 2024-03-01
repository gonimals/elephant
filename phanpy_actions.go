package elephant

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
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
	actionBlobCreate
	actionBlobRemove
	actionBlobRetrieve
)

var /*const*/ blobReflectType = reflect.TypeOf(make([]byte, 0))

func (e *phanpy) execManageType(inputType reflect.Type) error {
	if e.managedTypes[inputType] {
		return nil
	}
	// Type is not managed. Start the managing actions
	learntType, err := examineType(inputType)
	if err != nil {
		return err
	}
	e.managedTypes[inputType] = true
	e.learntTypes[inputType] = learntType

	data, err := db.dbRetrieveAll(e.getTableName(inputType))
	if err != nil {
		log.Fatalln("Error reading data from database:", err)
	}
	e.data[inputType] = make(map[string]interface{})
	for key, value := range data {
		e.data[inputType][key] = loadObjectFromJson(inputType, []byte(value))
	}
	return nil
}

func (e *phanpy) execRetrieve(inputType reflect.Type, key string) (output interface{}) {
	return e.data[inputType][key]
}

func (e *phanpy) execRetrieveBy(inputType reflect.Type, attribute string, object interface{}) interface{} {
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

func (e *phanpy) execRetrieveAll(inputType reflect.Type) (output interface{}) {
	return e.data[inputType]
}

func (e *phanpy) execRemove(inputType reflect.Type, input interface{}) error {
	key, err := getKey(input)
	if err != nil {
		return fmt.Errorf("elephant: cannot get id from element")
	}
	return e.execRemoveByKey(inputType, key)
}

func (e *phanpy) execRemoveByKey(inputType reflect.Type, key string) (err error) {
	if !e.execExists(inputType, key) {
		return fmt.Errorf("elephant: there is not element with such id")
	}
	err = db.dbRemove(e.getTableName(inputType), key)
	if err == nil {
		delete(e.data[inputType], key)
	}
	return
}

func (e *phanpy) execCreate(inputType reflect.Type, object interface{}) (output interface{}) {
	key, err := getKey(object)
	if err != nil {
		return err
	} else if key == "" {
		key = e.execNextID(inputType)
		setKey(inputType, object, key)
	} else if e.data[inputType][key] != nil {
		return fmt.Errorf("elephant: trying to create an object with id in use")
	}
	e.data[inputType][key] = copyEntireObject(object)
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

func (e *phanpy) execBlobCreate(key string, contents *[]byte) error {
	if e.data[blobReflectType][key] != nil {
		return fmt.Errorf("elephant: trying to create a blob with id in use")
	}
	e.data[blobReflectType][key] = copyEntireObject(contents)
	if len(contents) > MaxStructLength {
		return fmt.Errorf()
	}
	err = db.dbCreate(e.getTableName(blobReflectType), key, string(objectString))
	if err != nil {
		delete(e.data[blobReflectType], key)
	}
	return key
}

func (e *phanpy) execUpdate(inputType reflect.Type, object interface{}) (err error) {
	key, err := getKey(object)
	if err != nil {
		return
	}
	oldObject, existingObject := e.data[inputType][key]
	if !existingObject {
		return fmt.Errorf("elephant: trying to update unexistent object")
	}
	// e.data[inputType][key] = object
	objectString, err := json.Marshal(object)
	if err != nil {
		log.Fatalln("elephant: can't convert object to json:", object)
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

func (e *phanpy) execExists(inputType reflect.Type, key string) (output bool) {
	_, output = e.data[inputType][key]
	return
}

func (e *phanpy) execExistsBy(inputType reflect.Type, attribute string, object interface{}) bool {
	return e.execRetrieveBy(inputType, attribute, object) != nil
}

func (e *phanpy) execNextID(inputType reflect.Type) string {
	//TODO: Yes, this is not the best way to search
	var outputInt int
	for outputInt = 0; e.data[inputType][strconv.Itoa(outputInt)] != nil; outputInt++ {
	}
	return strconv.Itoa(outputInt)
}

func (e *phanpy) mainRoutine() {
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
		case actionBlobCreate:
			action.output <- e.execBlobCreate(action.inputType, action.object[0])
		case actionBlobRemove:
		case actionBlobRetrieve:
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
