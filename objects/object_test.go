package objects

import (
	"fmt"
	"testing"
)

func TestParseHeader(t *testing.T) {
	blob := NewBlob(10)
	sampleHeader := Header(blob)
	obType, size, err := parseHeader(sampleHeader)

	if err != nil {
		t.Errorf("Unexpected error when parsing header:\n%s", err.Error())
	}

	if obType != Blob {
		t.Errorf("Expected object type to be 'blob', Got: %s", obType)
	}

	if size != 10 {
		t.Errorf("Expected size to be '10', Got: %d", size)
	}
}

func TestHeaderGen(t *testing.T) {
	sampleHeader := make([]byte, 8, 8)
	sampleHeader[0] = 'b'
	sampleHeader[1] = 'l'
	sampleHeader[2] = 'o'
	sampleHeader[3] = 'b'
	sampleHeader[4] = 0x20
	sampleHeader[5] = '1'
	sampleHeader[6] = '0'
	sampleHeader[7] = 0x00

	blob := NewBlob(10)
	blobHeader := Header(blob)

	if !equalByteSlices(sampleHeader, blobHeader) {
		t.Errorf("Header() function returned an unexpected result")
	}

}

//Compress and Decompress using results from git
func TestHash(t *testing.T) {
	blob := NewBlob(0)
	blob.Deserialize([]byte("hello world"))
	hash := Hash(blob)
	expectedHash := "95d09f2b10159347eece71399a7e2e907ea3df4f"

	if hash != expectedHash {
		t.Errorf("Expected hash to be: \n%s\nGot:\n%s", expectedHash, hash)
		fmt.Println("Byte contents of blob:")
		printBytes(blob.Serialize())
	}

	blob.Deserialize([]byte("what is up, doc?"))
	hash = Hash(blob)
	expectedHash = "bd9dbf5aae1a3862dd1526723246b20206e5fc37"
	if hash != expectedHash {
		t.Errorf("Expected hash to be: \n%s\nGot:\n%s", expectedHash, hash)
		fmt.Println("Byte contents of blob:")
		printBytes(blob.Serialize())
	}
}

func printBytes(src []byte) {
	l := len(src)
	for i := 0; i < l; i++ {
		fmt.Printf("%d ", src[i])
	}
	fmt.Println("")
}
