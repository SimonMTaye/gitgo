package objects

import "testing"


func TestBasicBlobFunctions(t *testing.T) {
    blob := NewBlob(10)

    if blob.Size() != "10" {
        t.Errorf("Expected '10' from Size(), Got: '%s'", blob.Size())
    }

    if blob.Type() != Blob {
        t.Errorf("Expected %s from objectType, Got: %s", Blob, blob.Type())
    }
}

func TestBlobWithData(t *testing.T) {
    blob := NewBlob(10)
    blob.Deserialize([]byte("hello world"))

    if !equalByteSlices(blob.Serialize(), []byte("hello world")) {
        t.Errorf("blob.Serialize() returned an unexpected value")
    }

    if blob.Size() != "11" {
        t.Errorf("Expected blob.Size() to be 11, Got: %s", blob.Size())
    }

}

func equalByteSlices(slice1 []byte, slice2 []byte) bool {
    slLen := len(slice1)
    if  slLen != len(slice2) {
        return false
    }
    for i := 0; i < slLen; i++ {
        if slice1[i] != slice2[i] {
            return false
        }
    }
    return true
}
