package elephant

import (
	"os"
	"reflect"
	"testing"
)

func TestCreateFile(t *testing.T) {
	os.Remove(temporaryDB)
	err := Initialize("sqlite3://" + temporaryDB)
	if err != nil {
		t.Error("Initialization failed", err)
	}
	defer Close()
	fileCheckType := reflect.TypeOf((*fileStructCheck)(nil))

	var fileExample = []byte{0x00, 0x01, 0x02, 0x03}

	fileCapsule := fileStructCheck{
		Filename: "test",
		Contents: fileExample,
	}

	if _, err = MainContext.Create(&fileCapsule); err != nil {
		t.Error("Creation failed:", err)
	}
	if !compareInstances(MainContext.Retrieve(fileCheckType, fileCapsule.Filename), fileCapsule) {
		t.Error("Retrieved instance and the original one should be equal")
	}

}
