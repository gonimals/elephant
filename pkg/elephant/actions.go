package elephant

import (
	"encoding/json"
	"reflect"

	"github.com/gonimals/elephant/internal/util"
	"github.com/google/uuid"
)

type internalAction struct {
	code      int
	inputType reflect.Type
	object    []any
	output    chan actionOutput
}

type actionOutput struct {
	data any
	err  error
}

// internalAction codes
const (
	actionRetrieve = iota
	actionRetrieveBy
	actionRetrieveAll
	actionUpdate
	actionUpsert
	actionRemove
	actionRemoveById
	actionCreate
	actionExists
	actionExistsBy
	actionNextID
	actionBlobRetrieve
	actionBlobCreate
	actionBlobRemove
	actionBlobUpdate
	actionBlobUpsert
	actionBlobExists
)

var /*const*/ blobReflectType = reflect.TypeOf(&[]byte{})

func mainRoutine() {
	for {
		action := <-channel
		if action == nil {
			//Received nil action. Shutting down mainRoutine
			break
		}
		err := execManageType(action.inputType)
		if err != nil {
			action.output <- actionOutput{nil, err}
			continue
		}
		output := actionOutput{}
		switch action.code {
		case actionRetrieve:
			output.data, output.err = execRetrieve(action.inputType, action.object[0].(string))
		case actionRetrieveAll:
			output.data = execRetrieveAll(action.inputType)
		case actionRetrieveBy:
			output.data, output.err = execRetrieveBy(action.inputType, action.object[0].(string), action.object[1])
		case actionRemove:
			output.err = execRemove(action.inputType, action.object[0])
		case actionRemoveById:
			output.err = execRemoveById(action.inputType, action.object[0].(string))
		case actionCreate:
			output.data, output.err = execUpsert(action.inputType, action.object[0], false, true)
		case actionUpdate:
			output.data, output.err = execUpsert(action.inputType, action.object[0], true, false)
		case actionUpsert:
			output.data, output.err = execUpsert(action.inputType, action.object[0], true, true)
		case actionExists:
			output.data = execExists(action.inputType, action.object[0].(string))
		case actionExistsBy:
			output.data, output.err = execExistsBy(action.inputType, action.object[0].(string), action.object[1])
		case actionNextID:
			output.data, output.err = execNewID(action.inputType)
		case actionBlobRetrieve:
			output.data, output.err = execBlobRetrieve(action.object[0].(string))
		case actionBlobCreate:
			output.err = execBlobUpsert(action.object[0].(string), action.object[1].(*[]byte), false, true)
		case actionBlobRemove:
			output.err = execBlobRemove(action.object[0].(string))
		case actionBlobUpdate:
			output.err = execBlobUpsert(action.object[0].(string), action.object[1].(*[]byte), true, false)
		case actionBlobUpsert:
			output.err = execBlobUpsert(action.object[0].(string), action.object[1].(*[]byte), true, true)
		case actionBlobExists:
			output.data, output.err = dbDriver.BlobExists(action.object[0].(string))
		default:
			output.err = util.Errorf("unknown action")
		}
		action.output <- output
	}
	waitgroup.Done()
}

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
		return util.Errorf("error reading data from database: %v", err)
	}
	data[inputType] = make(map[string]any)
	var loadErrors []error
	for id, value := range retrieved {
		valueObject, err := util.LoadObjectFromJson(inputType, []byte(value))
		if err != nil {
			loadErrors = append(loadErrors, err)
			continue
		}
		data[inputType][id] = valueObject
	}
	if len(loadErrors) > 0 {
		return util.Errorf("error loading data from database: %v", loadErrors)
	}
	return nil
}

func execRetrieve(inputType reflect.Type, id string) (output any, err error) {
	if object, exists := data[inputType][id]; exists {
		output, err := util.CopyEntireObject(object)
		if err != nil {
			return nil, err
		}
		return output, nil
	}
	return nil, nil
}

func execRetrieveBy(inputType reflect.Type, attribute string, object any) (output any, err error) {
	//TODO: Yes, this is not the best way to search
	lt := learntTypes[inputType]
	filterType := lt.Fields[attribute]
	if filterType == nil || reflect.TypeOf(object) != filterType {
		//log.Println("RetrieveBy executed with invalid arguments:", filterType, reflect.TypeOf(object))
		return nil, util.Errorf("cannot retrieve by attribute named %s with type %v: filter type is %v", attribute, reflect.TypeOf(object), filterType)
	}
	for _, elem := range data[inputType] {
		if object == reflect.ValueOf(elem).Elem().FieldByName(attribute).Interface() {
			output, err := util.CopyEntireObject(elem)
			if err != nil {
				return nil, err
			}
			return output, nil
		}
	}
	return nil, nil
}

