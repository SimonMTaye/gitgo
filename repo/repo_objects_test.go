package repo

import (
	"github.com/SimonMTaye/gitgo/objects"
	"os"
	"path"
	"testing"
)

// Test saving objects in a repo
func TestSaveObjects(t *testing.T) {
	tmpDir := t.TempDir()
	err := CreateRepo(tmpDir, "", "")
	if err != nil {
		t.Fatalf("Unexpected error when creating repo:\n%s", err.Error())
	}
	repo, err := OpenRepo(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error when opening repo:\n%s", err.Error())
	}
	blob := &objects.GitBlob{}
	blob.Deserialize([]byte("hello world"))
	blobHash := objects.Hash(blob)

	err = repo.SaveObject(blob)
	if err != nil {
		t.Fatalf("Unexpected error when saving object:\n%s", err.Error())
	}

	file, err := os.Open(path.Join(tmpDir, ".git", "objects", blobHash[:2], blobHash[2:]))
	if err != nil {
		t.Fatalf("Unexpected error when opening object file:\n%s", err.Error())
	}

	readBlob, err := objects.DecompressAndRead(file)
	if err != nil {
		t.Fatalf("Unexpected error when reading object file:\n%s", err.Error())
	}

	newHash := objects.Hash(readBlob)

	if newHash != blobHash {
		t.Errorf("Expected object read from disk to have the same hash as original"+
			"object. Object read from disk:\n%s", readBlob.String())
	}
}

// Test reading and finding objects
// Same procedure as the SaveObject test but reads the on-disk object using
// repo.GetObject() function instead of manually determining the path
func TestReadObject(t *testing.T) {
	tmpDir := t.TempDir()
	err := CreateRepo(tmpDir, "", "")
	if err != nil {
		t.Fatalf("Unexpected error when creating repo:\n%s", err.Error())
	}
	repo, err := OpenRepo(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error when opening repo:\n%s", err.Error())
	}
	blob := &objects.GitBlob{}
	blob.Deserialize([]byte("hello world"))
	blobHash := objects.Hash(blob)

	err = repo.SaveObject(blob)
	if err != nil {
		t.Fatalf("Unexpected error when saving object:\n%s", err.Error())
	}

	readBlob, err := repo.GetObject(blobHash)
	if err != nil {
		t.Fatalf("Unexpected error when reading object file:\n%s", err.Error())
	}

	newHash := objects.Hash(readBlob)

	if newHash != blobHash {
		t.Errorf("Expected object read from disk to have the same hash as original"+
			"object. Object read from disk:\n%s", readBlob.String())
	}
}

// Test that the FindObject functions works on partial hashes and the order in which it
// reads objects
func TestFindObject(t *testing.T) {
	tmpDir := t.TempDir()
	err := CreateRepo(tmpDir, "", "")
	if err != nil {
		t.Fatalf("Unexpected error when creating repo:\n%s", err.Error())
	}
	repo, err := OpenRepo(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error when opening repo:\n%s", err.Error())
	}
	blob := &objects.GitBlob{}
	blob.Deserialize([]byte("hello world"))
	blobHash := objects.Hash(blob)

	err = repo.SaveObject(blob)
	if err != nil {
		t.Fatalf("Unexpected error when saving object:\n%s", err.Error())
	}

	err = repo.SaveTag("test", blobHash)
	if err != nil {
		t.Fatalf("Unexpected error when saving tag:\n%s", err.Error())
	}
	// Check that tags are searched
	tagSearchHash, err := repo.FindObject("test")
	if err != nil {
		t.Fatalf("Unexpected error when searching for tag:\n%s", err.Error())
	}

	if tagSearchHash != blobHash {
		t.Errorf("Expected returned tag hash to be: %s\nGot: %s", blobHash, tagSearchHash)
	}

	// Check that objects can be found with the first few letters of the hash
	objSearchHash, err := repo.FindObject(blobHash[:4])
	if err != nil {
		t.Fatalf("Unexpected error when searching for object:\n%s", err.Error())
	}

	if objSearchHash != blobHash {
		t.Errorf("Expected returned object hash to be: %s\nGot: %s", blobHash, objSearchHash)
	}

	headsDir := path.Join(repo.GitDir, "refs", "heads")
	file, err := os.Create(path.Join(headsDir, "test"))

	if err != nil {
		t.Fatalf("Unexpected error when creating file 'test' in refs/heads:\n%s", err.Error())
	}

	randomString := "hello world"
	_, err = file.WriteString(randomString)
	if err != nil {
		t.Fatalf("Unexpected error when writing to file refs/heads/test:\n%s", err.Error())
	}
	file.Close()

	headSearch, err := repo.FindObject("test")
	if err != nil {
		t.Fatalf("Unexpected error when searching for head object:\n%s", err.Error())
	}

	if headSearch != randomString {
		t.Errorf("Expected returned head hash to be: %s\nGot: %s", randomString, headSearch)
	}
}
