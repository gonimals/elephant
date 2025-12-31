package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// CopyInstance Sets src values in the dest instance
func CopyInstance(src, dest any) error {
	//log.Println(reflect.TypeOf(src), reflect.TypeOf(dest))
	if reflect.TypeOf(src) != reflect.TypeOf(dest) || reflect.TypeOf(src).Kind() != reflect.Ptr {
		panic("Cannot copy one instance to other if type differs")
		//return fmt.Errorf("Cannot copy one instance to other if type differs")
	}
	srcValue := reflect.ValueOf(src).Elem()
	destValue := reflect.ValueOf(dest).Elem()
	destValue.Set(srcValue)
	return nil
}

// LoadObjectFromJson creates an instance from JSON bytes and type
func LoadObjectFromJson(objectType reflect.Type, objectString []byte) (any, error) {
	if objectType.Kind() != reflect.Pointer {
		panic("Cannot copy the entire object without pointer")
	}
	instance := reflect.New(objectType).Interface()

	err := json.Unmarshal(objectString, instance)
	if err != nil {
		return nil, fmt.Errorf("value cannot be unmarshalled \"%s\": %v", objectString, err)
	}
	return reflect.ValueOf(instance).Elem().Interface(), nil
}

// CopyEntireObject creates an instance from another instance
func CopyEntireObject(src any) (any, error) {
	objectString, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("elephant: can't convert object to json: %s", src)
	}
	return LoadObjectFromJson(reflect.TypeOf(src), objectString)
}

func CopyMapOfObjects[outputType any](src map[string]any) (map[string]outputType, error) {
	output := map[string]outputType{}
	for key, object := range src {
		objectCopy, err := CopyEntireObject(object)
		if err != nil {
			return nil, err
		}
		output[key] = objectCopy.(outputType)
	}
	return output, nil
}

// CompareInstances returns true if both instance JSONs are equal
func CompareInstances(first, second any) (bool, error) {
	string1, err := json.Marshal(first)
	if err != nil {
		return false, fmt.Errorf("elephant: can't convert object to json: %v", err)
	}
	string2, err := json.Marshal(second)
	if err != nil {
		return false, fmt.Errorf("elephant: can't convert object to json: %v", err)
	}
	return bytes.Equal(string1, string2), nil
}

func BlobsEqual(input1, input2 *[]byte) bool {
	if input1 == input2 {
		return true
	}
	if input1 == nil || input2 == nil {
		return false
	}
	return bytes.Equal(*input1, *input2)
}