func execRetrieveAll(inputType reflect.Type) (output any) {
	return data[inputType]
}

func execBlobRetrieve(id string) (any, error) {
	blob, _ := dbDriver.BlobRetrieve(id)
	// TODO: The error is ignored, as it will probably be norows
	return blob, nil
}

func execRemove(inputType reflect.Type, input any) error {
	id, err := util.GetId(input)
	if err != nil {
		return util.Errorf("cannot get id from element")
	}
	return execRemoveById(inputType, id)
}

func execRemoveById(inputType reflect.Type, id string) (err error) {
	if !execExists(inputType, id) {
		return util.Errorf("there is not element with such id")
	}
	err = dbDriver.Remove(getTableName(inputType), id)
	if err == nil {
		delete(data[inputType], id)
	}
	return
}

func execBlobRemove(id string) (err error) {
	return dbDriver.BlobRemove(id)
}

func execUpsert(inputType reflect.Type, object any, allowUpdate bool, allowCreate bool) (output any, err error) {
	if !allowUpdate && !allowCreate {
		return nil, util.Errorf("execUpsert internal error: no valid modes")
	}
	id, err := util.GetId(object)
	if err != nil {
		return nil, err
	}
	var existingObject bool
	var oldObject any
	if id != "" {
		oldObject, existingObject = data[inputType][id]
		if existingObject && !allowUpdate {
			return nil, util.Errorf("trying to create an object with id in use")
		} else if !existingObject && !allowCreate {
			return nil, util.Errorf("trying to update unexistent object")
		}
	} else {
		if !allowCreate {
			return nil, util.Errorf("trying to update object without id")
		}
		id, err = execNewID(inputType)
		if err != nil {
			return nil, err
		}
		util.SetId(inputType, object, id)
	}
	objectString, err := json.Marshal(object)
	if err != nil {
		return nil, util.Errorf("cannot convert object to json: %s error: %v", object, err)
	}
	if len(objectString) > util.MaxStructLength {
		return nil, util.Errorf("serialized object too long to be stored")
	}
	data[inputType][id], err = util.CopyEntireObject(object)
	if err != nil {
		return nil, err
	}
	if existingObject {
		err = dbDriver.Update(getTableName(inputType), id, string(objectString))
		if err != nil {
			data[inputType][id] = oldObject
			return nil, err
		}
	} else {
		err = dbDriver.Create(getTableName(inputType), id, string(objectString))
		if err != nil {
			delete(data[inputType], id)
			return nil, err
		}
	}
	return id, nil
}

func execBlobUpsert(id string, contents *[]byte, allowUpdate bool, allowCreate bool) (err error) {
	if len(*contents) > util.MaxBlobsLength {
		return util.Errorf("blob too big to be stored")
	}
	blobExists, err := dbDriver.BlobExists(id)
	if err != nil {
		return util.Errorf("cannot determine if blob exists: %v", err)
	}
	if blobExists {
		if !allowUpdate {
			return util.Errorf("trying to create an existing blob")
		}
		return dbDriver.BlobUpdate(id, contents)
	} else {
		if !allowCreate {
			return util.Errorf("trying to update an unexistent blob")
		}
		return dbDriver.BlobCreate(id, contents)
	}
}

func execExists(inputType reflect.Type, id string) (output bool) {
	_, output = data[inputType][id]
	return
}

func execExistsBy(inputType reflect.Type, attribute string, object any) (bool, error) {
	output, err := execRetrieveBy(inputType, attribute, object)
	return output != nil, err
}

func execNewID(inputType reflect.Type) (string, error) {
	for i := 0; i < 100; i++ {
		id := uuid.New().String()
		if !execExists(inputType, id) {
			return id, nil
		}
	}
	return "", util.Errorf("cannot generate a unique UUID after 100 attempts")
}

func newInternalAction(code int, inputType reflect.Type, object ...any) *internalAction {
	return &internalAction{
		code:      code,
		inputType: inputType,
		object:    object,
		output:    make(chan actionOutput, 1)}
}

func getTableName(inputType reflect.Type) (output string) {
	typeDescriptor, err := util.ExamineType(inputType)
	if err != nil {
		panic(err)
	}
	output += typeDescriptor.Name
	return
}
