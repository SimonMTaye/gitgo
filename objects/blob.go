package objects

import (
    "fmt"
    )

type GitBlob struct {
    size int
    data []byte
}

func NewBlob(size int) *GitBlob {
    return &GitBlob{size:size}
}

// Simply returns the data of the blob
func (obj *GitBlob) Serialize() []byte {
    return obj.data
}

// Returns the blob's content
func (blob *GitBlob) Stringer() string {
    return string(blob.data)
}

// Sets the data field of the GitBlob struct
func (obj *GitBlob) Deserialize(src []byte) {
    obj.data = src
    obj.size = len(src)
}

// Returns 'blob'
func (obj *GitBlob) Type() GitObjectType{
    return Blob
}
// Returns an ASCII representation of the size of the blob
func (obj *GitBlob) Size() string {
    return fmt.Sprint(obj.size)
}
