package iniparse

import (
    "fmt"
    "strings"
)

type Section map[string]string
type IniFile map[string]Section

//Allocate a new empty ini and return it
func NewIni() IniFile {
    return make(IniFile)
}

//Set a key-value pair under a certain ini file section
//If the section doesn't exist, it will be created
func (iFile *IniFile) SetProperty (section string, key string, value string) {
    _, ok := (*iFile)[section]
    if !ok {
        (*iFile)[section] =  make(Section)
    }
    (*iFile)[section][key] = value
}

//Converts an ini-file map into a string that is typically found in a .ini file
func (iFile *IniFile) String () string {
    myString := "" 

    defaultSection, ok := (*iFile)[""]
    if ok {
        myString += defaultSection.String()
    }

    for section, sectionMap := range *iFile {
        if section == "" {
            continue
        }
        myString += "[" + section + "]\n"
        myString += sectionMap.String()
    }
    return strings.Trim(myString, "\n")
}

//Returns key-value pairs in the section separated by new-lines
//Intended as a helper function for IniFile.String()
func (section *Section) String () string {
    sectionString := ""
    for key,value := range *section {
        sectionString += fmt.Sprintf("%s = %s\n", key, value)
    }
    return sectionString
}

//Compares two IniFiles and returns 'true' if both Inis have identical
//key-value pairs
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
// Merge two ini, into one.
// For properties in both inis, the option in ini2 will be selected
func MergeInis(ini1 *IniFile, ini2 *IniFile) *IniFile {
    newIni := &IniFile{}
    // Duplicate all the sections in ini1
    for section, sectionMap := range *ini1 {
        newSection := Section{}
        for k, v := range sectionMap {
            newSection[k] = v
        }
        (*newIni)[section] = newSection
    }
    // Duplicate stuff in ini2.
    // Will replace stuff in ini1
    // If section already exists, replace any old properties with new ones while keeping
    // the rest
    // If it doesn't exist, create it
    for section, sectionMap := range *ini1 {
        selectedSection, ok := (*newIni)[section]
        if !ok {
            selectedSection = Section{}
        }
        for k, v := range sectionMap {
            selectedSection[k] = v
        }
        (*newIni)[section] = selectedSection
    }
    return newIni
}
