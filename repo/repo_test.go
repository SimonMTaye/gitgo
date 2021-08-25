package repo

import (
    "testing"
    "os"
    "path"
    "github.com/SimonMTaye/gitgo/iniparse"
)



func TestFindRepo(t *testing.T) {
    tmpDir := t.TempDir()

    nestedPath := path.Join(tmpDir, "nest1", "nest2", "nest3")
    err := os.MkdirAll(nestedPath, DIR_FILEMODE)
    if err != nil {
        t.Fatalf("Error creating directories for testing\n")
    }

    err = os.Mkdir(path.Join(tmpDir, ".git"), DIR_FILEMODE)
    if err != nil {
        t.Fatalf("Error '.git' directory for testing\n")
    }

    repo1, err := FindRepo(tmpDir)
    if repo1 != tmpDir {
        t.Errorf("Wrong directory returned by FindRepo; Expected %s, Got %s\n", 
        tmpDir, repo1)
    }

    if err != nil {
        t.Errorf("Unexpected error returned by FindRepo:\n%s", err.Error())
    }

    repo2, err := FindRepo(nestedPath)
    if repo2 != tmpDir {
        t.Errorf("Wrong directory returned by FindRepo; Expected %s, Got %s\n", 
        tmpDir, repo2)
    }

    if err != nil {
        t.Errorf("Unexpected error returned by FindRepo:\n%s", err.Error())
    }
}

func TestFindRepoOnNoRepo(t *testing.T) {
    tmpDir := t.TempDir()

    nestedPath := path.Join(tmpDir, "nest1", "nest2", "nest3")
    err := os.MkdirAll(nestedPath, DIR_FILEMODE)
    if err != nil {
        t.Fatalf("Error creating directories for testing\n")
    }

    repo1, err := FindRepo(nestedPath)
    if repo1 != "" {
        t.Errorf("FindRepo returned an unexpected string:\n%s", repo1)
    }

    _, ok := err.(*ErrNoRepositoryFound)
    if !ok {
        if err != nil {
            t.Errorf("FindRepo returned an unexpected error:\n%s", err.Error())
        } else {
            t.Errorf("Expected FindRepo to return an error, but nothing was returned")
        }

    }
}


func TestOpenRepo(t *testing.T) {
    tmpDir := t.TempDir()
    err := CreateRepo(tmpDir, "", "random")
    if err != nil {
        t.Fatalf("Unexpected error when creating a new repository for testing:\n%s",
        err.Error())
    }

    repo, err := OpenRepo(tmpDir)
    if err != nil {
        t.Fatalf("Unexpected error when reading repository info for testing:\n%s",
        err.Error())
    }

    if repo.GitDir != path.Join(tmpDir, ".git") || len(repo.Branches) != 0 || repo.Worktree != "random" {
        t.Errorf("Repo object is different from expected.\n" +
        "Expected repoPath: %s, Got: %s\n" +
        "Expected worktree: %s, Got: %s\n" +
        "Expected no repo.Branches, Got: %d\n", 
        tmpDir, repo.GitDir, "random", repo.Worktree, len(repo.Branches))
    }

}


func TestOpenRepoWithBranches(t *testing.T) {
    tmpDir := t.TempDir()
    err := CreateRepo(tmpDir, "", "")
    if err != nil {
        t.Fatalf("Unexpected error when creating a new repository for testing:\n%s",
        err.Error())
    }
    // Open config file and modify it
    configPath := path.Join(tmpDir, ".git", "config")
    configFile, err := os.Open(configPath)
    if err != nil {
        t.Fatalf("Unexpected error when opening repository config file for testing:\n%s", 
        err.Error())
    }

    configIni, err := iniparse.ParseIni(configFile)
    if err != nil {
        t.Fatalf("Unexpected error when parsing config file for testing:\n%s",err.Error())
    }
    configFile.Close()

    configIni.SetProperty("branch \"main\"", "remote", "origin")
    configIni.SetProperty("branch \"main\"", "merge", "refs/heads/main")

    configFile, err = os.Create(configPath)
    if err != nil {
        t.Fatalf("Unexpected error when writing to repository config file for testing:\n%s", 
        err.Error())
    }

    configFile.WriteString(configIni.String())
    configFile.Close()

    repo, err := OpenRepo(tmpDir)
    if err != nil {
        t.Fatalf("Unexpected error when reading repository info for testing:\n%s",
        err.Error())
    }

    if repo.GitDir != path.Join(tmpDir, ".git") || len(repo.Branches) != 1 || repo.Worktree != tmpDir {
        t.Errorf("Repo object is different from expected.\n" +
        "Expected repoPath: %s, Got: %s\n" +
        "Expected worktree: %s, Got: %s\n" +
        "Expected repo branch: 1, Got: %d\n", 
        tmpDir, repo.GitDir, tmpDir, repo.Worktree, len(repo.Branches))
    }

    if repo.Branches[0].name != "main" {
        t.Errorf("Expected branch name to be 'main', Got: %s", repo.Branches[0].name)
    }
}

