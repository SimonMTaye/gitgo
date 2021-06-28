package repo

import ( 
    "os"
    "path"
    "strings"
    "github.com/SimonMTaye/gitgo/iniparse"
    )


type Repo struct {
    gitDir string
    worktree string
    branches []Branch
}

type Branch struct {
    name string
}

// Indicates that the given directory does not contain a repository
type ErrNoRepository struct {
    dir string
}

// Indicates that a repository could not be found in the current dir or any of its parents
type ErrNoRepositoryFound struct {
    dir string
}


func (e ErrNoRepository) Error() string {
    return "Directory does not contain a repository: "+ e.dir
}

func (e ErrNoRepositoryFound) Error() string {
    return "Could not find repository in the directory or its parents: " + e.dir
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
        return nil, ErrNoRepository{dir: dir}
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

    repo := Repo{ gitDir: gitDir, branches: branches}

    worktree, ok := configIni["core"]["worktree"]
    if ok {
        repo.worktree = worktree
    } else {
        repo.worktree = dir
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
    return "", ErrNoRepositoryFound{dir: cwd}
}

// Read Branch information from the config file
// TODO if atleast one  branch is a necessity, then this function should return an error
// Else, return an empty slice when there are no branches (this is current behavior)
func getBranchesFromConfigIni (configIni *iniparse.IniFile) ([]Branch, error) {
    branches := make([]Branch, 0)
    for sections := range *configIni {
        if strings.HasPrefix(sections, "branch") {
            branches = append(branches, Branch {name: strings.Trim(strings.Split(sections, " ")[1], "\"")})
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
