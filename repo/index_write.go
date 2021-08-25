package repo
import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"os"
	"path"
    "syscall"
    "errors"
    "sort"
    "fmt"
)
// Returns a header struct with the signature bits set to the default
func CreateIndexHeader(version int, entries int) IndexHeader {
    return IndexHeader{
                Signature:defaultSignature,
                Version:int32(version),
                NumEntry:int32(entries),
            }
}
// Currently doesn't support modifying the stage bit
// Could use an int for the stage parameter and panic if it is out of bounds (i.e. > 4 or < 0)
func CreateFlag(assumeValid bool, extended bool,  name string) entryFlags {
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
// Process the file mode returned by a 'stat' call into the format git expects
func parseFileMode (statMode uint32) uint32 {
    // TODO Filemode might need to be modified for git, not unsure if it is
    // compatible with the filemode returned by stat
    return statMode
}
// Read a file path and create an entry
func CreateEntry(repoPath string, filePath string) (*IndexEntry, error) {
    fileInfo, err := os.Stat(path.Join(repoPath, filePath))
    if err != nil {
        return nil, err
    }
    stat, ok := fileInfo.Sys().(*syscall.Stat_t)
    if !ok {
        return nil, errors.New("error getting 'stat' information for file: "+ 
                                path.Join(repoPath, filePath))
    }
    fileBytes, err := os.ReadFile(path.Join(repoPath, filePath))
    if err != nil {
        return nil, err
    }
    hash := sha1.Sum(fileBytes)
    entryMetdata := &indexEntryMetadata{
                        Ctime: CovertTimespec(stat.Ctim),
                        Mtime: CovertTimespec(stat.Mtim),
                        Ino: uint32(stat.Ino),
                        Dev: uint32(stat.Dev),
                        Uid: stat.Uid,
                        Gid: stat.Gid,
                        FileMode: parseFileMode(stat.Mode),
                        FileSize: int32(fileInfo.Size()),
                        Flags: CreateFlag(false, false, filePath),
                        ObjHash: hash,
                    }
    idxEntry := &IndexEntry{Metadata: entryMetdata, Name: filePath, V3Flags: nil}
    return idxEntry, nil
}
// Use the binary package to covert the metadata directly to bytes as no processing
// needs to be done
func (idxMdt *indexEntryMetadata) ToBytes() []byte{
    buf := &bytes.Buffer{}
    err := binary.Write(buf, binary.BigEndian, idxMdt)
    if err != nil {
        panic(err)
    }
    return buf.Bytes()
}    
// Convert an IndexEntry into bytes
func (idx *IndexEntry) ToBytes () []byte {
    data := idx.Metadata.ToBytes()
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
    if len(data) %8 != 0 {
        fillerNum := 8 - (len(data) % 8)
        // This assumes that the zero value for a byte is 0
        nulls := make([]byte, fillerNum, fillerNum)
        data = append(data, nulls...)
    }
    return data
}
// Conveinience function for getting the hash of an object
func (idx *IndexEntry) Hash() []byte {
    return idx.Metadata.ObjHash[:]
}
// Convert a file mode (stored as a uint32) into a human-readable string
func FormattedMode (mode uint32) string {
    modestr := ""
    for i := 0; i < 3; i++ {
        // +16 because the first 16 bits are always set to 0 and the data
        // begins at bit 16 (i.e. the 17th bit)
        if bitSet32(mode, i + 16) {
            modestr += "1"
        } else {
            modestr += "0"
        }
    }
    for i := 0; i < 3; i ++ {
        // Check the 3-bit values that represent user, group and everyone's permission
        // each permission is set by 3 bits (i.e. value of 0-7)
        bitgroup := mode >> (6 - (i * 3))
        bitgroup = bitgroup & 0x7
        modestr += fmt.Sprint(bitgroup)
    }
    return modestr
}
// Converts an entry into its string form
func (idx *IndexEntry) Stringer() string {
    hashString := hex.EncodeToString(idx.Hash())
    modeString := FormattedMode(idx.Metadata.FileMode)
    return fmt.Sprintf("%s %s %d\t%s", modeString, hashString, idx.Stage(), idx.Name)
}
// Convert an Extension's metadata into a byte slice
func (extMeta *ExtensionMetadata) ToBytes () []byte {
    buf := &bytes.Buffer{}
    err := binary.Write(buf, binary.BigEndian, extMeta)
    if err != nil {
        panic(err)
    }
    return buf.Bytes()
}
// Convert an extension into a byte slice
func (ext *Extension) ToBytes() []byte {
    data := ext.Metadata.ToBytes()
    data = append(data, ext.Data...)
    return data
}
// Convert a header into a byte slice
func (hdr *IndexHeader) ToBytes() []byte{
    bytes := make([]byte, 12, 12)    
    n := copy(bytes, hdr.Signature[:4])
    if n != 4 {
        panic("Bytes were not copied correctly")
    }
    // Write the version into the byte slice
    binary.BigEndian.PutUint32(bytes[4:8], uint32(hdr.Version))
    // Write the name into the byte slice
    binary.BigEndian.PutUint32(bytes[8:12], uint32(hdr.NumEntry))
    return bytes
}
// Covert an Index struct into a byte slice
func (idx *Index) bytesWithoutHash () []byte {
    bytes := make([]byte, 0 ,0)
    bytes = append(bytes, idx.Header.ToBytes()...)
    // Add entries
    for _, entry := range idx.Entries {
        entryBytes := entry.ToBytes()
        bytes = append(bytes, entryBytes...)
        
    }
    // Add extensions
    for _, ext := range idx.Extensions {
        bytes = append(bytes, ext.ToBytes()...)
    }
    return bytes
}
// Adds the hash to the []byte returned by bytesWithoutHash. Functions are separated
// for use when computing the hash it self
func (idx *Index) ToBytes() []byte {
    return append(idx.bytesWithoutHash(), idx.Hash...)
}
// Sets the indexs hash field based on the data it contains
func (idx *Index) CalculateHash () error {
    bytes := idx.bytesWithoutHash()
    hash := sha1.Sum(bytes)
    idx.Hash = hash[:]
    // In case there is an error in this process, return it 
    return nil
}
// Check if an entry with the specified name exists
func (idx *Index) EntryExists (name string) (bool, int) {
    for i, entries := range idx.Entries {
        if entries.Name == name {
            return true, i
        }
    }
    return false, -1
}
// Add an entry to an index struct 
// TODO should check for existing entry first
func (idx *Index) AddEntry (entry *IndexEntry) error {
    // Increment the entry num
    idx.Header.NumEntry ++
    idx.Entries = append(idx.Entries, entry)
    idx.SortEntries()
    // Return CalculateHash since it also returns an error
    return idx.CalculateHash()
}
// Delete an Entry from the index
func (idx *Index) DeleteEntry (pos int) error {
    if pos < 0 || pos >= len(idx.Entries){
        return errors.New("The position provided for deletion is invalid")
    }
    // Costly operation
    idx.Entries = append(idx.Entries[:pos], idx.Entries[pos+1:]...)
    return nil
}
// Adds an entry to the index struct if it doesn't exist or replaces the existing entry
// with the provided one if it already there
func (idx *Index) UpdateEntry (entry *IndexEntry) error {
    exists, pos := idx.EntryExists(entry.Name)
    if exists {
        err := idx.DeleteEntry(pos)
        if err != nil {
            return err
        }
    }
    return idx.AddEntry(entry)
}
// Sorts the entries in an index based on their Name (i.e. file name) and if they match
// their staging values, 
func (idx *Index) SortEntries() {
// Sort the index entries whenver a new entry is added
    // Index is ALWAYS assumed to be sorted
    sort.Slice(idx.Entries, func(i, j int) bool {
        if idx.Entries[i].Name ==  idx.Entries[j].Name {
            return idx.Entries[i].Stage() < idx.Entries[j].Stage()
        }
        return idx.Entries[i].Name < idx.Entries[j].Name
    })

}
//TODO Add search entries for quickly finding  an index
// Write an extension with its data into the index. Determines the size value of stored
// in the extension header based on the length of data
func (idx *Index) AddExtension (signature [4]byte, data []byte) error {
    extHeader := &ExtensionMetadata{Signature: signature, Size: int32(len(data))}
    if int(extHeader.Size) < len(data) {
        return errors.New("Extension has too much data")

    }
    ext := &Extension{Metadata:extHeader, Data:data}
    idx.Extensions = append(idx.Extensions, ext)
    // Return error incase there is an error somewhere
    return nil
}
