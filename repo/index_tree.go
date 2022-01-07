package repo

import (
	"encoding/hex"
	"github.com/SimonMTaye/gitgo/index"
	"github.com/SimonMTaye/gitgo/objects"
	"strings"
)

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

//addEntry add an entry to the treemap
func (tm *treeMap) addEntry(entry *index.Entry) {
	mode := parseFileModeBits(entry.Metadata.FileMode)
	hash := hex.EncodeToString(entry.Metadata.ObjHash[:])
	pathlist := strings.Split(entry.Name, "/")
	tm.addEntryHelper(pathlist, hash, mode)
}

func (tm *treeMap) addEntryHelper(pathlist []string, hash string, mode objects.EntryFileMode) {
	if len(pathlist) > 1 {
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
func (tm *treeMap) toTree() *objects.GitTree {
	tree := &objects.GitTree{}
	// Add all the regular files
	for _, file := range tm.files {
		tree.AddEntry(file.mode, file.name, file.hash)
	}
	for name, subtree := range tm.subTrees {
		subGitTree := subtree.toTree()
		tree.AddEntry(objects.Directory, name, objects.Hash(subGitTree))
	}
	return tree
}

// allTrees Creates all the GitTree objects that represent the root dir as well as all the subdirs
// Necessary for saving them to Disk
func (tm *treeMap) allTrees() []*objects.GitTree {
	trees := make([]*objects.GitTree, 0)
	trees = append(trees, tm.toTree())
	for _, subtree := range tm.subTrees {
		trees = append(trees, subtree.allTrees()...)
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
		if index.BitSet32(FileMode, i+offset) {
			str += "1"
		} else {
			str += "0"
		}
	}
	if str == "1000" {
		// The last 9 bits of FileMode are permission. We are checking the
		// last bit (which corresponds to 'everyone' execution permission) and
		// 2 bits behind the last bit (which corresponds to 'user' execution permission)
		if index.BitSet32(FileMode, 31) || index.BitSet32(FileMode, 29) {
			return objects.Executable
		} else {
			return objects.Normal
		}
	} else {
		return objects.SymbolicLink
	}

}

func emptyTreeMap() *treeMap {
	return &treeMap{
		subTrees: make(map[string]*treeMap),
		files:    make([]*treeEntry, 0),
	}
}

// indexToTreeMap converts the index into a treeMap, which is  more convenient format for saving to disk
func indexToTreeMap(idx *index.Index) *treeMap {
	treeMap := emptyTreeMap()
	for _, entry := range idx.Entries {
		treeMap.addEntry(entry)
	}
	return treeMap
}
