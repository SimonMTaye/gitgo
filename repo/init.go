package repo

import (
	"os"
	"path/filepath"

	"github.com/SimonMTaye/gitgo/config"
	"github.com/SimonMTaye/gitgo/iniparse"
)

const EmptyDescription = "Unnamed repository; edit this file 'description' to name the repository."

// DirFilemode Unix permission bits
//Represents 001 - 111 - 111 - 111
//Or:          d - rwx - rwx - rwx
const DirFilemode = 1023

// NormalFilemode Represents 000 - 110 - 110 - 100
//Or:          d - rwx - rwx - rwx
const NormalFilemode = 436
const DefaultBranchName = "main"

// CreateRepo Create the ".git" directory and the necessary files and dirs
// Will throw and Error if ".git" already exists
// cwd: Current working directory where ".git" folder will be created
// description: repo description
// worktree: location for worktree
func CreateRepo(cwd string, description string, worktree string) error {
	gitDir, err := filepath.Abs(filepath.Join(cwd, ".git"))
	if err != nil {
		return nil
	}
	err = os.Mkdir(gitDir, DirFilemode)
	if err != nil {
		return err
	}

	err = os.Mkdir(filepath.Join(gitDir, "objects"), DirFilemode)
	if err != nil {
		return err
	}

	err = os.Mkdir(filepath.Join(gitDir, "branches"), DirFilemode)
	if err != nil {
		return err
	}

	refsDir := filepath.Join(gitDir, "refs")
	err = os.Mkdir(refsDir, DirFilemode)
	if err != nil {
		return err
	}

	err = os.Mkdir(filepath.Join(refsDir, "tags"), DirFilemode)
	if err != nil {
		return err
	}

	err = os.Mkdir(filepath.Join(refsDir, "heads"), DirFilemode)
	if err != nil {
		return err
	}
	// Create config file
	configFile, err := os.Create(filepath.Join(gitDir, "config"))
	if err != nil {
		return err
	}
	defer func(configFile *os.File) {
		err := configFile.Close()
		if err != nil {
		}
	}(configFile)

	_, err = configFile.Write([]byte(defaultConfig(worktree)))
	if err != nil {
		return err
	}
	// Get default branch name; use hard coded value if not available
	configData, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}
	initBranchName := DefaultBranchName
	initSection, ok := (*configData.All)["init"]
	if ok {
		branchName, ok := initSection["defaultBranch"]
		if ok {
			initBranchName = branchName
		}
	}
	// Create main branch ref
	_, err = os.Create(filepath.Join(refsDir, "heads", initBranchName))
	if err != nil {
		return err
	}
	headFile, err := os.Create(filepath.Join(gitDir, "HEAD"))
	// Set new branch as head
	_, err = headFile.WriteString("ref: refs/heads/" + initBranchName + "\n")
	if err != nil {
		return err
	}
	defer func(headFile *os.File) {
		err := headFile.Close()
		if err != nil {
		}
	}(headFile)
	if err != nil {
		return err
	}
	descriptionFile, err := os.Create(filepath.Join(gitDir, "description"))
	if err != nil {
		return err
	}
	defer func(descriptionFile *os.File) {
		err := descriptionFile.Close()
		if err != nil {
		}
	}(descriptionFile)

	if description != "" {
		_, err = descriptionFile.Write([]byte(description))
	} else {
		_, err = descriptionFile.Write([]byte(EmptyDescription))
	}

	return err
}

//Returns a string representation of the default config file used for .git directories
func defaultConfig(worktree string) string {
	configIni := make(iniparse.IniFile)
	configIni.SetProperty("core", "repositoryformatversion", "0")
	configIni.SetProperty("core", "filemode", "false")
	configIni.SetProperty("core", "bare", "false")
	if worktree != ".." && worktree != "" {
		configIni.SetProperty("core", "worktree", worktree)
	}
	return configIni.String()
}
