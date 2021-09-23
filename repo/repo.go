// Functions for finding and opening a repo and creating a repo struct
package repo

import ( 
    "os"
    "path"
    "strings"
    "errors"
    "github.com/SimonMTaye/gitgo/iniparse"
    )


type Repo struct {
    GitDir string
    Worktree string
    Branches []Branch
    detached bool
}

type Branch struct {
    name string
    ref string
}
// Indicates that the given directory does not contain a repository
type ErrNoRepository struct {
    dir string
}

// Indicates that a repository could not be found in the current dir or any of its parents
type ErrNoRepositoryFound struct {
    dir string
}

func (e *ErrNoRepository) Error() string {
    return "Directory does not contain a repository: "+ e.dir
}

func (e *ErrNoRepositoryFound) Error() string {
    return "Could not find repository in the directory or its parents: " + e.dir
}

// Checks the current directory for a ".git" directory, returns an error if it is not found
// Reads the "config" file in the ".git" directory and returns a Repo struct with the 
// current repos properties
func OpenRepo (dir string) (*Repo, error) {
    // Look for the ".git" directory
    dirs, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }

    if !exists(dirs, ".git") {
        return nil, &ErrNoRepository{dir: dir}
    }

    // Open the "config" file
    gitDir := path.Join(dir, ".git")
    configFile, err := os.Open(path.Join(gitDir, "config"))
    if err != nil {
        return nil, err
    }
    // Parse the config file
    configIni, err := iniparse.ParseIni(configFile)
    if err != nil {
        return nil, err
    }

    branches, err := getBranchesFromConfigIni(&configIni)
    if err != nil {
        return nil, err
    }

    repo := Repo{ GitDir: gitDir, Branches: branches}

    worktree, ok := configIni["core"]["worktree"]
    if ok {
        repo.Worktree = worktree
    } else {
        repo.Worktree = dir
    }
    return &repo, nil
    
}

// Recursively checks parent directory until a ".git" is found or "/" is reached
func FindRepo (cwd string) (string, error) {
    curDir := cwd
    for curDir != "/" {
        dirs, err := os.ReadDir(curDir)
        if err != nil {
            return "", err
        }

        if exists(dirs, ".git") {
            return curDir, nil
        }

        curDir = path.Join(curDir, "..")
    }
    return "", &ErrNoRepositoryFound{dir: cwd}
}

// Read Branch information from the config file
// TODO if atleast one  branch is a necessity, then this function should return an error
// Else, return an empty slice when there are no branches (this is current behavior)
func getBranchesFromConfigIni (configIni *iniparse.IniFile) ([]Branch, error) {
    branches := make([]Branch, 0)
    for section := range *configIni {
        if strings.HasPrefix(section, "branch") {
            // branch names are stored as 'branch [name]' in config file, this code removes
            // that
            // ref to hash is stored in the merge property of the branch section
            branches = append(branches, Branch {name: strings.Trim(strings.Split(section, " ")[1], "\""), ref: "ref: " + (*configIni)[section]["merge"]})
        }
    }
    return branches, nil
}

// Helper function to iterate over list of diretories/files and check for certain names
func exists (entries []os.DirEntry, name string) bool {
    for _, entry := range entries {
        if entry.Name() == name {
            return true
        }
    }
    return false
}
// Parse the index file of repo and return a struct representing the staging area
func (repo *Repo) Index() (*Index, error) {
    indexFile, err := os.Open(path.Join(repo.GitDir, "index"))
    if err != nil {
        return nil, err
    }
    return ParseIndex(indexFile)
}
// Write an Index struct to the index file of the repo
func (repo *Repo) WriteIndex(index *Index) error {
    indexFile, err := os.Open(path.Join(repo.GitDir, "index"))
    if err != nil {
        return err
    }
    indexBytes := index.ToBytes()
    n, err := indexFile.Write(indexBytes)
    if err != nil {
        return err
    }
    if n != len(indexBytes) {
        return errors.New("The bytes written to the index file are inconsistent")
    }
    return nil
}
// Updates the current branch to point to the new hash. If there is no branch (i.e. HEAD
// is detached) then HEAD will now point to the new hash.
func (repo *Repo) UpdateCurrentBranch(hash string)  error {
    headPath := path.Join(repo.GitDir, "HEAD")
    data, err := os.ReadFile(headPath)
    if err != nil {
        return err
    }
    contents := string(data)
    if isRef(contents) {
        return repo.updateBranchRef(hash) 

    } else {
        return os.WriteFile(headPath, []byte(hash), NORMAL_FILEMODE)
    }

}


