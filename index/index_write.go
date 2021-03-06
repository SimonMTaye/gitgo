package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
)

// Returns a header struct with the signature bits set to the default
func createIndexHeader(version int, entries int) indexHeader {
	return indexHeader{
		Signature: defaultSignature,
		Version:   int32(version),
		NumEntry:  int32(entries),
	}
}

// Currently doesn't support modifying the stage bit
// Could use an int for the stage parameter and panic if it is out of bounds (i.e. > 4 or < 0)
func createFlag(assumeValid bool, extended bool, name string) entryFlags {
	data := int16(0)
	if assumeValid {
		// Set the assume valid bit and then shift it one bit to left
		// so the rest of the flags can be set
		data += 1
		data = data << 1
	}
	if extended {
		// Do the same as for the assumeValid bit
		data += 1
		data = data << 1
	}
	// We won't be setting the stage bits, so we will just shift the data to left by 2
	// bits, which will set them to 0, which is the default
	data = data << 2
	// Finaly, get the name length, clamp it to a value that can fit in 12 bits and add
	// it to the data
	nameLen := int16(len(name))
	// Clamp the length so it doesn't overflow 12 bits
	if nameLen > 4095 {
		nameLen = 4095
	}
	data += nameLen
	return entryFlags(data)
}

// Serialize Use the binary package to covert the metadata directly to bytes as no processing
// needs to be done
func (idxMdt *indexEntryMetadata) Serialize() []byte {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, idxMdt)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// Serialize Convert an IndexEntry into bytes
func (idx *Entry) Serialize() []byte {
	data := idx.Metadata.Serialize()
	data = append(data, idx.Name...)
	// Name must be null terminated
	data = append(data, 0x0)
	if idx.Extended() && idx.V3Flags != nil {
		// TODO Untested
		flagBytes := make([]byte, 0, 2)
		binary.BigEndian.PutUint16(flagBytes, uint16(*idx.V3Flags))
		data = append(data, flagBytes...)
	}
	// Add padding so each entry is a multiple of 8 bytes (only done in v2 and v3
	if len(data)%8 != 0 {
		fillerNum := 8 - (len(data) % 8)
		// This assumes that the zero value for a byte is 0
		nulls := make([]byte, fillerNum)
		data = append(data, nulls...)
	}
	return data
}

// Hash Conveinience function for getting the hash of an object
func (idx *Entry) Hash() []byte {
	return idx.Metadata.ObjHash[:]
}

// Convert a file mode (stored as a uint32) into a human-readable string
func formattedMode(mode uint32) string {
	modestr := ""
	for i := 0; i < 3; i++ {
		// +16 because the first 16 bits are always set to 0 and the data
		// begins at bit 16 (i.e. the 17th bit)
		if BitSet32(mode, i+16) {
			modestr += "1"
		} else {
			modestr += "0"
		}
	}
	for i := 0; i < 3; i++ {
		// Check the 3-bit values that represent user, group and everyone's permission
		// each permission is set by 3 bits (i.e. value of 0-7)
		bitgroup := mode >> (6 - (i * 3))
		bitgroup = bitgroup & 0x7
		modestr += fmt.Sprint(bitgroup)
	}
	return modestr
}

// Converts an entry into its string form
func (idx *Entry) String() string {
	hashString := hex.EncodeToString(idx.Hash())
	modeString := formattedMode(idx.Metadata.FileMode)
	return fmt.Sprintf("%s %s %d\t%s", modeString, hashString, idx.stage(), idx.Name)
}

// Serialize Convert an Extension's metadata into a byte slice
func (extMeta *ExtensionMetadata) Serialize() []byte {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, extMeta)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// Serialize Convert an extension into a byte slice
func (ext *Extension) Serialize() []byte {
	data := ext.Metadata.Serialize()
	data = append(data, ext.Data...)
	return data
}

// Serialize Convert a header into a byte slice
func (hdr *indexHeader) Serialize() []byte {
	dataInBytes := make([]byte, 12)
	n := copy(dataInBytes, hdr.Signature[:4])
	if n != 4 {
		panic("Bytes were not copied correctly")
	}
	// Write the version into the byte slice
	binary.BigEndian.PutUint32(dataInBytes[4:8], uint32(hdr.Version))
	// Write the name into the byte slice
	binary.BigEndian.PutUint32(dataInBytes[8:12], uint32(hdr.NumEntry))
	return dataInBytes
}

