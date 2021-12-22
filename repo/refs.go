// Functions for finding and resolving refs in a repo
package repo

import (
	"os"
	"path"
	"strings"

	"github.com/SimonMTaye/gitgo/objects"
)

type ErrTagAlreadyExists struct {
	name string
}

func (e *ErrTagAlreadyExists) Error() string {
	return "refs/tags/" + e.name + " already exists"
}

// Plumbing function; find the hash a 'ref' refers too;
// usually this is simply finding the ref file and reading the hash
// The function may also recursively call itself in the case where refs point to other
// refs
func readRef(gitDir string, refPath string) (string, error) {
	refData, err := os.ReadFile(path.Join(gitDir, refPath))
	if err != nil {
		return "", err
	}
	ref := string(refData)

	// This ref is a ref to another ref
	if isRef(ref) {
		newRefPath := strings.TrimPrefix(ref, "ref: ")
		return readRef(gitDir, newRefPath)
	}
	return ref, nil
}

// Checks whether the passed string is a ref or not
// Helper function for clarity
func isRef(unknown string) bool {
	return strings.HasPrefix(unknown, "ref: ")
}

// Find all the refs in a repo and the hashes of what they are pointing to
func findAllRefs(gitDir string) (map[string]string, error) {
	refs, err := recursiveFindFiles(gitDir, "refs")
	if err != nil {
		return nil, err
	}
	refMap := make(map[string]string)
	for _, refPath := range refs {
		ref, err := readRef(gitDir, refPath)
		if err != nil {
			refMap[refPath] = "Error reading ref"
		} else {
			refMap[refPath] = ref
		}
	}
	return refMap, nil
}

// Returns the name of all files in root/subpath and nested directories prefixed with subpath
// for example, the dir /root/refs/heads which contains main and /branch/feature and where
// the root is /root would return refs/heads/main and refs/head/feature
func recursiveFindFiles(root string, subpath string) ([]string, error) {
	entries, err := os.ReadDir(path.Join(root, subpath))
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, 10)

	for _, entry := range entries {
		// If error is found in the recursive calls, the list so far and the error is returned
		if entry.IsDir() {
			newSubpath := path.Join(subpath, entry.Name())
			recursiveRefs, err := recursiveFindFiles(root, newSubpath)
			files = append(files, recursiveRefs...)
			// Return results so far if error is nil
			if err != nil {
				return files, err
			}
		} else {
			files = append(files, path.Join(subpath, entry.Name()))
		}
	}
	return files, nil
}

// Find a ref in a repo. Simply calls the readRef function defined in refs.go
func (repo *Repo) FindRef(refPath string) (string, error) {
	if isRef(refPath) {
		return readRef(repo.GitDir, strings.TrimPrefix(refPath, "ref: "))
	}
	return readRef(repo.GitDir, refPath)
}

// Return a map off all refs and what they point to as a key-value pair, respectively.
// Calls findAllRefs defined in refs.go
func (repo *Repo) GetAllRefs() (map[string]string, error) {
	return findAllRefs(repo.GitDir)
}

// Saves a tag with the given 'name' that points the object of the corresponding hash
// This function does not verify that the hash is valid, that is the caller's responsibility
// An error is thrown if the tag already exists
func (repo *Repo) SaveTag(name string, hash string) error {
	tagsDir := path.Join(repo.GitDir, "refs", "tags")
	// Check that the tag doesn't already exist
	entries, err := os.ReadDir(tagsDir)
	if err != nil {
		return err
	}
	if exists(entries, name) {
		return &ErrTagAlreadyExists{name: name}
	}
	file, err := os.Create(path.Join(tagsDir, name))
	if err != nil {
		return err
	}
	_, err = file.WriteString(hash)
	defer file.Close()
	return err
}

// Deletes a tag from the list of tags.
// If the tag points to a tag object, delete that too
func (repo *Repo) DeleteTag(name string) error {
	tagsDir := path.Join(repo.GitDir, "refs", "tags")
	entries, err := os.ReadDir(tagsDir)
	if err != nil {
		return err
	}
	if exists(entries, name) {
		tagPath := path.Join(tagsDir, name)
		contents, err := os.ReadFile(tagPath)
		if err != nil {
			return err
		}
		obj, err := repo.GetObject(string(contents))
		if err != nil {
			return err
		}
		// If the tag reference points to a tag object, delete the object too
		if obj.Type() == objects.Tag {
			err = repo.DeleteObject(objects.Hash(obj))
			if err != nil {
				return err
			}
		}
		return os.Remove(tagPath)
	} else {
		// Tag doesn't exist
		return &ErrObjectNotFound{query: name}
	}
}

// Update a branch ref to a new hash
func (repo *Repo) updateBranchRef(branch string, hash string) error {
	branchRef := path.Join(repo.GitDir, "refs", "heads", branch)
	return os.WriteFile(branchRef, []byte(hash), NORMAL_FILEMODE)
}
