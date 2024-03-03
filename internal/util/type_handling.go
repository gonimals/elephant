package util

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

const ContextSymbol = "."

// MaxStructLength defines how long can be a structure converted to JSON to be stored
const MaxStructLength = 65535 //64k

// MaxBlobsLength defines how big can be blobs stored
const MaxBlobsLength = 65535 //64k

// Structs
type LearntType struct {
	Name    string
	Key     string //only needed if struct will be a db table
	Fields  map[string]reflect.Type
	Updates map[string]struct{}
}

var LearntTypes = map[reflect.Type]*LearntType{}

// examineType will check that the type can be transformed into JSON and has an Id parameter
func ExamineType(input reflect.Type) (output *LearntType, err error) {
	if input.Kind() != reflect.Ptr || input.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is not a pointer to struct. Kind: %s",
			input.String(), input.Kind().String())
	}
	input = input.Elem()
	output = LearntTypes[input]
	if output != nil {
		return // type was already processed
	}
	output = new(LearntType)
	output.Fields = make(map[string]reflect.Type)
	output.Updates = make(map[string]struct{})
	output.Name = input.Name()
	if strings.Contains(output.Name, ContextSymbol) {
		log.Fatalln("Type name " + output.Name + " contains the character " + ContextSymbol + " which is also used as context symbol. This should never happen.")
	}

	for i := 0; i < input.NumField(); i++ {
		field := input.Field(i)
		output.Fields[field.Name] = field.Type
		if field.Tag != "" {
			tag := field.Tag.Get("db")
			if tag == "key" {
				output.Key = field.Name
				if field.Type.Kind() != reflect.String {
					return nil, fmt.Errorf("%s has a parameter with the annotation `db:\"key\"` which is not a string",
						input.String())
				}
			} else if tag == "update" {
				output.Updates[field.Name] = struct{}{}
			}
		}
	}
	if output.Key == "" {
		return nil, fmt.Errorf("%s hasn't got an string parameter with the annotation `db:\"key\"`",
			input.String())
	}
	LearntTypes[input] = output
	return
}

// Returns the key from the input ptr object
func GetKey(input interface{}) (output string, err error) {
	typeDescriptor, err := ExamineType(reflect.TypeOf(input))
	if err != nil {
		return "", err
	}
	if reflect.ValueOf(input).IsNil() {
		return "", fmt.Errorf("no nil key creation allowed")
	}
	output = reflect.ValueOf(input).Elem().FieldByName(typeDescriptor.Key).String()
	return
}

func SetKey(inputType reflect.Type, input interface{}, key string) {
	typeDescriptor, err := ExamineType(inputType)
	if err != nil {
		panic(err)
	}
	reflect.ValueOf(input).Elem().FieldByName(typeDescriptor.Key).SetString(key)
}
