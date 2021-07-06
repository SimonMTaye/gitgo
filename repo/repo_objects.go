// Functions for finding and saving objects to a repo
package repo

import (
    "strings"
    "os"
    "path"
    "github.com/SimonMTaye/gitgo/objects"
)

type ErrObjectNotFound struct {
    query string
}

func (e *ErrObjectNotFound) Error() string {
        return e.query + " is not a valid object name\n"
}

//TODO test
// Resolves name to a object hash-id
// Functions works in this order:
//     Check branch heads
//     Look for tags
//     Check for object-refs (treat name as the first few chars of a hash id)
// The last step requires that the hash id be at least 3 chars
func (repo *Repo) FindObject(name string) (string, error) {
    if name == "HEAD" {
        bytes, err := os.ReadFile(path.Join(repo.gitDir, "HEAD"))
        if err != nil {
            return "", err
        }
        read := string(bytes)
        if isRef(read) {
            obj, err := repo.FindRef(read)
            if err != nil {
                return "", err
            }
            return obj, nil
        } 
        return read, nil
    }
    refs, err := repo.GetAllRefs()
    if err != nil {
        return "", err
    }
    // Check branch heads
    obj, ok := refs["refs/heads/" + name]
    if ok {
        return obj, nil
    }
    // Check tags
    obj, ok = refs["refs/tags/" + name]
    if ok {
        return obj, nil
    }

    if len(name) < 3 {
        return "", &ErrObjectNotFound{query:name}
    }

    objectDirs, err := os.ReadDir(path.Join(repo.gitDir, "objects"))
    if err != nil {
        return "", nil
    }
    present := exists(objectDirs, name[:2])
    if !present {
        return "", &ErrObjectNotFound{query:name}
    }

    objs, err := os.ReadDir(path.Join(repo.gitDir, "objects", name[:2]))
    if err != nil {
        return "", err
    }

    matches := make([]string, 0, len(objs))
    for _, obj := range objs {
        if strings.HasPrefix(obj.Name(), name[2:]) {
            matches = append(matches, name[:2] + obj.Name())
        }
    }
    // We only want one match
    if len(matches) != 1 {
        return "", &ErrObjectNotFound{query:name}
    }
    return matches[0], nil
}

// TODO test
// Return a GitObject from a valid hash
// returns an error if the object is not found
func (repo *Repo) GetObject(objectHash string) (objects.GitObject, error)  {
    dir := path.Join(repo.gitDir, "objects", objectHash[:2])
    objs, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }
    for _, obj := range objs {
        if obj.Name() == objectHash[2:] {
            file, err := os.Open(path.Join(dir, obj.Name()))
            if err != nil {
                return nil, err
            }
            return objects.DecompressAndRead(file)
        }
    }
    return nil, &ErrObjectNotFound{query:objectHash}
}

// TODO Test
// Save a git object to the repo
func (repo *Repo) SaveObject(obj objects.GitObject) error {
    hash := objects.Hash(obj)
    // Check if the dir with the first two letters of the hash exists 
    // (eg. objects/0a where the hash is 0a32e1...)
    objectsDir := path.Join(repo.gitDir, "objects")
    entries, err := os.ReadDir(objectsDir)
    if err != nil {
        return err
    }

    hashDir := path.Join(objectsDir, hash[:2])
    // If the dir where the object will be stored doesn't exist, create it
    if !exists(entries, hash[:2]) {
        err = os.Mkdir(hashDir,  DIR_FILEMODE)
        if err != nil {
            return err
        }
    }

    // Check if the object already exists
    entries, err = os.ReadDir(hashDir)
    if err != nil {
        return err
    }
    if !exists(entries, hash[2:]) {
        file, err := os.Create(path.Join(hashDir, hash[2:]))
        if err != nil {
            return err
        }
        return objects.CompressAndSave(file, obj)
        
    }
    return nil
}
