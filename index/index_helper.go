package index

import (
	"encoding/hex"
	"errors"
	"github.com/SimonMTaye/gitgo/objects"
	"os"
)

const REGULAR_0644 uint32 = 33188
const REGULAR_0755 uint32 = 33261
const SYMBOLIC_LINK uint32 = 40960
const GIT_LINK uint32 = 57344

func getHash(filepath string) ([20]byte, error) {
	hashBytes := [20]byte{}
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return hashBytes, err
	}
	emptyBlob := &objects.GitBlob{}
	emptyBlob.Deserialize(fileBytes)
	hashStr := objects.Hash(emptyBlob)
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		return hashBytes, err
	}
	n := copy(hashBytes[:], hash)
	if n != 20 {
		return hashBytes, errors.New("expected 20 hash bytes")
	}
	return hashBytes, nil
}

// Process the file mode returned by a 'stat' call into the format git expects
func getFileMode(info os.FileInfo) uint32 {
	if info.Mode().IsRegular() {
		return REGULAR_0644
	}
	return GIT_LINK
}
