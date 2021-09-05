package config

import (
    "github.com/SimonMTaye/gitgo/iniparse"
    "os"
    "path/filepath"
)

const SYSTEM_PATH = "/etc/gitconfig"
const GLOBAL_PATH_FIRST = "~/.gitconfig"
const GLOBAL_PATH_SECOND = "~/.config/git/config"


type Config struct {
    System  *iniparse.IniFile
    Global  *iniparse.IniFile
    Local   *iniparse.IniFile
    All     *iniparse.IniFile
}
// Read the config files used by git
// Will throw an error if there was a problem parsing any of the config files
// Will NOT throw an error if any of the files are not found (expect the localfile; panic on localfile not being present)
func LoadConfig(localpath string) (*Config, error) {
    localfile, err := os.Open(localpath)
    if err != nil {
        return nil, err
    }
    localIni, err := iniparse.ParseIni(localfile)
    if err != nil {
        return nil, err
    }
    systemIni, err := findAndRead(SYSTEM_PATH)
    if err != nil {
        return nil, err
    }
    globalIni, err := findAndRead(GLOBAL_PATH_FIRST)
    if err != nil {
        return nil, err
    }

    return &Config{System: systemIni, 
            Global: globalIni, 
            Local: &localIni, 
            All: iniparse.MergeInis(iniparse.MergeInis(systemIni, globalIni), &localIni)},
        nil
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
        return nil, nil
    }
}
