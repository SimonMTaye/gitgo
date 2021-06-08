package iniparse

import (
    "fmt"
    "testing"
    "strings"
)


func TestSimpleIni (t *testing.T) {
    fullIniFile :=  "[Header]\n" +
                    "title=initest\n" +
                    "purpose=straightforward"
    stringReader := strings.NewReader(fullIniFile)
    parsedFile, err := ParseIni(stringReader)
    if err != nil {
        t.Errorf("Got unexpected error from parser function")
        fmt.Println("Function output:")
        outputIniFile(&parsedFile)
    }

    expectedIni := make(map[string]map[string]string)
    expectedIni["Header"] = make(map[string]string)
    expectedIni["Header"]["title"] = "initest"
    expectedIni["Header"]["purpose"] = "straightforward"

    if !equalInis(&parsedFile, &expectedIni) {
        t.Errorf("IniFile is different from expected")
        fmt.Println("Expected:")
        outputIniFile(&expectedIni)
        fmt.Println("Got:")
        outputIniFile(&parsedFile)
    }

}

func TestSimpleIniWithComments (t *testing.T) {
    fullIniFile :=  "[Header]\n" +
                    "# Some random comment\n" +
                    "title=initest\n" + 
                    "; Some other comment\n" +
                    "purpose=straightforward"
    stringReader := strings.NewReader(fullIniFile)
    parsedFile, err := ParseIni(stringReader)
    if err != nil {
        t.Errorf("Got unexpected error from parser function")
        fmt.Println("Function output:")
        outputIniFile(&parsedFile)
    }

    expectedIni := make(map[string]map[string]string)
    expectedIni["Header"] = make(map[string]string)
    expectedIni["Header"]["title"] = "initest"
    expectedIni["Header"]["purpose"] = "straightforward"

    if !equalInis(&parsedFile, &expectedIni) {
        t.Errorf("IniFile is different from expected")
        fmt.Println("Expected:")
        outputIniFile(&expectedIni)
        fmt.Println("Got:")
        outputIniFile(&parsedFile)
    }

}

func TestMultipleSections (t *testing.T) {
    fullIniFile :=  "[Header]\n" +
                    "title=initest\n" +
                    "purpose=straightforward\n" +
                    "[Footer]\n" +
                    "copyright=mit\n" +
                    "purpose=straightforward"

    stringReader := strings.NewReader(fullIniFile)
    parsedFile, err := ParseIni(stringReader)
    if err != nil {
        t.Errorf("Got unexpected error from parser function")
        fmt.Println("Function output:")
        outputIniFile(&parsedFile)
    }

    expectedIni := make(map[string]map[string]string)
    expectedIni["Header"] = make(map[string]string)
    expectedIni["Header"]["title"] = "initest"
    expectedIni["Header"]["purpose"] = "straightforward"
    expectedIni["Footer"] = make(map[string]string)
    expectedIni["Footer"]["purpose"] = "straightforward"
    expectedIni["Footer"]["copyright"] = "mit"

    if !equalInis(&parsedFile, &expectedIni) {
        t.Errorf("IniFile is different from expected")
        fmt.Println("Expected:")
        outputIniFile(&expectedIni)
        fmt.Println("Got:")
        outputIniFile(&parsedFile)
    }
}

func TestDefaultSection(t *testing.T) {
    fullIniFile :=  "title=initest\n" +
                    "purpose=straightforward"

    stringReader := strings.NewReader(fullIniFile)
    parsedFile, err := ParseIni(stringReader)
    if err != nil {
        t.Errorf("Got unexpected error from parser function")
        fmt.Println("Function output:")
        outputIniFile(&parsedFile)
    }

    expectedIni := make(map[string]map[string]string)
    expectedIni[""] = make(map[string]string)
    expectedIni[""]["title"] = "initest"
    expectedIni[""]["purpose"] = "straightforward"

    if !equalInis(&parsedFile, &expectedIni) {
        t.Errorf("IniFile is different from expected")
        fmt.Println("Expected:")
        outputIniFile(&expectedIni)
        fmt.Println("Got:")
        outputIniFile(&parsedFile)
    }
}

func equalInis(ini1 *IniFile, ini2 *IniFile) bool {
    if len(*ini1) != len (*ini2) {
        return false
    }
    for section, ini1_entries := range *ini1 {
        ini2_entries, ok := (*ini2)[section]
        if len(ini1_entries) != len(ini2_entries) {
            return false
        }
        if !ok {
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

func outputIniFile(i *IniFile) {
    for section, maps := range *i {
        fmt.Println(section)
        for key,value := range maps {
            fmt.Printf("\t%s : %s\n", key, value)
        }
    }

}

