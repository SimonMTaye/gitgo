package objects

import (
	"fmt"
	"os"
)

type GitBlob struct {
	size int
	data []byte
}

func NewBlob(size int) *GitBlob {
	return &GitBlob{size: size}
}

func FileBlob(path string) (*GitBlob, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	blob := &GitBlob{}
	blob.Deserialize(contents)
	return blob, nil

}

// Serialize Simply returns the data of the blob
func (obj *GitBlob) Serialize() []byte {
	return obj.data
}

// Returns the blob's content
func (obj *GitBlob) String() string {
	return string(obj.data)
}

// Deserialize Sets the data field of the GitBlob struct
func (obj *GitBlob) Deserialize(src []byte) {
	obj.data = src
	obj.size = len(src)
}

// Type Returns 'blob'
func (obj *GitBlob) Type() GitObjectType {
	return Blob
}

// Size Returns an ASCII representation of the size of the blob
func (obj *GitBlob) Size() string {
	return fmt.Sprint(obj.size)
}
