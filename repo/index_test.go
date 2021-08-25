package repo

import (
    "testing"
    "bytes"
    "encoding/binary"
    "crypto/sha1"
    "time"
    )

// V3 Flag parsing/reading NOT tested
// Test that bits from a num are correctly read as '0' or '1'
func TestBitSet(t *testing.T) {
    var num uint16
    num = 1
    num = num << 2
    num += 1
    num = num << 3
    num += 1
    num = num << 2
    num += 1
    num = num << 8
    num += 1
    // Num should be 1010 0101 0000 0001
    if !bitSet(num, 0) {
        t.Errorf("Expected bitSet(num, 0) to be true\nBits: " + PrintBits(num))
    }

    if bitSet(num, 1) {
        t.Errorf("Expected bitSet(num, 1) to be false\nBits: " + PrintBits(num))
    }
    if !bitSet(num, 7) {
        t.Errorf("Expected bitSet(num, 7) to be true\nBits: " + PrintBits(num))
    }

    if bitSet(num, 10) {
        t.Errorf("Expected bitSet(num, 10) to be false\nBits: " + PrintBits(num))
    }
}

func TestHeaderReading(t *testing.T) {
    data := make([]byte, 0,12)
    data = append(data, "DIRC"...)
    // Append 0000 0000 0000 0011 to the data (i.e. 3)
    data = append(data, 0x0)
    data = append(data, 0x0)
    data = append(data, 0x0)
    data = append(data, 0x3)
    // Append 10 to the data
    data = append(data, 0x0)
    data = append(data, 0x0)
    data = append(data, 0x0)
    data = append(data, 0xa)
    
    if len(data) != 12 {
        t.Fatalf("Expected []byte to be 12 bytes long, Got: %d", len(data))
    }

    src := bytes.NewReader(data)
    header, err := parseHeader(src)
    if err != nil {
        t.Fatalf("Unexpected error when parsing bytes:\n %s", err.Error())
    }
    if header == nil {
        t.Fatalf("Expected header to be a indexHeader, not null")
    }

    if !header.ValidSignature() {
        t.Errorf("Expected ValidSignature to be true, Signature: %s", 
        header.Signature)
    }

    if header.NumEntry != 10{
        t.Errorf("Expected numEntry to be 10, Got: %d", header.NumEntry)
    }
    if header.VersionNum() != 3 {
        t.Errorf("Expected version num to be 3, Got: %d", header.VersionNum())
    }
}

