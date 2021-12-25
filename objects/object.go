package objects

import (
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
)

type ErrBadObject struct {
	reason string
}

func (e *ErrBadObject) Error() string {
	return "Could not read object: " + e.reason
}

// GitObject Interface for all Git objects
type GitObject interface {
	ByteDeserializable
	ByteSerialized
	fmt.Stringer
	// Type A string of the object type; using a custom type for clarity
	Type() GitObjectType
	// Size An ASCII representation of the size of the object
	Size() string
}

type ByteSerialized interface {
	// Serialize Convert object data into bytes
	// Header should NOT be part of the bytes
	Serialize() []byte
}

type ByteDeserializable interface {
	ByteSerialized
	// Deserialize Parse a bunch of bytes into meaningful data
	// Function will return an error if the object is incorrectly formatted
	Deserialize(data []byte)
}

// GitObjectType Denote the object type
type GitObjectType string

const (
	Blob   GitObjectType = "blob"
	Commit GitObjectType = "commit"
	Tree   GitObjectType = "tree"
	Tag    GitObjectType = "tag"
)

// CompressAndSave TODO test
// Compress a git object and writes it to an io.Writer
func CompressAndSave(dst io.Writer, obj GitObject) error {
	zWriter := zlib.NewWriter(dst)
	// Write the object Header
	zWriter.Write(Header(obj))
	// Write the data itself
	zWriter.Write(obj.Serialize())
	return zWriter.Close()
}

// DecompressAndRead decompress contents in src and parse the resulting data as an object
func DecompressAndRead(src io.Reader) (GitObject, error) {
	zReader, err := zlib.NewReader(src)
	defer zReader.Close()
	if err != nil {
		return nil, err
	}
	// Arbitrarily chosen. Most git objects are <200 from experience blobs may be higher.
	// 300 feels like a good middle ground between too many allocations and too much memory
	data, err := io.ReadAll(zReader)
	if err != nil {
		return nil, err
	}
	return ObjectFromBytes(data)
}

// ObjectFromBytes read a bunch of bytes and return the correct object
func ObjectFromBytes(src []byte) (GitObject, error) {
	nulPos := 0
	// Increment nulPos until it is the index of the null byte at the end of the header
	l := len(src)
	for ; src[nulPos] != 0x00; nulPos++ {
		if nulPos == l-1 {
			return nil, &ErrBadObject{reason: "object is badly formed"}
		}
	}
	objType, _, err := parseHeader(src[:nulPos+1])
	if err != nil {
		return nil, err
	}
	var obj GitObject
	switch objType {
	case Commit:
		obj = &GitCommit{}
	case Tree:
		obj = &GitTree{}
	case Tag:
		obj = &GitTag{}
	case Blob:
		obj = &GitBlob{}
	default:
		return nil, &ErrBadObject{reason: string(objType) + " is not a valid type"}
	}
	// TODO Added error checking once it has been added to objects
	obj.Deserialize(src[nulPos+1:])
	return obj, nil
}

// Helper function for reading an object's header and returing the relevant information
func parseHeader(header []byte) (GitObjectType, int, error) {

	var objType GitObjectType
	spacePos := 0

	if string(header[0:3]) == "tag" {
		spacePos = 3
		objType = Tag
	} else if string(header[0:4]) == "blob" {
		spacePos = 4
		objType = Blob
	} else if string(header[0:4]) == "tree" {
		spacePos = 4
		objType = Tree
	} else if string(header[0:6]) == "commit" {
		spacePos = 6
		objType = Commit
	}
	// Convert the ASCII size representation into an int
	size, err := strconv.Atoi(string(header[spacePos+1 : len(header)-1]))
	if err != nil {
		return "", 0, err
	}
	return objType, size, nil
}

// Read compressed GitObject data and return an appropirate GitObject
// Such as a GitBlob, GitTree (Unimplemented), GitCommit (Unimplemented) or GitTag (Unimplemented)
//func Decompress(src io.Reader) (GitObject, error) {
//}

// Hash A SHA1 representation of an object and its header
func Hash(obj GitObject) string {
	hashBytes := sha1.Sum(append(Header(obj), obj.Serialize()...))
	return hex.EncodeToString(hashBytes[:])
}

// Header A git header for the object
// A header should be in the format:
// Object Type + Space (0x20) + Object Size + Null Byte (0x00)
func Header(obj GitObject) []byte {
	typeAndSpace := append([]byte(obj.Type()), 0x20)
	typeSize := append(typeAndSpace, []byte(obj.Size())...)
	return append(typeSize, 0x00)
}
