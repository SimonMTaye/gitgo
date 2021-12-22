package config

import (
	"github.com/SimonMTaye/gitgo/iniparse"
	"os"
	"path/filepath"
)

const SystemPath = "/etc/gitconfig"
const GlobalPathFirst = "~/.gitconfig"
const GlobalPathSecond = "~/.config/git/config"

type Config struct {
	System *iniparse.IniFile
	Global *iniparse.IniFile
	Local  *iniparse.IniFile
	All    *iniparse.IniFile
}

// LoadGlobalConfig Loads the config data not inlcuding data for the local repository
func LoadGlobalConfig() (*Config, error) {
	systemIni, err := findAndRead(SystemPath)
	if err != nil {
		return nil, err
	}
	globalIni, err := findAndRead(GlobalPathFirst)
	if err != nil {
		return nil, err
	}

	return &Config{
			System: systemIni,
			Global: globalIni,
			Local:  &iniparse.IniFile{},
			All:    iniparse.MergeInis(systemIni, globalIni)},
		nil
}

// LoadConfig Read the config files used by git
func LoadConfig(localpath string) (*Config, error) {
	localfile, err := os.Open(localpath)
	if err != nil {
		return nil, err
	}
	localIni, err := iniparse.ParseIni(localfile)
	if err != nil {
		return nil, err
	}
	gConfig, err := LoadGlobalConfig()
	if err != nil {
		return nil, err
	}
	gConfig.Local = &localIni
	gConfig.All = iniparse.MergeInis(iniparse.MergeInis(gConfig.Global, gConfig.Local), &localIni)
	return gConfig, nil
}

// Reduce duplication in LoadConfig.
// Finds a path and parses it as an ini file. If the path is not found, no error will
// be returned.
// Errors are returned when there is an error parsing the file or determining the absoulte
// path given
func findAndRead(path string) (*iniparse.IniFile, error) {
	// Change path to absoulte
	absPath, err := filepath.Abs(path)
	// I do not know when this would return an error but it is not handled here so it will
	// be propagated
	if err != nil {
		return nil, err
	}
	file, err := os.Open(absPath)
	if err == nil {
		ini, err := iniparse.ParseIni(file)
		// Parsing error means there is something with ini or code; not handled here
		if err != nil {
			return nil, err
		}
		return &ini, nil
	} else {
		return &iniparse.IniFile{}, nil
	}
}
