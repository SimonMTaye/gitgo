package main

import (
	"github.com/SimonMTaye/gitgo/objects"
	"github.com/SimonMTaye/gitgo/repo"
	"os"
	"path"
)

// AddHelper Helper functions for the cli
// Add a file to the index as an index entry
func AddHelper(repodir string, filepath string) error {
	repoStruct, err := repo.OpenRepo(repodir)
	if err != nil {
		return err
	}
	// TODO Handle the index not being created yet
	index, err := repoStruct.Index()
	if err != nil {
		return err
	}
	entry, err := repo.CreateEntry(repoStruct.GitDir, filepath)
	if err != nil {
		return err
	}
	blob, err := objects.FileBlob(path.Join(repoStruct.GitDir, filepath))
	if err != nil {
		return err
	}
	err = repoStruct.SaveObject(blob)
	if err != nil {
		return err
	}
	// If the entry already exists, update it
	err = index.UpdateEntry(entry)
	if err != nil {
		return err
	}
	err = repoStruct.WriteIndex(index)
	if err != nil {
		return err
	}
	return nil
}

// CatfileHelper Find an object based on a search string
func CatfileHelper(repodir string, srchstr string) (objects.GitObject, error) {
	repoStruct, err := repo.OpenRepo(repodir)
	if err != nil {
		return nil, err
	}
	objHash, err := repoStruct.FindObject(srchstr)
	if err != nil {
		return nil, err
	}
	obj, err := repoStruct.GetObject(objHash)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// CommitHelper Create a commit based on the contents of the index and previous commit

// ParseObjectHelper Parse an object and return it for printing

// FindandOpenRepo Shortcut for finding and openeing a repo
func FindandOpenRepo() (*repo.Repo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	repoDir, err := repo.FindRepo(cwd)
	if err != nil {
		return nil, err
	}
	return repo.OpenRepo(repoDir)
}
