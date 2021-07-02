// Functions for finding and resolving refs in a repo
package repo

import (
    "os"
    "strings"
    "path"
    )


// Plumbing function; find the hash a 'ref' refers too;
// usually this is simply finding the ref file and reading the hash
// The function may also recursively call itself in the case where refs point to other
// refs
func readRef(gitDir string, refPath string) (string, error) {
    refData, err := os.ReadFile(path.Join(gitDir,  refPath))
    if err != nil {
        return "", err
    }
    ref := string(refData)

    // This ref is a ref to another ref
    if isRef(ref) {
        newRefPath := strings.TrimPrefix(ref, "ref: ")
        return readRef(gitDir, newRefPath)
    }
    return string(ref), nil
}

// Checks whether the passed string is a ref or not
// Helper functino for clarity
func isRef(unknown string) bool {
    return strings.HasPrefix(unknown, "ref: ") 
}

// Find all the refs in a repo and the hashes of what they are pointing to
func findAllRefs (gitDir string) (map[string]string, error) {
    refs, err := recursiveFindFiles(gitDir, ".")
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
            newSubpath :=  path.Join(subpath, entry.Name())
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
    return readRef(repo.gitDir, refPath)
}

// Return a map off all refs and what they point to as a key-value pair, respectively. 
// Calls findAllRefs defined in refs.go
func (repo *Repo) GetAllRefs() (map[string]string, error) {
    return findAllRefs(repo.gitDir)
}
