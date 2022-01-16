package objects

import (
	"encoding/hex"
	"testing"
)

// Uses info from an actual git Tree object for test
// tests addEntry, Serialize, Size and that the byte layout is correct
func TestTree(t *testing.T) {
	tree := &GitTree{}
	tree.AddEntry(Normal, "fullini_test.go", "cc09a49865b0f9a62039b7155a20f9f29af312b4")
	tree.AddEntry(Normal, "parseline_test.go", "5b4c1b1891bc002cce7b22c2935257cccec9ef2e")
	tree.AddEntry(Normal, "read.go", "f08b2de690a4661751f5bfa60f396a3656563654")
	tree.AddEntry(Normal, "write.go", "ad5ac59ff07840e75684b4700a2eada5086b1308")

	if tree.size != 159 {
		t.Errorf("Expected tree size to be '159', Got: %d", tree.size)
	}

	expectedHash := "1335d19337aa47bb0b0ff5e8444a65cd8d63584c"
	if Hash(tree) != expectedHash {
		t.Errorf("Expected hash: %s\nGot: %s", expectedHash, Hash(tree))
	}
}

// Test that an entry object can be correctly constructed from bytes
func TestEntryConstruction(t *testing.T) {
	tree := &GitTree{}
	tree.AddEntry(Normal, "fullini_test.go", "cc09a49865b0f9a62039b7155a20f9f29af312b4")
	tree.AddEntry(Normal, "parseline_test.go", "5b4c1b1891bc002cce7b22c2935257cccec9ef2e")
	tree.AddEntry(Normal, "read.go", "f08b2de690a4661751f5bfa60f396a3656563654")
	tree.AddEntry(Normal, "write.go", "ad5ac59ff07840e75684b4700a2eada5086b1308")
	treeBytes := tree.Serialize()

	testEntry := byteToEntry(treeBytes[0:43])

	if testEntry.mode != Normal {
		t.Errorf("Expected 'Normal' file mode, Got: %s", testEntry.mode)
	}
	if testEntry.name != "fullini_test.go" {
		t.Errorf("Expected the name 'fullini_test.go', Got: %s", testEntry.name)
	}
	expectedHash := "cc09a49865b0f9a62039b7155a20f9f29af312b4"
	if hex.EncodeToString(testEntry.hash) != expectedHash {
		t.Errorf("Expected hash: %s\nGot: %s", expectedHash, hex.EncodeToString(testEntry.hash))
	}
}

// Test that the tree serialization and deserialization process is correct and that a new tree can be
// constructed from the byte slice returned from serialization
func TestTreeSerializing(t *testing.T) {
	tree := &GitTree{}
	tree.AddEntry(Normal, "fullini_test.go", "cc09a49865b0f9a62039b7155a20f9f29af312b4")
	tree.AddEntry(Normal, "parseline_test.go", "5b4c1b1891bc002cce7b22c2935257cccec9ef2e")
	tree.AddEntry(Normal, "read.go", "f08b2de690a4661751f5bfa60f396a3656563654")
	tree.AddEntry(Normal, "write.go", "ad5ac59ff07840e75684b4700a2eada5086b1308")

	newTree := &GitTree{}
	newTree.Deserialize(tree.Serialize())

	if newTree.size != 159 {
		t.Errorf("Expected tree size to be '159', Got: %d", newTree.size)
	}

	if Hash(tree) != Hash(newTree) {
		t.Errorf("Expected hash: %s\nGot:%s", Hash(tree), Hash(newTree))
	}

}
