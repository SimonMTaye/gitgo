package iniparse

import "fmt"

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
    return myString
}



