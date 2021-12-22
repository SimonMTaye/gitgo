package repo

import (
	"encoding/hex"
	"errors"
	"github.com/SimonMTaye/gitgo/config"
	"github.com/SimonMTaye/gitgo/objects"
	"strings"
)

// Used for constructing the trees from the index
type treeEntry struct {
	mode objects.EntryFileMode
	name string
	hash string
}

// Represents the overall tree structure that will be created
type treeMap struct {
	subTrees map[string]*treeMap
	files    []*treeEntry
}

// AddEntry TODO TreeMap generation needs to be tested
// Add an entry to the treeMap
func (tm *treeMap) AddEntry(entry *IndexEntry) {
	mode := parseFileModeBits(entry.Metadata.FileMode)
	hash := hex.EncodeToString(entry.Metadata.ObjHash[:])
	pathlist := strings.Split(entry.Name, "/")
	tm.addEntryHelper(pathlist, hash, mode)
}

func (tm *treeMap) addEntryHelper(pathlist []string, hash string, mode objects.EntryFileMode) {
	if len(pathlist) >= 1 {
		subtree, ok := tm.subTrees[pathlist[0]]
		if !ok {
			tm.subTrees[pathlist[0]] = &treeMap{}
			subtree = tm.subTrees[pathlist[0]]
		}
		subtree.addEntryHelper(pathlist[1:], hash, mode)
	} else {
		tm.files = append(tm.files, &treeEntry{mode, pathlist[0], hash})
	}
}

// Convert a treeMap into a regular GitTree
func (tm *treeMap) ToTree() *objects.GitTree {
	tree := &objects.GitTree{}
	// Add all the regular files
	for _, file := range tm.files {
		tree.AddEntry(file.mode, file.name, file.hash)
	}
	for name, subtree := range tm.subTrees {
		subGitTree := subtree.ToTree()
		tree.AddEntry(objects.Directory, name, objects.Hash(subGitTree))
	}
	return tree
}

// AllTrees Creates all the GitTree objects that represent the root dir as well as all the subdirs
// Necessary for saving them to Disk
func (tm *treeMap) AllTrees() []*objects.GitTree {
	// Returns the tree object form of the subtree and every sub tree
	// TODO This function must exist but calling this and ToTree is incredibly
	// inefficent and unneccessary; especially since both will always have to called
	// This will be left here now but it should be changed
	trees := make([]*objects.GitTree, 10)
	trees = append(trees, tm.ToTree())
	for _, subtree := range tm.subTrees {
		trees = append(trees, subtree.AllTrees()...)
	}
	return trees
}

// Convert the file mode bits stored in the uint32 into a string required by tree
// objects
// BREAKS ENCAPSULATION, this should be handled by trees themselves
func parseFileModeBits(FileMode uint32) objects.EntryFileMode {
	// bit 0-15 are empty
	offset := 15
	str := ""
	for i := 0; i < 4; i++ {
		if bitSet32(FileMode, i+offset) {
			str += "1"
		} else {
			str += "0"
		}
	}
	if str == "1000" {
		// The last 9 bits of FileMode are permission. We are checking the
		// last bit (which corresponds to 'everyone' execution permission) and
		// 2 bits behind the last bit (which corresponds to 'user' execution permission)
		if bitSet32(FileMode, 31) || bitSet32(FileMode, 29) {
			return objects.Executable
		} else {
			return objects.Normal
		}
	} else {
		return objects.SymbolicLink
	}

}

// Converts the index into a treeMap, which is  more convenient format for saving to disk
// TODO maybe the index should be parsed into this as a property of repo objects and not
// just when commiting?
func IndexToTreeMap(idx *Index) *treeMap {
	treeMap := &treeMap{}
	for _, entry := range idx.Entries {
		treeMap.AddEntry(entry)
	}
	return treeMap
}

// Commit Creates a new commit. Writes the index file to a tree and then creates a new commit object with the default author and comitter.
func (repo *Repo) Commit(msg string) error {
	idx, err := repo.Index()
	if err != nil {
		return err
	}
	if idx.IsEmpty() {
		return errors.New("index is empty, there is nothing to commit")
	}
	treeMap := IndexToTreeMap(idx)
	trees := treeMap.AllTrees()
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