// TODO Test Index Reading
// Create a sample Index entry by hand, (that contains extension data) and check that it is
// correctly read
// TODO Test Index Writing
// Recreate the index created previously by hand and check that it can be read and produces
// the same results as the one made by hand
// Tests that index made by hand following the index-file spec of git can be read correctly
func TestIndexRead (t *testing.T) {
    header := make([]byte, 0, 12)
    // Header signature
    header = append(header, defaultSignature[0])
    header = append(header, defaultSignature[1])
    header = append(header, defaultSignature[2])
    header = append(header, defaultSignature[3])
    // Header version number, version number 3
    header = append(header, 0x0)
    header = append(header, 0x0)
    header = append(header, 0x0)
    header = append(header, 0x3)
    // Number of index entries, 2
    header = append(header, 0x0)
    header = append(header, 0x0)
    header = append(header, 0x0)
    header = append(header, 0x2)
    // COMPLETE HEADER
    sampleHash := sha1.Sum(header)
    // Entry where most values are 0
    entry1, err := EntryBytes([2]int32{0, 1}, // ctime
                         [2]int32{0, 1}, //mtime
                         0, // dev
                         0, // ino
                         0, // mode
                         0, // uid
                         0, // gid
                         0, // fileSize
                         sampleHash, // hash
                         "sample entry")
    if err != nil {
        t.Fatalf("Unexpected error when creating first entry:\n%s", err.Error())
    }
    // Adding padding bytes
    if len(entry1) % 8 != 0 {
        entry1 = append(entry1, make([]byte, 8 - (len(entry1) % 8))...)
    }
    // Entry2 will be somewhat representative of a real entry and hold either
    // random data or data similar to real world values
    now := time.Now()
    ctime := [2]int32{int32(now.Second()), int32(now.Nanosecond())}
    // Random
    dev := uint32(2130821)
    // Real-world
    ino := uint32(402549)
    modeBytes := make([]byte, 2, 4)
    // 0x81 = 1000 0001
    modeBytes = append(modeBytes, 0x81)
    // 0xa4 = 1010 0100
    modeBytes = append(modeBytes, 0xa4)
    // Real-world
    mode := binary.BigEndian.Uint32(modeBytes)
    // Random
    uid := uint32(312415223)
    gid := uid
    fileSize := int32(1024)
    entry2, err := EntryBytes(ctime, // ctime
                         ctime, //mtime
                         dev, // dev
                         ino, // ino
                         mode, // mode
                         uid, // uid
                         gid, // gid
                         fileSize, // fileSize
                         sampleHash, // hash
                         "README.md")
    if err != nil {
        t.Fatalf("Unexpected error when creating first entry:\n%s", err.Error())
    }
    if len(entry2) % 8 != 0 {
        entry2 = append(entry2, make([]byte, 8 - (len(entry2) % 8))...)

    }
    // A simulated extension with garbage data
    ext := make([]byte, 0, 19)
    ext = append(ext, "TREE"...)
    // Append 0000-0000 0000-0000 0000-0000 0000-1011
    ext = append(ext, 0x0)
    ext = append(ext, 0x0)
    ext = append(ext, 0x0)
    ext = append(ext, 0xb)
    ext = append(ext, "hello world"...)
    // Combine all the data into one byte slice
    indexBytes := make([]byte, 0, len(header) + len(entry1) + len(entry2) + len(ext))
    indexBytes = append(indexBytes, header...)
    // entry2 before 1 because that is the order expected once they are sorted
    indexBytes = append(indexBytes, entry2...)
    indexBytes = append(indexBytes, entry1...)
    indexBytes = append(indexBytes, ext...)
    hash := sha1.Sum(indexBytes)
    indexBytes = append(indexBytes, hash[:]...)

    index, err := ParseIndex(bytes.NewReader(indexBytes))
    if err != nil {
        t.Fatalf("Unexpected error when parsing index:\n%s", err.Error())
    }
    readIndexBytes := index.ToBytes()
    equalIndex, i := equalSlices(readIndexBytes, indexBytes) 
    if !equalIndex {
        t.Errorf("Expected []byte array produced by index.ToBytes() to be identical"+
                 " to the hand made one but there is a difference at byte %d", i)
    }
    if index.Header.Signature != defaultSignature {
        t.Errorf("Expected index header signature to be 'DIRC', Got: %s", index.Header.Signature[:])
    }
    if index.Header.NumEntry != 2 {
        t.Errorf("Expected index header NumEntry to be 2, Got: %d", index.Header.NumEntry)
    }
    if index.Header.VersionNum() != 3 {
        t.Errorf("Expected index header NumEntry to be 2, Got: %d", index.Header.NumEntry)
    }
    // If these test fail, expand them and check each property
    // First entry should entry2 since they are sorted by file name
    readEntry2Bytes := index.Entries[0].ToBytes()
    equal, i := equalSlices(readEntry2Bytes, entry2)
    if !equal {
        readEntry2, _, _ := ParseEntry(bytes.NewReader(readEntry2Bytes), 3)
        entry2parsed, _, _ := ParseEntry(bytes.NewReader(entry2), 3)
        t.Errorf("Expected entry2 to be equal to the first entry in index but byte %d is different.\nEntries:\n%s\n%s\n",
            i, entry2parsed.Stringer(), readEntry2.Stringer())
    }
    readEntry1Bytes := index.Entries[1].ToBytes()
    equal, i = equalSlices(readEntry1Bytes, entry1) 
    if !equal {
        readEntry1, _, _ := ParseEntry(bytes.NewReader(readEntry1Bytes), 3)
        entry1parsed, _, _ := ParseEntry(bytes.NewReader(entry1), 3)
        t.Errorf("Expected entry1 to be equal to the second entry in index but byte %d is different.\nEntries:\n%s\n%s\n",
            i, entry1parsed.Stringer(), readEntry1.Stringer())
    }
    extension := index.Extensions[0]
    if extension.Metadata.Size != 11 {
        t.Errorf("Expected extension size to be 11, Got: %d", extension.Metadata.Size)
    }

    if extension.Metadata.Signature != [4]byte{'T', 'R', 'E', 'E'} {
        t.Errorf("Expected signature to be 'TREE', Got: %s", extension.Metadata.Signature[:])
    }

    if string(extension.Data) != "hello world" {
        t.Errorf("Expected extension data to be 'hello world', Got: %s", extension.Data) 
    }
}

func equalSlices (slice1, slice2 []byte) (bool, int) {
    l := len(slice1)
    if l != len(slice2) {
        return false, -1
    }
    for i := 0; i < l; i ++ {
        if slice2[i] != slice1[i] {
            return false, i
        }
    }
    return true, -1
}

func EntryBytes (ctime [2]int32, 
                 mtime [2]int32, 
                 dev uint32, 
                 ino uint32, 
                 mode uint32, 
                 uid uint32, 
                 gid uint32,
                 fileSize int32,
                 hash [20]byte,
                 name string) ([]byte, error) {
    buf := &bytes.Buffer{}
    err := binary.Write(buf, binary.BigEndian, ctime)
    if err != nil {
        return nil, err
    }
    err = binary.Write(buf, binary.BigEndian, mtime)
    if err != nil {
        return nil, err
    }

    err = binary.Write(buf, binary.BigEndian, dev)
    if err != nil {
        return nil, err
    }

    err = binary.Write(buf, binary.BigEndian, ino)
    if err != nil {
        return nil, err
    }
    err = binary.Write(buf, binary.BigEndian, mode)
    if err != nil {
        return nil, err
    }
    err = binary.Write(buf, binary.BigEndian, uid)
    if err != nil {
        return nil, err
    }
    err = binary.Write(buf, binary.BigEndian, gid)
    if err != nil {
        return nil, err
     }

    err = binary.Write(buf, binary.BigEndian, fileSize)
    if err != nil {
        return nil, err
    }
    err = binary.Write(buf, binary.BigEndian, hash)
    if err != nil {
        return nil, err
    }
    err = binary.Write(buf, binary.BigEndian, CreateFlag(false, false, name))
    if err != nil {
        return nil, err
    }
    data := buf.Bytes()
    data = append(data, name...)
    data = append(data, 0x0)
    return data, nil
 }

