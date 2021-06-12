package repo

import (
    "os"
    "github.com/SimonMTaye/gitgo/iniparse"
    "path/filepath"
)


const EMPTY_DESCRIPTION = "Unnamed repository; edit this file 'description' to name the repository."
//Unix permission bits
//Represents 001 - 111 - 111 - 111
//Or:          d - rwx - rwx - rwx
const DIR_FILEMODE = 1023


// Create the ".git" directory and the necessary files and dirs
// Will throw and error if ".git" already exists

// cwd: Current working directory where ".git" folder will be created
// description: repo description
// worktree: location for worktree
func CreateRepo(cwd string, description string, worktree string) error {
    gitDir, err := filepath.Abs(filepath.Join(cwd, ".git"))
    if err != nil {
        return nil
    }
    err = os.Mkdir(gitDir, DIR_FILEMODE)
    if err != nil {
        return err
    }

    err = os.Mkdir(filepath.Join(gitDir, "objects"), DIR_FILEMODE)
    if err != nil {
        return err
    }

    err = os.Mkdir(filepath.Join(gitDir, "branches"), DIR_FILEMODE)
    if err != nil {
        return err
    }

    refsDir := filepath.Join(gitDir, "refs")
    err = os.Mkdir(refsDir, DIR_FILEMODE)
    if err != nil {
        return err
    }

    err = os.Mkdir(filepath.Join(refsDir, "tags"), DIR_FILEMODE)
    if err != nil {
        return err
    }

    err = os.Mkdir(filepath.Join(refsDir, "heads"), DIR_FILEMODE)
    if err != nil {
        return err
    }

    head_file, err := os.Create(filepath.Join(gitDir, "HEAD"))
    defer head_file.Close()
    if err != nil {
        return err
    }

    config_file, err := os.Create(filepath.Join(gitDir, "config"))
    defer config_file.Close()
    if err != nil {
        return err
    }
    _, err = config_file.Write([]byte(defaultConfig(worktree)))
    if err != nil {
        return err
    }

    description_file, err := os.Create(filepath.Join(gitDir, "description"))
    defer description_file.Close()
    if err != nil {
        return err
    }

    if description != "" {
        _, err = description_file.Write([]byte(description))
    } else {
        _, err = description_file.Write([]byte(EMPTY_DESCRIPTION))
    }
    
    return err
}

func defaultConfig(worktree string) string {
    config := make(iniparse.IniFile)
    config.SetProperty("core", "repositoryformatversion", "0")
    config.SetProperty("core", "filemode", "false")
    config.SetProperty("core", "bare", "false")
    if worktree != ".." && worktree != "" {
        config.SetProperty("core", "worktree", worktree)
    }
    return config.String()
}



