package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
)

// LoadObjectFromJson creates an instance from JSON bytes and type
func LoadObjectFromJson(objectType reflect.Type, objectString []byte) (any, error) {
	if objectType.Kind() != reflect.Pointer {
		Errorf("cannot copy the entire object without pointer")
	}
	instance := reflect.New(objectType).Interface()

	err := json.Unmarshal(objectString, instance)
	if err != nil {
		return nil, Errorf("value cannot be unmarshalled \"%s\": %v", objectString, err)
	}
	return reflect.ValueOf(instance).Elem().Interface(), nil
}

// LoadFromJson creates an instance from JSON bytes and type
func LoadFromJson[objectType any](objectString []byte) (objectType, error) {
	var zero objectType
	t := reflect.TypeFor[objectType]()
	slog.Info(t.Name())
	if t.Kind() != reflect.Pointer {
		return zero, Errorf("should not copy the entire object without pointer")
	}

	instance := reflect.New(t.Elem()).Interface()

	err := json.Unmarshal(objectString, instance)
	if err != nil {
		return zero, Errorf("value cannot be unmarshalled \"%s\": %v", objectString, err)
	}

	return instance.(objectType), nil
}

// CopyEntireObject creates an instance from another instance
func CopyEntireObject(src any) (any, error) {
	objectString, err := json.Marshal(src)
	if err != nil {
		return nil, Errorf("cannot convert object to json \"%s\": %v", src, err)
	}
	return LoadObjectFromJson(reflect.TypeOf(src), objectString)
}

func CopyMapOfObjects[outputType map[string]valueType, valueType any](src map[string]any) (outputType, error) {
	output := outputType{}
	for id, object := range src {
		objectCopy, err := CopyEntireObject(object)
		if err != nil {
			return nil, err
		}
		var ok bool
		output[id], ok = objectCopy.(valueType)
		if !ok {
			return nil, Errorf("error casting map values")
		}
	}
	return output, nil
}

// CompareInstances returns true if both instance JSONs are equal
func CompareInstances(first, second any) (bool, error) {
	string1, err := json.Marshal(first)
	if err != nil {
		return false, Errorf("cannot convert object to json: %v", err)
	}
	string2, err := json.Marshal(second)
	if err != nil {
		return false, Errorf("cannot convert object to json: %v", err)
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

func Errorf(format string, a ...any) error {
	return fmt.Errorf("elephant: "+format, a...)
}
