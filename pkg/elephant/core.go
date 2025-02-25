package elephant

import (
	"log"
	"reflect"
	"sync"

	"github.com/gonimals/elephant/internal/db"
	"github.com/gonimals/elephant/internal/util"
)

var (
	data         map[reflect.Type](map[string]interface{})
	learntTypes  map[reflect.Type]*util.LearntType
	channel      chan *internalAction
	waitgroup    sync.WaitGroup
	managedTypes map[reflect.Type]bool
	dbDriver     db.Driver
)

func checkInitialization() {
	if dbDriver == nil {
		log.Panic("Trying to use an uninitialized instance")
	}
}
