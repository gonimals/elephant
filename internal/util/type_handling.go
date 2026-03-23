package util

import (
	"reflect"
)

// MaxStructLength defines how long can be a structure converted to JSON to be stored
const MaxStructLength = 65535 //64k

// MaxBlobsLength defines how big can be blobs stored
const MaxBlobsLength = 65535 //64k

// Structs
type LearntType struct {
	Name    string
	Id      string //only needed if struct will be a db table
	Fields  map[string]reflect.Type
	Updates map[string]struct{}
}

var LearntTypes = map[reflect.Type]*LearntType{}

// examineType will check that the type can be transformed into JSON and has an Id parameter
func ExamineType(input reflect.Type) (output *LearntType, err error) {
	if input.Kind() != reflect.Ptr || input.Elem().Kind() != reflect.Struct {
		return nil, Errorf("%s is not a pointer to struct. Kind: %s",
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

	for i := 0; i < input.NumField(); i++ {
		field := input.Field(i)
		output.Fields[field.Name] = field.Type
		if field.Tag != "" {
			tag := field.Tag.Get("db")
			if tag == "id" {
				output.Id = field.Name
				if field.Type.Kind() != reflect.String {
					return nil, Errorf("%s has a parameter with the annotation `db:\"id\"` which is not a string",
						input.String())
				}
			} else if tag == "update" {
				output.Updates[field.Name] = struct{}{}
			}
		}
	}
	if output.Id == "" {
		return nil, Errorf("%s hasn't got an string parameter with the annotation `db:\"id\"`",
			input.String())
	}
	LearntTypes[input] = output
	return
}

// Returns the id from the input ptr object
func GetId(input any) (output string, err error) {
	typeDescriptor, err := ExamineType(reflect.TypeOf(input))
	if err != nil {
		return "", err
	}
	if reflect.ValueOf(input).IsNil() {
		return "", Errorf("no nil id creation allowed")
	}
	output = reflect.ValueOf(input).Elem().FieldByName(typeDescriptor.Id).String()
	return
}

func SetId(inputType reflect.Type, input any, id string) {
	typeDescriptor, err := ExamineType(inputType)
	if err != nil {
		panic(err)
	}
	reflect.ValueOf(input).Elem().FieldByName(typeDescriptor.Id).SetString(id)
}

func IsNilable[T any]() bool {
	t := reflect.TypeFor[T]()
	k := t.Kind()

	switch k {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	default:
		return false
	}
}
