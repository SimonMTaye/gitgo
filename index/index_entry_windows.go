//go:build windows
// +build windows

package index

import (
	"errors"
	"os"
	"path"
	"syscall"
)

// createEntry Read a file path and create an entry
func createEntry(rootDir string, fileName string) (*Entry, error) {
	fileInfo, err := os.Stat(path.Join(rootDir, fileName))
	if err != nil {
		return nil, err
	}
	stat, ok := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	fileInfo.ModTime().Second()
	if !ok {
		return nil, errors.New("Error getting 'stat' information for file: " +
			path.Join(rootDir, fileName))
	}
	hashBytes, err := getHash(path.Join(rootDir, fileName))
	if err != nil {
		return nil, err
	}
	entryMetdata := &indexEntryMetadata{
		Ctime: convertNanosec(stat.CreationTime.Nanoseconds()),
		Mtime: TimePair{
			Sec:  int32(fileInfo.ModTime().Second()),
			Nsec: int32(fileInfo.ModTime().Nanosecond()),
		},
		// Ino, Dev, Uid and Gid will be ignored and set to 0 for windows
		Ino:      uint32(0),
		Dev:      uint32(0),
		Uid:      uint32(0),
		Gid:      uint32(0),
		FileMode: getFileMode(fileInfo),
		FileSize: int32(fileInfo.Size()),
		Flags:    createFlag(false, false, fileName),
		ObjHash:  hashBytes,
	}
	idxEntry := &Entry{Metadata: entryMetdata, Name: fileName, V3Flags: nil}
	return idxEntry, nil
}

func convertNanosec(nsec int64) TimePair {
	// Set whole number values of nsec to Sec and remaining fractional amount to Nsec
	secs := int32(nsec / 1000000000)
	nsecs := int32(nsec % 1000000000)
	return TimePair{Sec: secs, Nsec: nsecs}
}
