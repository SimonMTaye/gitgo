//go:build windows
// +build windows

package repo

import (
	"crypto/sha1"
	"errors"
	"os"
	"path"
	"syscall"
)

// CreateEntry Read a file path and create an entry
func CreateEntry(repoPath string, filePath string) (*IndexEntry, error) {
	fileInfo, err := os.Stat(path.Join(repoPath, filePath))
	if err != nil {
		return nil, err
	}
	stat, ok := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return nil, errors.New("Error getting 'stat' information for file: " +
			path.Join(repoPath, filePath))
	}
	fileBytes, err := os.ReadFile(path.Join(repoPath, filePath))
	if err != nil {
		return nil, err
	}
	hash := sha1.Sum(fileBytes)
	entryMetdata := &indexEntryMetadata{
		Ctime: ConvertNanosec(stat.CreationTime.Nanoseconds()),
		Mtime: ConvertNanosec(stat.LastWriteTime.Nanoseconds()),
		// Ino, Dev, Uid and Gid will be ignored and set to 0 for windows
		Ino:      uint32(0),
		Dev:      uint32(0),
		Uid:      uint32(0),
		Gid:      uint32(0),
		FileMode: parseFileMode(uint32(fileInfo.Mode())),
		FileSize: int32(fileInfo.Size()),
		Flags:    createFlag(false, false, filePath),
		ObjHash:  hash,
	}
	idxEntry := &IndexEntry{Metadata: entryMetdata, Name: filePath, V3Flags: nil}
	return idxEntry, nil
}

func ConvertNanosec(nsec int64) timePair {
	// Set whole number values of nsec to Sec and remaining fractional amount to Nsec
	secs := int32(nsec / 1000000000)
	nsecs := int32(nsec % 1000000000)
	return timePair{Sec: secs, Nsec: nsecs}
}
