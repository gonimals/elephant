package util

import (
	"bytes"
	"encoding/json"
	"log"
	"reflect"
)

// CopyInstance Sets src values in the dest instance
func CopyInstance(src, dest interface{}) error {
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
func LoadObjectFromJson(objectType reflect.Type, objectString []byte) interface{} {
	if objectType.Kind() != reflect.Ptr {
		panic("Cannot copy the entire object without pointer")
	}
	instance := reflect.New(objectType).Interface()

	err := json.Unmarshal(objectString, instance)
	if err != nil {
		log.Println("Can't unmarshall this value:", objectString)
		log.Fatalln(err)
	}
	return reflect.ValueOf(instance).Elem().Interface()
}

// CopyEntireObject creates an instance from another instance
func CopyEntireObject(src interface{}) interface{} {
	objectString, err := json.Marshal(src)
	if err != nil {
		log.Fatalln("elephant: can't convert object to json:", src)
	}
	return LoadObjectFromJson(reflect.TypeOf(src), objectString)
}

func CopyMapOfObjects(src map[string]interface{}) map[string]interface{} {
	output := map[string]interface{}{}
	for key, object := range src {
		output[key] = CopyEntireObject(object)
	}
	return output
}

// CompareInstances returns true if both instance JSONs are equal
func CompareInstances(first, second interface{}) bool {
	string1, err := json.Marshal(first)
	if err != nil {
		log.Fatalln("elephant: can't convert object to json:", first)
	}
	string2, err := json.Marshal(second)
	if err != nil {
		log.Fatalln("elephant: can't convert object to json:", first)
	}
	return bytes.Equal(string1, string2)
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
