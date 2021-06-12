package iniparse

import (
    "fmt"
    "strings"
)

type Section map[string]string
type IniFile map[string]Section

func NewIni() IniFile {
    return make(IniFile)
}

func (iFile *IniFile) SetProperty (section string, key string, value string) {
    _, ok := (*iFile)[section]
    if !ok {
        (*iFile)[section] =  make(Section)
    }
    (*iFile)[section][key] = value
}

func (iFile *IniFile) String () string {
    myString := "" 
    for section, maps := range *iFile {
        if section != "" {
            myString += "[" + section + "]\n"
        }
        for key,value := range maps {
            myString += fmt.Sprintf("%s = %s\n", key, value)
        }
    }
    return strings.Trim(myString, "\n")
}

func EqualInis(ini1 *IniFile, ini2 *IniFile) bool {
    if len(*ini1) != len (*ini2) {
        return false
    }
    for section, ini1_entries := range *ini1 {
        ini2_entries, ok := (*ini2)[section]
       if !ok {
            return false
        } 
        if len(ini1_entries) != len(ini2_entries) {
            return false
        }
        
        for key := range ini1_entries {
            if ini1_entries[key] != ini2_entries[key] {
                return false
            }
        }
    }
    return true
}
