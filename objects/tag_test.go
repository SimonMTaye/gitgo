package objects

import (
    "testing"
)

// Test that tags are serialzed correctly, and that size of hashes are computed accurately
func TestTag (t *testing.T) {
    tag := &GitTag{}
    tag.SetTaggerAndTime("Simon Taye", "mulat.simon@gmail.com", 1625088346, 0)
    tag.tagName = "test"
    tag.msg = "Test tag"
    tag.SetObject(Commit, "369fb3f5db3baf1b96032979af0cae946d4fd134")

    expectedHash := "83bf8e887fe355870191c95a4ed1dc106db81d29"
    hash := Hash(tag)
    size := tag.computeSize()

    if size != 138 {
        t.Errorf("Expected tag size to be 138, Got: %d\nTag Object:\n%s", size, tag.String())
    }

    if hash != expectedHash {
        t.Errorf("Expected hash to be:\n%s\n Got:\n %s\nTag Object:\n%s ",
            expectedHash,
            hash,
            tag.String())
    }
}

// Test that a tag object can be recreated from a well formed byte source
func TestTagDeserialize(t *testing.T) {
    sampleTag := &GitTag{}
    sampleTag.SetTaggerAndTime("Simon Taye", "mulat.simon@gmail.com", 1625088346, 0)
    sampleTag.tagName = "test"
    sampleTag.msg = "Test tag"
    sampleTag.SetObject(Commit, "369fb3f5db3baf1b96032979af0cae946d4fd134")

    tag := &GitTag{}
    tag.Deserialize(sampleTag.Serialize())

    expectedHash := "83bf8e887fe355870191c95a4ed1dc106db81d29"
    hash := Hash(tag)
    size := tag.computeSize()

    if size != 138 {
        t.Errorf("Expected tag size to be 138, Got: %d\nTag Object:\n%s", size, tag.String())
    }

    if hash != expectedHash {
        t.Errorf("Expected hash to be:\n%s\n Got:\n %s\nTag Object:\n%s ",
            expectedHash,
            hash,
            tag.String())
        }
    
}
