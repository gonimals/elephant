package elephant

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

const ContextSymbol = "."

// Structs
type learntType struct {
	name    string
	key     string //only needed if struct will be a db table
	fields  map[string]reflect.Type
	updates map[string]struct{}
}

type dbDriver interface {
	dbClose()
	dbRetrieve(inputType string, key int64) (output string, err error)
	dbRetrieveAll(inputType string) (output map[int]string, err error)
	dbCreate(inputType string, key int64, input string) (err error)
	dbUpdate(inputType string, key int64, input string) (err error)
	dbRemove(inputType string, key int64) (err error)
}

var currentElephants map[string]*Elephant
var learntTypes map[reflect.Type]*learntType

//copyInstance Sets src values in the dest instance
func copyInstance(src, dest interface{}) error {
	if reflect.TypeOf(src) != reflect.TypeOf(dest) || reflect.TypeOf(src).Kind() != reflect.Ptr {
		return fmt.Errorf("Cannot copy one instance to other if type differs")
	}
	srcValue := reflect.ValueOf(src).Elem()
	destValue := reflect.ValueOf(dest).Elem()
	destValue.Set(srcValue)
	return nil
}

//examineType will check that the type can be transformed into JSON and has an Id parameter
func examineType(input reflect.Type) (output *learntType) {
	if input.Kind() != reflect.Ptr || input.Elem().Kind() != reflect.Struct {
		panic(errors.New(input.String() + " is not a pointer to struct"))
	}
	input = input.Elem()
	output = learntTypes[input]
	if output != nil {
		return // type was already processed
	}
	output = new(learntType)
	output.fields = make(map[string]reflect.Type)
	output.updates = make(map[string]struct{})
	output.name = input.Name()
	if strings.Contains(output.name, ContextSymbol) {
		log.Fatalln("Type name " + output.name + " contains the character " + ContextSymbol + " which is also used as context symbol. This should never happen.")
	}

	for i := 0; i < input.NumField(); i++ {
		field := input.Field(i)
		output.fields[field.Name] = field.Type
		if field.Tag != "" {
			tag := field.Tag.Get("db")
			if tag == "key" {
				output.key = field.Name
			} else if tag == "update" {
				output.updates[field.Name] = struct{}{}
			}
		}
	}
	learntTypes[input] = output
	return
}

// Returns the key from the input ptr object
func getKey(inputType reflect.Type, input interface{}) (output int64, err error) {
	typeDescriptor := examineType(inputType)
	if reflect.ValueOf(input).IsNil() {
		return 0, errors.New("No nil key creation allowed")
	}
	output = reflect.ValueOf(input).Elem().FieldByName(typeDescriptor.key).Int()
	return
}

func setKey(inputType reflect.Type, input interface{}, key int64) {
	typeDescriptor := examineType(inputType)
	reflect.ValueOf(input).Elem().FieldByName(typeDescriptor.key).SetInt(key)
}

func (e *Elephant) getTableName(inputType reflect.Type) (output string) {
	if e.Context != "" {
		output = e.Context + ContextSymbol
	}
	output += examineType(inputType).name
	return
}
