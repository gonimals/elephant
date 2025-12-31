package elephant

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
	object    []any
	output    chan any
}

// internalAction codes
const (
	actionRetrieve = iota
	actionRetrieveBy
	actionRetrieveAll
	actionUpdate
	actionUpsert
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

func execManageType(inputType reflect.Type) error {
	if managedTypes[inputType] {
		return nil
	}
	// Type is not managed. Start the managing actions
	learntType, err := util.ExamineType(inputType)
	if err != nil {
		return err
	}
	managedTypes[inputType] = true
	learntTypes[inputType] = learntType

	retrieved, err := dbDriver.RetrieveAll(getTableName(inputType))
	if err != nil {
		return fmt.Errorf("error reading data from database: %v", err)
	}
	data[inputType] = make(map[string]any)
	var loadErrors []error
	for key, value := range retrieved {
		valueObject, err := util.LoadObjectFromJson(inputType, []byte(value))
		if err != nil {
			loadErrors = append(loadErrors, err)
			continue
		}
		data[inputType][key] = valueObject
	}
	if len(loadErrors) > 0 {
		return fmt.Errorf("error loading data from database: %v", loadErrors)
	}
	return nil
}

func execRetrieve(inputType reflect.Type, key string) (output any) {
	if object, exists := data[inputType][key]; exists {
		output, err := util.CopyEntireObject(object)
		if err != nil {
			return err
		}
		return output
	}
	return nil
}

func execRetrieveBy(inputType reflect.Type, attribute string, object any) (output any) {
	//TODO: Yes, this is not the best way to search
	lt := learntTypes[inputType]
	filterType := lt.Fields[attribute]
	if filterType == nil || reflect.TypeOf(object) != filterType {
		//log.Println("RetrieveBy executed with invalid arguments:", filterType, reflect.TypeOf(object))
		return fmt.Errorf("cannot retrieve by attribute named %s with type %v: filter type is %v", attribute, reflect.TypeOf(object), filterType)
	}
	for _, elem := range data[inputType] {
		if object == reflect.ValueOf(elem).Elem().FieldByName(attribute).Interface() {
			output, err := util.CopyEntireObject(elem)
			if err != nil {
				return err
			}
			return output
		}
	}
	return nil
}

func execRetrieveAll(inputType reflect.Type) (output any) {
	return data[inputType]
}

func execBlobRetrieve(key string) (output any) {
	blob, _ := dbDriver.BlobRetrieve(key)
	// The error is ignored, as it will probably be norows
	return blob
}

func execRemove(inputType reflect.Type, input any) error {
	key, err := util.GetKey(input)
	if err != nil {
		return fmt.Errorf("elephant: cannot get id from element")
	}
	return execRemoveByKey(inputType, key)
}

func execRemoveByKey(inputType reflect.Type, key string) (err error) {
	if !execExists(inputType, key) {
		return fmt.Errorf("elephant: there is not element with such id")
	}
	err = dbDriver.Remove(getTableName(inputType), key)
	if err == nil {
		delete(data[inputType], key)
	}
	return
}

func execBlobRemove(key string) (err error) {
	return dbDriver.BlobRemove(key)
}

func execCreate(inputType reflect.Type, object any) (output any) {
	key, err := util.GetKey(object)
	if err != nil {
		return err
	} else if key == "" {
		key = execNextID(inputType)
		util.SetKey(inputType, object, key)
	} else if data[inputType][key] != nil {
		return fmt.Errorf("elephant: trying to create an object with id in use")
	}
	data[inputType][key], err = util.CopyEntireObject(object)
	if err != nil {
		return err
	}
	objectString, err := json.Marshal(object)
	if err != nil {
		log.Fatalln("elephant: can't convert object to json:", object)
	}
	if len(objectString) > util.MaxStructLength {
		return fmt.Errorf("elephant: serialized object too long to be stored")
	}
	err = dbDriver.Create(getTableName(inputType), key, string(objectString))
	if err != nil {
		delete(data[inputType], key)
	}
	return key
}

func execBlobCreate(key string, contents *[]byte) error {
	if len(*contents) > util.MaxBlobsLength {
		return fmt.Errorf("elephant: blob too big to be stored")
	}
	return dbDriver.BlobCreate(key, contents)
}

func execUpdate(inputType reflect.Type, object any) (err error) {
	key, err := util.GetKey(object)
	if err != nil {
		return
	}
	oldObject, existingObject := data[inputType][key]
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
	err = dbDriver.Update(getTableName(inputType), key, string(objectString))
	if err != nil {
		data[inputType][key] = oldObject
	} else {
		data[inputType][key], err = util.CopyEntireObject(object)
		if err != nil {
			return err
		}
	}
	return
}

func execBlobUpdate(key string, contents *[]byte) (err error) {
	return dbDriver.BlobUpdate(key, contents)
}

func execUpsert(inputType reflect.Type, object any) (output any) {
	key, err := util.GetKey(object)
	if err != nil {
		return err
	}
	var existingObject bool
	var oldObject any
	if key != "" {
		oldObject, existingObject = data[inputType][key]
	} else {
		key = execNextID(inputType)
		util.SetKey(inputType, object, key)
	}
	objectString, err := json.Marshal(object)
	if err != nil {
		log.Fatalln("elephant: can't convert object to json:", object)
	}
	if len(objectString) > util.MaxStructLength {
		return fmt.Errorf("elephant: serialized object too long to be stored")
	}
	data[inputType][key], err = util.CopyEntireObject(object)
	if err != nil {
		return err
	}
	if existingObject {
		err = dbDriver.Update(getTableName(inputType), key, string(objectString))
		if err != nil {
			data[inputType][key] = oldObject
			return err
		}
	} else {
		err = dbDriver.Create(getTableName(inputType), key, string(objectString))
		if err != nil {
			delete(data[inputType], key)
			return err
		}
	}
	return key
}

func execExists(inputType reflect.Type, key string) (output bool) {
	_, output = data[inputType][key]
	return
}

func execExistsBy(inputType reflect.Type, attribute string, object any) (bool, error) {
	output := execRetrieveBy(inputType, attribute, object)
	switch v := output.(type) {
	case error:
		return false, v
	default:
		return v != nil, nil
	}
}

func execNextID(inputType reflect.Type) string {
	//TODO: Yes, this is not the best way to search
	var outputInt int
	for outputInt = 0; data[inputType][strconv.Itoa(outputInt)] != nil; outputInt++ {
	}
	return strconv.Itoa(outputInt)
}

func mainRoutine() {
	for {
		action := <-channel
		if action == nil {
			//Received nil action. Shutting down mainRoutine
			break
		}
		err := execManageType(action.inputType)
		if err != nil {
			action.output <- err
			continue
		}
		switch action.code {
		case actionRetrieve:
			action.output <- execRetrieve(action.inputType, action.object[0].(string))
		case actionRetrieveAll:
			action.output <- execRetrieveAll(action.inputType)
		case actionRetrieveBy:
			action.output <- execRetrieveBy(action.inputType, action.object[0].(string), action.object[1])
		case actionRemove:
			action.output <- execRemove(action.inputType, action.object[0])
		case actionRemoveByKey:
			action.output <- execRemoveByKey(action.inputType, action.object[0].(string))
		case actionCreate:
			action.output <- execCreate(action.inputType, action.object[0])
		case actionUpdate:
			action.output <- execUpdate(action.inputType, action.object[0])
		case actionUpsert:
			action.output <- execUpsert(action.inputType, action.object[0])
		case actionExists:
			action.output <- execExists(action.inputType, action.object[0].(string))
		case actionExistsBy:
			output, err := execExistsBy(action.inputType, action.object[0].(string), action.object[1])
			if err != nil {
				action.output <- err
			} else {
				action.output <- output
			}
		case actionNextID:
			action.output <- execNextID(action.inputType)
		case actionBlobRetrieve:
			action.output <- execBlobRetrieve(action.object[0].(string))
		case actionBlobCreate:
			action.output <- execBlobCreate(action.object[0].(string), action.object[1].(*[]byte))
		case actionBlobRemove:
			action.output <- execBlobRemove(action.object[0].(string))
		case actionBlobUpdate:
			action.output <- execBlobUpdate(action.object[0].(string), action.object[1].(*[]byte))
		default:
			action.output <- nil
		}
	}
	waitgroup.Done()
}

func newInternalAction(code int, inputType reflect.Type, object ...any) *internalAction {
	return &internalAction{
		code:      code,
		inputType: inputType,
		object:    object,
		output:    make(chan any)}
}

func getTableName(inputType reflect.Type) (output string) {
	typeDescriptor, err := util.ExamineType(inputType)
	if err != nil {
		panic(err)
	}
	output += typeDescriptor.Name
	return
}
