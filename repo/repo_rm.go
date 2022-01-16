package repo

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/SimonMTaye/gitgo/objects"
)

func (repo *Repo) Rm(file string) error {

	idx, err := repo.Index()
	if err != nil {
		return err
	}

	exists, pos := idx.EntryExists(file)
	if exists {
		hash, err := repo.getFileLastHash(file)
		// If the error is not real, we assume the file does not exist in the previous commits tree, thus the rm
		// file deletes the file from the index
		if err != nil {
			fmt.Println(err)
			err = idx.DeleteEntry(pos)
			if err != nil {
				return err
			}
			err = repo.WriteIndex(idx)
			if err != nil {
				return err
			}
			return nil
		}
		// If the file does exist, revert the hash
		hashBytes, err := hashToBytes(hash)
		if err != nil {
			return err
		}
		err = idx.ModifyFileHash(file, hashBytes)
		return err

	} else {
		return errors.New(fmt.Sprintf("File %s not found in index", file))
	}
}

func (repo *Repo) getFileLastHash(file string) (string, error) {
	headHash, err := repo.FindObject("HEAD")
	if err != nil {
		return "", err
	}
	obj, err := repo.GetObject(headHash)
	if err != nil {
		fmt.Println("Could not open head")
		return "", err
	}
	commit, ok := obj.(*objects.GitCommit)
	if !ok {
		return "", errors.New("HEAD is not a commit")
	}
	treeObj, err := repo.GetObject(commit.TreeHash)
	if err != nil {
		fmt.Println("Could not open head commit")
		return "", err
	}
	tree, ok := treeObj.(*objects.GitTree)
	if !ok {
		return "", errors.New(fmt.Sprintf("commit tree hash does not point to tree\nHash: %s", commit.TreeHash))
	}
	return tree.GetEntryHash(file)
}

func hashToBytes(hash string) (*[20]byte, error) {
	hashBytes := new([20]byte)
	n, err := hex.Decode(hashBytes[:], []byte(hash))
	if err != nil {
		return nil, err
	}
	if n != 20 {
		return nil, fmt.Errorf("%d characters were read, but 20 were expected", n)
	}
	return hashBytes, nil
}
