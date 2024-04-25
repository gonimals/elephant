package sql

import (
	"log"
	"strings"
	"testing"

	"github.com/gonimals/elephant/internal/util"
)

func testDriver(t *testing.T, db *driver) {
	testBasicUsage(t, db)
	testLimits(t, db)
}

func testBasicUsage(t *testing.T, db *driver) {
	err := db.Create("test_table", "1", "{ \"fdsa\": [] }")
	if err != nil {
		t.Error("simple create operation fails:", err)
	}
	output, err := db.Retrieve("test_table", "1")
	if err != nil {
		t.Error("simple retrieve operation fails:", err)
	}
	if output != "{ \"fdsa\": [] }" {
		t.Error("retrieved string is not the original")
	}
	err = db.Update("test_table", "1", "{ \"asdf\": 2 }")
	if err != nil {
		t.Error("simple update operation fails:", err)
	}
	err = db.Update("test_table", "1", "[]")
	if err != nil {
		t.Error("simple update operation fails:", err)
	}
	output, err = db.Retrieve("test_table", "1")
	if err != nil {
		t.Error("simple retrieve operation fails:", err)
	}
	if output != "[]" {
		t.Errorf("retrieved string is not the updated one")
	}
	err = db.Remove("test_table", "1")
	if err != nil {
		t.Error("simple delete operation fails:", err)
	}
	output, err = db.Retrieve("test_table", "1")
	if err == nil {
		t.Error("retrieve operation of deleted item doesn't give error")
	} else if output != "" {
		t.Error("retrieve operation of deleted item gives output:", output)
	}
	err = db.BlobCreate("1", &[]byte{0x00})
	if err != nil {
		t.Error("blob create operation fails:", err)
	}
	blob, err := db.BlobRetrieve("1")
	if err != nil {
		t.Error("blob retrieve operation fails:", err)
	}
	if !util.BlobsEqual(blob, &[]byte{0x00}) {
		t.Error("retrieved blob is not the original")
	}
	err = db.BlobUpdate("1", &[]byte{0x01})
	if err != nil {
		t.Error("blob update operation fails:", err)
	}
	blob, err = db.BlobRetrieve("1")
	if err != nil {
		t.Error("blob retrieve operation fails:", err)
	}
	if !util.BlobsEqual(blob, &[]byte{0x01}) {
		t.Error("retrieved blob is not the updated one")
	}
	err = db.BlobRemove("1")
	if err != nil {
		t.Error("blob delete operation fails:", err)
	}
	err = db.BlobRemove("1")
	if err == nil {
		t.Error("blob delete operation should fail")
	}
	blob, err = db.BlobRetrieve("1")
	if err == nil {
		t.Error("retrieve operation of deleted blob doesn't give error")
	} else if util.BlobsEqual(blob, &[]byte{0x00}) {
		t.Error("retrieve operation of deleted blob gives output:", blob)
	}
}

func testLimits(t *testing.T, db *driver) {
	longString := strings.Repeat("A", MaxKeyLength+1)
	err := db.Create("limits", longString, "asdf")
	if err == nil {
		t.Error("too long key should fail")
	}
	err = db.BlobCreate(longString, &[]byte{0x00})
	if err == nil {
		t.Error("too long key should fail")
	}
	longString = strings.Repeat("A", 1024*1024*1024) //1GB
	log.Println("A very big query will be executed. A database driver error is acceptable right now")
	err = db.Create("limits", "longvalue", longString)
	if err == nil {
		t.Error("too long value should fail")
	}
	log.Println("The big query has finished. No more errors until the next big query are acceptable")

	err = db.BlobRemove("longvalue")
	if err == nil {
		t.Error("long value delete operation should fail")
	}
	longString = ""

	longBlob := []byte(strings.Repeat("A", 1024*1024*1024)) //1GB
	log.Println("A very big query will be executed. A database driver error is acceptable right now")
	err = db.BlobCreate("longvalue", &longBlob)
	if err == nil {
		t.Error("too long blob should fail")
	}
	log.Println("The big query has finished. No more errors until the next big query are acceptable")
	err = db.BlobRemove("longvalue")
	if err == nil {
		t.Error("blob delete operation should fail")
	}
}
