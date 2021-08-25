package objects

import  (
    "fmt"
    "encoding/hex"
)

// Format of tree objects:
// Header-[Entries]
// where entry: [mode] [name]0x00[SHA-1 of blob/tree being referenced in BINARY]
// there is no separation between trees


// A GitTree object that fullfils the GitObject interface
type GitTree struct {
    size int
    entries []*treeEntry
}

// The possible file modes for a Tree Entry
// based on the git documentation
type EntryFileMode string
const (
    Normal EntryFileMode = "100644"
    Executable EntryFileMode = "100755"
    SymbolicLink EntryFileMode = "120000"
    Directory EntryFileMode = "040000"
)

// Represents a single entry in a GitTree
type treeEntry struct {
    mode EntryFileMode
    name string
    hash []byte
}
// Process a byte slice into a an tree entry
func byteToEntry(data []byte) treeEntry  {
    spaceByte := 0
    for ; data[spaceByte] != 0x20; spaceByte++ {
    }

    mode := EntryFileMode(data[0:spaceByte])
    
    nullByte := spaceByte + 1
    for ; data[nullByte] != 0x00; nullByte++ {
    }
    name := string(data[spaceByte+1:nullByte])
    hash := data[nullByte+1:]
    return treeEntry{mode:mode, name:name, hash:hash}
}
// Convert an entry into a byte slice for serializing
func (entry *treeEntry) toBytes() []byte {
    bytes := make([]byte, 0, entry.Size())
    bytes = append(bytes, []byte(entry.mode)...)
    bytes = append(bytes, 0x20)
    bytes = append(bytes, []byte(entry.name)...)
    bytes = append(bytes, 0x00)
    bytes = append(bytes, entry.hash...)

    return bytes
}
// Return the size of the entry
func (entry *treeEntry) Size() int {
    return len(entry.mode) + len(entry.name) + len(entry.hash) + 2
}

func (entry *treeEntry) String() string {
    return string(entry.mode) + " " + entry.name + " " + hex.EncodeToString(entry.hash)
}

//Return the size of the tree data (excluding header) as a string
func (tree *GitTree) Size() string {
    return fmt.Sprint(tree.size)
}
// Return 'tree'
func (tree *GitTree) Type() GitObjectType {
    return Tree
}
// Convert tree into a []byte ready for writing into a file; includes header
func (tree *GitTree) Serialize() []byte {
    serializedTree := make([]byte, 0, tree.size)
    for _, entry := range tree.entries {
        serializedTree = append(serializedTree, entry.toBytes()...)
    }
    return serializedTree
}

// Returns the tree as a string
func (tree *GitTree) String() string {
    treeString := ""
    for _, entry := range tree.entries {
        treeString +=  entry.String() + "\n"
    }
    return treeString
}

// Convert a byte slice into a tree
// Should not include the header bytes
func (tree *GitTree) Deserialize(src []byte) {
    entries := make([]*treeEntry, 0, 5)
    curByte, startByte := 0, 0
    size := len(src)
    for curByte < size {
        // Set curByte to the position of the null byte
        for ;src[curByte] != 0x00; curByte++ {
        }
        // Process from startByte to curByte+20 and append it to the tree
        // curByte is at the null byte and the SHA1 is 20 bytes long. So the end of the
        // entry is null byte + 20; we set it to null byte + 21 since the last byte is 
        // ignored when slicing.
        entry := byteToEntry(src[startByte:curByte+21])
        entries = append(entries, &entry)
        // Start the loop again by setting startByte and curByte 
        //to the beginning of the next entry
        startByte = curByte + 21
        curByte = startByte
    }
    tree.entries = entries
    tree.size = size
}
// Add a entry into the tree object
func (tree *GitTree) AddEntry(mode EntryFileMode, name string, hash string) {
    hashBytes, err := hex.DecodeString(hash)
    if err != nil {
        panic(err)
    }
    entry := treeEntry{mode:mode, name:name, hash:hashBytes}
    tree.entries = append(tree.entries, &entry)
    tree.size += entry.Size() 
}
