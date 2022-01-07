package repo

import (
	"fmt"
	"github.com/SimonMTaye/gitgo/objects"
	"os"
	"path"
	"testing"
)

//Create a new temp directory. Init a new repo in that directory.
func createNewRepo(t *testing.T) (*Repo, error) {
	dir := t.TempDir()
	err := CreateRepo(dir, "", "")
	if err != nil {
		t.Fatalf("Error creating repo: %s", err)
	}
	return OpenRepo(dir)
}

func TestAdd(t *testing.T) {
	repoStruct, err := createNewRepo(t)
	if err != nil {
		t.Fatalf("Error opening repo: %s", err)
	}
	//Create a new file in the directory.
	file, err := os.Create(path.Join(repoStruct.Worktree, "test.txt"))
	if err != nil {
		t.Fatalf("Error creating file: %s", err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		t.Fatalf("Error getting file info: %s", err)
	}
	fileSize := fileInfo.Size()
	if err != nil {
		t.Fatalf("Error getting file size: %s", err)
	}
	// Use AddFile to add it to the repo
	err = repoStruct.AddFile("test.txt")
	if err != nil {
		t.Fatalf("Error adding file: %s", err)
	}
	idx, err := repoStruct.Index()
	if err != nil {
		t.Fatalf("Error getting index: %s", err)
	}
	//Test that the index is updated with the new file
	if ok, _ := idx.EntryExists("test.txt"); !ok {
		t.Errorf("Error: test.txt not in index")
	}
	//Test that blob representing the file exists
	blob, err := objects.FileBlob(path.Join(repoStruct.Worktree, "test.txt"))
	if err != nil {
		t.Fatalf("Error getting blob: %s", err)
	}
	hash := objects.Hash(blob)
	obj, err := repoStruct.GetObject(hash)
	if err != nil {
		t.Errorf("Error getting object: %s", err)
	}
	if err == nil {
		if obj.Type() != objects.Blob {
			t.Errorf("Error: object is not a blob")
		}
		if len(obj.Serialize()) != int(fileSize) {
			t.Errorf("Error: blob size is not correct")
		}
	}

}

// Add a file to the directory and make a commit.
// Test that the commit is added to the repo
// Test that the commit has the appropirate message
// Test that the tree created from the index matches what we expect
func TestCommit(t *testing.T) {
	repoStruct, err := createNewRepo(t)
	if err != nil {
		t.Fatalf("Error opening repo: %s", err)
	}
	_, err = os.Create(path.Join(repoStruct.Worktree, "test.txt"))
	if err != nil {
		t.Fatalf("Error creating file: %s", err)
	}
	// Use AddFile to add it to the repo
	err = repoStruct.AddFile("test.txt")
	if err != nil {
		t.Fatalf("Error adding file: %s", err)
	}
	for i := 0; i < 3; i++ {
		err := repoStruct.Commit(fmt.Sprintf("test commit %d", i))
		if err != nil {
			t.Fatalf("Error committing: %s", err)
		}
	}
	parent := "HEAD"
	for i := 2; i <= 0; i-- {
		hash, err := repoStruct.FindObject(parent)
		if err != nil {
			t.Fatalf("Error getting hash of object: %s", err)
		}
		obj, err := repoStruct.GetObject(hash)
		if err != nil {
			t.Fatalf("Error getting object: %s", err)
		}
		commit, ok := obj.(*objects.GitCommit)
		if !ok {
			t.Fatalf("Error: object is not a commit")
		}
		expectedMessage := fmt.Sprintf("test commit %d", i)
		if commit.Msg != expectedMessage {
			t.Errorf("Error: commit message is not correct.\n Expected: %s\n Got: %s", expectedMessage, commit.Msg)
		}
		parent = commit.ParentHash
	}
}
