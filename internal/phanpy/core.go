package phanpy

import (
	"log"
	"reflect"
	"sync"

	"github.com/gonimals/elephant/internal/db"
	"github.com/gonimals/elephant/internal/util"
)

// phanpy is the initial Elephant implementation
type Phanpy struct {
	Context      string
	Data         map[reflect.Type](map[string]interface{})
	learntTypes  map[reflect.Type]*util.LearntType
	channel      chan *internalAction
	waitgroup    sync.WaitGroup
	managedTypes map[reflect.Type]bool
	dbDriver     db.Driver
}

// CheckInitialization
func checkInitialization(e *Phanpy) {
	if e == nil {
		log.Panic("Trying to use an uninitialized instance")
	}
}

func CreatePhanpy(context string, dbDriver db.Driver) *Phanpy {
	p := new(Phanpy)
	p.Context = context
	p.Data = make(map[reflect.Type](map[string]interface{}))
	p.learntTypes = make(map[reflect.Type]*util.LearntType)
	p.channel = make(chan *internalAction)
	p.managedTypes = make(map[reflect.Type]bool)
	p.learntTypes[blobReflectType] = &util.LearntType{
		Name: "blob",
	}
	p.managedTypes[blobReflectType] = true
	p.waitgroup.Add(1)
	p.dbDriver = dbDriver
	go p.mainRoutine()
	return p
}
