//go:build linux || darwin

package repo

import (
	"crypto/sha1"
	"errors"
	"os"
	"path"
	"syscall"
)

// Contains code for reading a file a creating an entry in the index
// linux or darwin only. Windows code is found in index_entry_windows.go

// Read a file path and create an entry
func CreateEntry(repoPath string, filePath string) (*IndexEntry, error) {
	fileInfo, err := os.Stat(path.Join(repoPath, filePath))
	if err != nil {
		return nil, err
	}
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, errors.New("error getting 'stat' information for file: " +
			path.Join(repoPath, filePath))
	}
	fileBytes, err := os.ReadFile(path.Join(repoPath, filePath))
	if err != nil {
		return nil, err
	}
	hash := sha1.Sum(fileBytes)
	entryMetdata := &indexEntryMetadata{
		Ctime:    CovertTimespec(stat.Ctim),
		Mtime:    CovertTimespec(stat.Mtim),
		Ino:      uint32(stat.Ino),
		Dev:      uint32(stat.Dev),
		Uid:      stat.Uid,
		Gid:      stat.Gid,
		FileMode: parseFileMode(stat.Mode),
		FileSize: int32(fileInfo.Size()),
		Flags:    CreateFlag(false, false, filePath),
		ObjHash:  hash,
	}
	idxEntry := &IndexEntry{Metadata: entryMetdata, Name: filePath, V3Flags: nil}
	return idxEntry, nil
}

func CovertTimespec(timeSpec syscall.Timespec) TimePair {
	return TimePair{Sec: int32(timeSpec.Sec), Nsec: int32(timeSpec.Nsec)}
}
