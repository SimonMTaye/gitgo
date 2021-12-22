package repo

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/SimonMTaye/gitgo/iniparse"
)

func TestDefaultConfig(t *testing.T) {
	expectedIni := make(iniparse.IniFile)
	expectedIni.SetProperty("core", "repositoryformatversion", "0")
	expectedIni.SetProperty("core", "filemode", "false")
	expectedIni.SetProperty("core", "bare", "false")

	defaultIni, err := iniparse.ParseIni(strings.NewReader(defaultConfig("")))
	if err != nil {
		t.Errorf("Unexpected error when parsing ini:\n%s", err.Error())
	}
	if !iniparse.EqualInis(&expectedIni, &defaultIni) {
		t.Errorf("Generated config is different from expected\nExpected:\n%s\nGot:\n%s",
			expectedIni.String(), defaultIni.String())
	}

	expectedIni.SetProperty("core", "worktree", "random")
	defaultIni, err = iniparse.ParseIni(strings.NewReader(defaultConfig("random")))
	if err != nil {
		t.Errorf("Unexpected error when parsing ini:\n%s", err.Error())
	}
	if !iniparse.EqualInis(&expectedIni, &defaultIni) {
		t.Errorf("Generated config is different from expected\nExpected:\n%s\nGot:\n%s",
			expectedIni.String(), defaultIni.String())
	}
}

func TestCreateRepoWithDefaultParams(t *testing.T) {
	tmpDir := t.TempDir()

	err := CreateRepo(tmpDir, "", "")
	if err != nil {
		t.Errorf("CreateRepo returned unexpected error:\n%s", err.Error())
	}

	dirs, err := os.ReadDir(tmpDir)
	if !exists(dirs, ".git") {
		t.Errorf("'.git' directory not found after running CreateRepo")
	}

	if err != nil {
		t.Errorf("Unexpected error when reading temp directory:\n%s", err.Error())
	}

	gitDir := path.Join(tmpDir, ".git")
	dirs, _ = os.ReadDir(gitDir)

	if !exists(dirs, "objects") {
		t.Errorf("Expected 'objects' directory to exist in '.git'")
	}

	if !exists(dirs, "branches") {
		t.Errorf("Expected 'branches' directory to exist in '.git'")
	}

	if !exists(dirs, "refs") {
		t.Errorf("Expected 'refs' directory to exist in '.git'")
	}

	refsDir := path.Join(gitDir, "refs")
	refDirItems, _ := os.ReadDir(refsDir)

	if !exists(refDirItems, "heads") {
		t.Errorf("Expected 'heads' directory to exist in 'refs'")
	}

	if !exists(refDirItems, "tags") {
		t.Errorf("Expected 'tags' directory to exist in 'refs'")
	}

	if !exists(dirs, "HEAD") {
		t.Errorf("Expected 'HEAD' file to exist in '.git'")
	}

	if !exists(dirs, "description") {
		t.Errorf("Expected 'refs' directory to exist in '.git'")
	}
	descriptionContents, err := os.ReadFile(path.Join(gitDir, "description"))
	if err != nil {
		t.Errorf("Unexpected error when reading 'description' file:\n%s", err.Error())
	}

	if string(descriptionContents) != EmptyDescription {
		t.Errorf("'description' file\nExpected:\n%s\nGot:\n%s", EmptyDescription,
			string(descriptionContents))
	}

	if !exists(dirs, "config") {
		t.Errorf("Expected 'config' directory to exist in '.git'")
	}
	configContents, err := os.ReadFile(path.Join(gitDir, "config"))
	if err != nil {
		t.Errorf("Unexpected error when reading 'config' file:\n%s", err.Error())
	}
	configIni, err := iniparse.ParseIni(strings.NewReader(string(configContents)))
	if err != nil {
		t.Errorf("Unexpected error when parsing 'config' file as ini:\n%s", err.Error())
	}

	expectedIni, err := iniparse.ParseIni(strings.NewReader(defaultConfig("")))
	if err != nil {
		t.Errorf("Unexpected error when parsing default config as ini:\n%s", err.Error())
	}

	if !iniparse.EqualInis(&configIni, &expectedIni) {
		t.Errorf("'config' file\nExpected:\n%s\nGot:\n%s", expectedIni.String(),
			configIni.String())
	}
}

func TestCreateRepoWorktreeAndDescription(t *testing.T) {
	tmpDir := t.TempDir()

	sampleDescription := "hello world"
	worktreeDir := path.Join(tmpDir, "random")
	sampleConfig := defaultConfig(worktreeDir)
	err := CreateRepo(tmpDir, sampleDescription, worktreeDir)
	if err != nil {
		t.Errorf("CreateRepo returned unexpected error:\n%s", err.Error())
	}
	configContents, err := os.ReadFile(path.Join(tmpDir, ".git", "config"))
	if err != nil {
		t.Errorf("Unexpected error when reading 'config' file:\n%s", err.Error())
	}
	configIni, err := iniparse.ParseIni(strings.NewReader(string(configContents)))
	if err != nil {
		t.Errorf("Unexpected error when parsing 'config' file as ini:\n%s", err.Error())
	}

	expectedIni, err := iniparse.ParseIni(strings.NewReader(sampleConfig))
	if err != nil {
		t.Errorf("Unexpected error when parsing sample string as ini:\n%s", err.Error())
	}

	if !iniparse.EqualInis(&configIni, &expectedIni) {
		t.Errorf("'config' file\nExpected:\n%s\nGot:\n%s", expectedIni.String(),
			configIni.String())
	}
}
