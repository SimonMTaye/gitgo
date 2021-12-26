package main

import (
	"github.com/SimonMTaye/gitgo/objects"
	"github.com/SimonMTaye/gitgo/repo"
	"os"
)

// AddHelper Helper functions for the cli
// Add a file to the index as an index entry
func AddHelper(repodir string, filepath string) error {
	repoStruct, err := repo.OpenRepo(repodir)
	if err != nil {
		return err
	}
	return repoStruct.AddFile(filepath)
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
