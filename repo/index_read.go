package repo

import (
    "io"
    "encoding/binary"
    "syscall"
    "fmt"
    "bytes"
    )

type ErrIndexBadlyFormated struct {
    reason string
}
func (e *ErrIndexBadlyFormated) Error() string {
    return "Error parsing index file: " + e.reason
}
// Default signature found in index files (other signature indicate extensions
// *docs specify version number as "4 byte version number" instead of a 32-bit number like
// numEntry
type IndexHeader struct {
    Signature [4]byte
    Version int32
    NumEntry int32
}
// As stated by the git docs
var defaultSignature = [4]byte{'D','I','R','C'}
// Check if an index has expected signature bytes
func (indexhd *IndexHeader) ValidSignature() bool {
    return indexhd.Signature == defaultSignature
}
// Returns the version num as an int
func (indexhd *IndexHeader) VersionNum() int {
    return int(indexhd.Version)
}

// Check if a certain bit in a 16 bit is a 0 or a 1. 'true' means 1 and 'false' means 0
// Used by flag structs to check if a certain flag was set or not
// pos is the 0-index left-to-right position of the bit that is to be checked 
func bitSet (bits uint16, pos int) bool {
    if pos > 15 || pos < 0 {
        return false
    }
    wantedBits := bits >> (15 - pos)
    // Set all bits except the final bit to 0
    wantedBits = wantedBits & 0x1
    return wantedBits == 1
}
// Check if a certain bit in a 32 bit is a 0 or a 1. 'true' means 1 and 'false' means 0
// Used by flag structs to check if a certain flag was set or not
// pos is the 0-index left-to-right position of the bit that is to be checked 
func bitSet32 (bits uint32, pos int) bool {
    if pos > 31 || pos < 0 {
        return false
    }
    wantedBits := bits >> (31 - pos)
    // Set all bits except the final bit to 0
    wantedBits = wantedBits & 0x1
    return wantedBits == 1
}
// Represents 16 bits that indicate flags (in left-to-right order)
// 1-bit  : assume valid flag
// 1-bit  : extended falg
// 2-bit  : stage (bits used for handling merge conflicts, not used by this programs)
// 12-bit : name length (if all bits are 1, name may be larger)
// uint16 is used instead of [2]byte for easy bitwise operations
type entryFlags uint16
// Check if the 'extended' flag is set
func (eF *entryFlags) Extended() bool {
    return bitSet(uint16(*eF), 1)
}
// Return the length of name. A value of 4095 (the max for 12 bits) means the name may be
// greater. Currently, these names will be unsupported by the program
func (eF *entryFlags) NameLength() int {
    // Set the first four bits of the flag to zero and return the remaining number
    return int(*eF & 0xfff)
}
// Returns the contents of the stage bits (an int with a value of 0-3)
// These bits are used for resolving merge conflicts
// 0 -  Normal (no conflicts)
// 1 -  Base (Ancestor)
// 2 -  HEAD (Local)
// 3 -  External (Version being merged)
// Sourced from : https://mincong.io/2018/04/28/git-index on 21:50, July 16, 2021
func (eF *entryFlags) Stage() int {
    // Shift the bits so that the staging bits are at the right end (assuming bits are l-t-r)
    wantedBits := *eF >> 13
    // Zero out all preceding bits, leaving only the staging ones
    wantedBits = wantedBits & 0x11
    return int(wantedBits)
}
// Alias for flag.Extended()
func (idxEntry *IndexEntry) Extended() bool {
    return idxEntry.Metadata.Flags.Extended()
}
// Alias for flag.NameLength()
func (idxEntry *IndexEntry) NameLength() int {
    return idxEntry.Metadata.Flags.NameLength()
}
// Alias for flag.Stage()
func (idxEntry *IndexEntry) Stage() int {
    return idxEntry.Metadata.Flags.Stage()
}
type version3Flags uint16
// Struct for storing Sec and Nano-sec pair (for c-time and m-time) as an int32 pair
type TimePair struct {
    Sec int32
    Nsec int32
}
func CovertTimespec(timeSpec syscall.Timespec) TimePair {
    return TimePair{Sec: int32(timeSpec.Sec), Nsec: int32(timeSpec.Nsec) }
}
// Metadata for an index entry, does not include a name. 
// This struct is used as the length of the struct is fixed allowing for easy parsing
// Variable length items (such as v3 headers and names) will be parsed separetly
type indexEntryMetadata struct {
    // Last time metadata changed. First int is seconds, second one is fractional nanoseconds
    Ctime TimePair
    // Last time file data changed in seconds
    Mtime TimePair
    // Inode number and device id for the file being represnted by this entries
    Ino uint32
    Dev uint32
    // Object type (regular, symbolic link or gitlink) - 4 bits and 
                 // 1000,    1010             1110
    // unix permission - 12 bits
        // first three bits are always 0
    // this adds up to 16 bits 
    // The first 16 bits are all 0 with the next 16 bits containing the data described above
    FileMode uint32
    // User and group id of the owner of the file
    Uid uint32
    Gid uint32
    // size of file in bytes
    FileSize int32
    // flags, are 16 bits wide
    ObjHash [20]byte
    Flags entryFlags
}
// Struct that represents a single index entry, usually a file
type IndexEntry struct {
    Metadata *indexEntryMetadata
    Name string
    V3Flags *version3Flags
}
// First 8 bytes of an extension which is the same for all extensions
type ExtensionMetadata struct {
    Signature [4]byte
    Size int32
}
// Holds index extension information. This program won't handle extensions, this is merely
// for parsing  an index file correctly
type Extension struct {
    Metadata *ExtensionMetadata
    Data []byte
}
// Struct that represents an index
type Index struct {
    Header *IndexHeader
    Entries []*IndexEntry
    Extensions []*Extension
    // 20 because that's how long sha1 hashes are
    Hash []byte
}
// Parse bytes from src into an IndexHeader struct
func parseHeader(src io.Reader) (*IndexHeader, error) {
    header := &IndexHeader{}
    err := binary.Read(src, binary.BigEndian, header)
    if err !=  nil {
        return nil,  err
    }
    return header, nil
}
// Parse bytes from io.Reader into an IndexEntry
func ParseEntry(src io.Reader, idxVersion int) (*IndexEntry, int, error) {
    // Does not handle version 4 index so throw error
    if idxVersion > 3 {
        return nil, 0, &ErrIndexBadlyFormated{reason: "only index version 2 and 3 are supported"}
    }
    metadata := &indexEntryMetadata{}
    err := binary.Read(src, binary.BigEndian, metadata)
    if err != nil {
        return nil, 0, err
    } 
    // Since the size of the metadata is 62 bytes
    bytesRead := 62
    entry := &IndexEntry{Metadata: metadata}
    // Shortcut to minimize repition 
    flags := entry.Metadata.Flags
    //Read version3 flags if the extended bit is set and the version is 3 or higher
    var v3Flags *version3Flags
    if idxVersion > 2 && flags.Extended() {
        err := binary.Read(src, binary.BigEndian, *v3Flags)
        if err != nil {
            return nil, bytesRead, err
        }
        entry.V3Flags = v3Flags
        // The v3Flags are 16bits or 2 bytes
        bytesRead += 2    
    } else if idxVersion == 2 && flags.Extended() {
        // If index is version 2, then flags.Extended must not be set
        return nil, bytesRead, &ErrIndexBadlyFormated{reason:"index version is 2 but extended flag is set"}
    }
    // Parse the name
    // Does NOT handle names that are larger than set in the flag (this can happen
    // if the name length is greater than the value that can be stored in 12 bits
    // TODO temp variable for debugging
    nLength := flags.NameLength()
    nameBytes := make([]byte, nLength + 1)
    // TODO Error with the below function call, (possible because we read after using binary package?)
    n, err := src.Read(nameBytes)
    // Set the amount of bytes read to the length of the slice. This is possible because
    // the length of the slice is initially 0, so the len() will indicate the amount of bytes
    // read
    if err != nil {
        return nil, bytesRead, err
    }
    // Add the length of the name to the amount of bytesRread
    bytesRead += n
    if n != (flags.NameLength() + 1) {
        reason := fmt.Sprintf("error reading index entry name; expected %d bytes, read %d bytes", 
                                flags.NameLength() + 1,
                                n)
        return nil, bytesRead, &ErrIndexBadlyFormated{reason: reason }
    } 
    entry.Name = string(nameBytes[:len(nameBytes) - 1])
    return entry, bytesRead, nil
}
// Parses all the extensions in the given byte slice and returns them, along with the
// number of bytes read.
// If an error is encounterd, the extensions read so far and the number of bytes read
// successfully will be returned
// Expects the data and not an io.reader because the length of the remaining data is
// required to determine when the extension data is over and the hash data begins
func ParseExtension(data []byte) ([]*Extension, int, error) {
    n := 0
    extensions := make([]*Extension, 0)
    for len(data) - n > 20 {
        // TODO This is very wasteful, find a better way
        breader := bytes.NewReader(data[n:])
        extMeta := &ExtensionMetadata{}
        err := binary.Read(breader, binary.BigEndian, extMeta)
        if err != nil {
            return extensions, n, err
        }
        // The size of the metadata is 8 bytes long
        n += 8
        // Read starting from n (which denotes the bytes  read so far, including the current
        // extensions metadata) upto n + size of the extension data
        ext := Extension {Metadata: extMeta, Data: data[n: n + int(extMeta.Size)]}
        n += int(extMeta.Size)
        extensions = append(extensions, &ext)
    }
    return extensions, n, nil

}
// Parse an Index file
func ParseIndex(src io.Reader) (*Index, error) {
    header, err := parseHeader(src)
    if err != nil {
        return nil, err
    }
    if header.Version < 2 || header.Version > 3 {
        return nil, &ErrIndexBadlyFormated{reason: "only index version 2 and 3 are supported"}
    }
    entries := make([]*IndexEntry, 0, header.NumEntry)
    for i := int32(0); i < header.NumEntry; i++ {
        // Amount of padding bytes : total bytes of entry % 8 (i.e. it is so the entry takes up a multiple of 8 amount of bytes 
        entry, n, err := ParseEntry(src, header.VersionNum())
        // If the version is 2 or 3, reading the padding bytes
        // This check is here even though the version is already guaranteed to be 2 or 3
        // in case version 4 is supported by future versions
        if n % 8 != 0 && header.Version < 4 {
            // Read the padding bytes that will be present
            // They are there to make each entry a multiple of 8 bytes (hence the modulo)
            tempSlice := make([]byte,  8 - (n % 8))
            src.Read(tempSlice)
        }
        if err != nil {
            return nil, err
        }
        entries = append(entries, entry)
    }
    //Extensions are currently ignored, return error if extensions are found Extensions
    idx := &Index{Entries:entries, Header:header}
    // Storage for extension data, won't be parsed for now
    remaining, err := io.ReadAll(src)
    if err != nil {
        return nil, err
    } 
    extensions, n, err := ParseExtension(remaining)
    if err != nil {
        return nil, err
    }
    // The hash will be the reminaing bytes without the the extension data
    hash := remaining[n:]
    
    // if the hash is not 20 bytes long (i.e. it is not a valid sha1 hash, return an
    // error)
    hashLen := 20
    if len(hash) < hashLen {
        return nil, &ErrIndexBadlyFormated{reason: "index is invalid; expeceted more data"}
    } else if len(hash )> hashLen {
        return nil, &ErrIndexBadlyFormated{reason: "index is invalid; contains unexpected data"}
    }
    idx.Extensions = extensions
    idx.Hash = remaining[n:]
    idx.SortEntries()
    return idx, nil
}
