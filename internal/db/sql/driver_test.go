package sql

import (
	"testing"

	"github.com/gonimals/elephant/internal/util"
)

func testDriver(t *testing.T, db *driver) {
	err := db.Create("test_table", "1", "asdfasdf")
	if err != nil {
		t.Error("simple create operation fails:", err)
	}
	output, err := db.Retrieve("test_table", "1")
	if err != nil {
		t.Error("simple retrieve operation fails:", err)
	}
	if output != "asdfasdf" {
		t.Error("retrieved string is not the original")
	}
	err = db.Update("test_table", "1", "fdsafdsa")
	if err != nil {
		t.Error("simple update operation fails:", err)
	}
	err = db.Update("test_table", "1", "fdsafdsa")
	if err != nil {
		t.Error("simple update operation fails:", err)
	}
	output, err = db.Retrieve("test_table", "1")
	if err != nil {
		t.Error("simple retrieve operation fails:", err)
	}
	if output != "fdsafdsa" {
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
