package repo

import (
	"errors"
	"github.com/SimonMTaye/gitgo/config"
	"github.com/SimonMTaye/gitgo/objects"
	"path"
)

// Used for constructing the trees from the index

// Commit Creates a new commit. Writes the index file to a tree and then creates a new commit object with the default author and comitter.
func (repo *Repo) Commit(msg string) error {
	idx, err := repo.Index()
	if err != nil {
		return err
	}
	if idx.IsEmpty() {
		return errors.New("index is empty, there is nothing to commit")
	}
	treeMap := indexToTreeMap(idx)
	trees := treeMap.allTrees()
	for _, tree := range trees {
		// Save all the tree objects that will be referenced in our commit
		err := repo.SaveObject(tree)
		if err != nil {
			return err
		}
	}
	configs, err := config.LoadConfig(repo.GitDir)
	if err != nil {
		return err
	}
	headHash, err := repo.FindObject("HEAD")
	if err != nil {
		return err
	}
	user := (*configs.All)["user"]["name"]
	email := (*configs.All)["user"]["email"]
	commit := &objects.GitCommit{}
	// Create commit object
	// Root tree should be at the beginning, needs to be checked
	commit.TreeHash = objects.Hash(trees[0])
	commit.Msg = msg
	commit.SetAuthor(user, email)
	// TODO write function that takes committer information from user
	commit.SetCommitter(user, email)
	commit.ParentHash = headHash
	err = repo.SaveObject(commit)
	if err != nil {
		return err
	}
	// Update the current branch/head to point to our commit
	return repo.UpdateCurrentBranch(objects.Hash(commit))
}

// AddFile adds a file entry to the index
func (repo *Repo) AddFile(filepath string) error {
	blob, err := objects.FileBlob(path.Join(repo.GitDir, filepath))
	if err != nil {
		return err
	}
	// Save Object handles duplicates so we don't need to check for it
	err = repo.SaveObject(blob)
	if err != nil {
		return err
	}
	// Update the staging area with the new file
	idx, err := repo.Index()
	err = idx.AddFile(repo.Worktree, filepath)
	if err != nil {
		return err
	}
	err = repo.WriteIndex(idx)
	if err != nil {
		return err
	}
	return nil
}