// Covert an Index struct into a byte slice
func (idx *Index) bytesWithoutHash() []byte {
	dataInBytes := make([]byte, 0)
	dataInBytes = append(dataInBytes, idx.Header.Serialize()...)
	// Add entries
	for _, entry := range idx.Entries {
		entryBytes := entry.Serialize()
		dataInBytes = append(dataInBytes, entryBytes...)

	}
	// Add extensions
	for _, ext := range idx.Extensions {
		dataInBytes = append(dataInBytes, ext.Serialize()...)
	}
	return dataInBytes
}

// Sets the indexs hash field based on the data it contains
func (idx *Index) calculateHash() error {
	dataInBytes := idx.bytesWithoutHash()
	hash := sha1.Sum(dataInBytes)
	idx.Hash = hash[:]
	// In case there is an Error in this process, return it
	return nil
}

// Serialize Adds the hash to the []byte returned by bytesWithoutHash. Functions are separated
// for use when computing the hash it self
func (idx *Index) Serialize() []byte {
	return append(idx.bytesWithoutHash(), idx.Hash...)
}

// EntryExists Check if an entry with the specified name exists
func (idx *Index) EntryExists(name string) (bool, int) {
	for i, entries := range idx.Entries {
		if entries.Name == name {
			return true, i
		}
	}
	return false, -1
}

// addEntry Add an entry to an index struct
func (idx *Index) addEntry(entry *Entry) error {
	// Increment the entry num
	idx.Header.NumEntry++
	idx.Entries = append(idx.Entries, entry)
	idx.sortEntries()
	// Return calculateHash since it also returns an Error
	return idx.calculateHash()
}

// Adds an entry to the index struct if it doesn't exist or replaces the existing entry
// with the provided one if it already there
func (idx *Index) updateEntry(entry *Entry) error {
	exists, pos := idx.EntryExists(entry.Name)
	if exists {
		err := idx.DeleteEntry(pos)
		if err != nil {
			return err
		}
	}
	return idx.addEntry(entry)
}

// AddFile AddFiles adds a file to the index or updates its information if it already exists
func (idx *Index) AddFile(rootDir string, fileName string) error {
	entry, err := createEntry(rootDir, fileName)
	if err != nil {
		return err
	}
	// If the entry already exists, update it
	err = idx.updateEntry(entry)
	return err
}

// ModifyFileHash Sets the hash of an entry to the provided hash
func (idx *Index) ModifyFileHash(fileName string, newHash *[20]byte) error {
	exists, pos := idx.EntryExists(fileName)
	if !exists {
		return fmt.Errorf("file %s does not exist in index", fileName)
	}
	entry := idx.Entries[pos]
	if entry.Metadata.ObjHash == *newHash {
		return fmt.Errorf("file %s already has same hash", fileName)
	}
	return nil
}

// DeleteEntry Delete an Entry from the index
func (idx *Index) DeleteEntry(pos int) error {
	if pos < 0 || pos >= len(idx.Entries) {
		return errors.New("the position provided for deletion is invalid")
	}
	// Costly operation
	idx.Entries = append(idx.Entries[:pos], idx.Entries[pos+1:]...)
	return nil
}

// Sorts the entries in an index based on their Name (i.e. file name) and if they match
// their staging values,
func (idx *Index) sortEntries() {
	// Sort the index entries whenver a new entry is added
	// Index is ALWAYS assumed to be sorted
	sort.Slice(idx.Entries, func(i, j int) bool {
		if idx.Entries[i].Name == idx.Entries[j].Name {
			return idx.Entries[i].stage() < idx.Entries[j].stage()
		}
		return idx.Entries[i].Name < idx.Entries[j].Name
	})

}

//TODO Add search entries for quickly finding  an index

// AddExtension Write an extension with its data into the index. Determines the size value of stored
// in the extension header based on the length of data
func (idx *Index) AddExtension(signature [4]byte, data []byte) error {
	extHeader := &ExtensionMetadata{Signature: signature, Size: int32(len(data))}
	if int(extHeader.Size) < len(data) {
		return errors.New("extension has too much data")

	}
	ext := &Extension{Metadata: extHeader, Data: data}
	idx.Extensions = append(idx.Extensions, ext)
	// Return Error incase there is an Error somewhere
	return nil
}

func (idx *Index) IsEmpty() bool {
	return idx.Header.NumEntry == 0
}

// EmptyIndex Create an empty index. Used for repositories where the staging file is not present
func EmptyIndex() *Index {
	header := createIndexHeader(2, 0)
	index := &Index{
		Header:     &header,
		Entries:    make([]*Entry, 0),
		Extensions: make([]*Extension, 0),
		Hash:       make([]byte, 20),
	}
	err := index.calculateHash()
	if err != nil {
		return nil
	}
	return index
}
