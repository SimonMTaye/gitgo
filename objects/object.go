package objects
import (
    "compress/zlib"
    "crypto/sha1"
    "encoding/hex"
    "strconv"
    "io"
)

// Interface for all Git objects
type GitObject interface {
    // A string of the object type; using a custom type for clarity
    Type() GitObjectType
    // An ASCII representation of the size of the object
    Size() string
    // Parse a bunch of bytes into meaningful data
    // Header should NOT be part of the bytes
    Deserialize(data []byte)
    // Convert object data into bytes
    Serialize() []byte
    //String form of object. Most objects can simply use their serialize functions
    Stringer() string
}

// Denote the object type
type GitObjectType string
const (
    Blob GitObjectType = "blob"
    Commit GitObjectType = "commit"
    Tree GitObjectType = "tree"
    Tag GitObjectType = "tag"
)

// Compress a git object and writes it to an io.Writer
func Compress(dst io.Writer, obj GitObject) error {
    zWriter := zlib.NewWriter(dst)
    // Write the object Header
    zWriter.Write(Header(obj))
    // Write the data itself
    zWriter.Write(obj.Serialize())
    return zWriter.Close()
}

func parseHeader(header []byte) (GitObjectType, int, error) {

    var objType GitObjectType
    spacePos := 0

    if string(header[0:2]) == "tag" {
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
    size, err := strconv.Atoi(string(header[spacePos+1  : len(header) - 1]))
    if err != nil {
        return "", 0, err
    }
    return objType, size, nil
}

// Read compressed GitObject data and return an appropirate GitObject
// Such as a GitBlob, GitTree (Unimplemented), GitCommit (Unimplemented) or GitTag (Unimplemented)
//func Decompress(src io.Reader) (GitObject, error) {
//}

// A SHA1 representation of an object and its header
func Hash(obj GitObject) string {
    hashBytes := sha1.Sum(append(Header(obj), obj.Serialize()...))
    return hex.EncodeToString(hashBytes[:])
}

// A git header for the object
// A header should be in the format: 
// Object Type + Space (0x20) + Object Size + Null Byte (0x00)
func Header(obj GitObject) []byte {
    typeAndSpace :=  append([]byte(obj.Type()),  0x20) 
    typeSize := append(typeAndSpace, []byte(obj.Size())...)
    return append(typeSize, 0x00)
}

// Return a file path for the object based on its hash 
// Returns: first-two-chars-of-hash/rest-of-hash
func RelPath(obj GitObject) string {
    hash := Hash(obj)
    return hash[0:3] + "/" + hash[3:]
}


