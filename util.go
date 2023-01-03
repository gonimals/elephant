package elephant

import (
	"bytes"
	"encoding/json"
	"log"
	"reflect"
)

// copyInstance Sets src values in the dest instance
func copyInstance(src, dest interface{}) error {
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

// loadObjectFromJson creates an instance from JSON bytes and type
func loadObjectFromJson(objectType reflect.Type, objectString []byte) interface{} {
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

// copyEntireObject creates an instance from another instance
func copyEntireObject(src interface{}) interface{} {
	objectString, err := json.Marshal(src)
	if err != nil {
		log.Fatalln("elephant: can't convert object to json:", src)
	}
	return loadObjectFromJson(reflect.TypeOf(src), objectString)
}

// compareInstances returns true if both instance JSONs are equal
func compareInstances(first, second interface{}) bool {
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

// checkInitialization
func checkInitialization(e *Elephant) {
	if e == nil {
		log.Panic("Trying to use an uninitialized instance")
	}
}
