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
        fmt.Printf("%s", parsedFile.String())
    }

    expectedIni := make(IniFile)
    expectedIni.SetProperty("Header", "title", "initest")
    expectedIni.SetProperty("Header","purpose",  "straightforward")

    if !EqualInis(&parsedFile, &expectedIni) {
        t.Errorf("IniFile is different from expected")
        fmt.Println("Expected:")
        fmt.Printf("%s", expectedIni.String())
        fmt.Println("Got:")
        fmt.Printf("%s", parsedFile.String())
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
        fmt.Printf("%s", parsedFile.String())
    }

    expectedIni := make(IniFile)
    expectedIni.SetProperty("Header", "title", "initest")
    expectedIni.SetProperty("Header","purpose",  "straightforward")

    if !EqualInis(&parsedFile, &expectedIni) {
        t.Errorf("IniFile is different from expected")
        fmt.Println("Expected:")
        fmt.Printf("%s", expectedIni.String())
        fmt.Println("Got:")
        fmt.Printf("%s", parsedFile.String())
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
        fmt.Printf("%s", parsedFile.String())
    }

    expectedIni := make(IniFile)
    expectedIni.SetProperty("Header", "title", "initest")
    expectedIni.SetProperty("Header","purpose",  "straightforward")
    expectedIni.SetProperty("Footer","purpose" , "straightforward")
    expectedIni.SetProperty("Footer","copyright" , "mit")

    if !EqualInis(&parsedFile, &expectedIni) {
        t.Errorf("IniFile is different from expected")
        fmt.Println("Expected:")
        fmt.Printf("%s", expectedIni.String())
        fmt.Println("Got:")
        fmt.Printf("%s", parsedFile.String())
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
        fmt.Printf("%s", parsedFile.String())
    }

    expectedIni := make(IniFile)
    expectedIni.SetProperty("","title" , "initest")
    expectedIni.SetProperty("","purpose" , "straightforward")

    if !EqualInis(&parsedFile, &expectedIni) {
        t.Errorf("IniFile is different from expected")
        fmt.Println("Expected:")
        fmt.Printf("%s", expectedIni.String())
        fmt.Println("Got:")
        fmt.Printf("%s", parsedFile.String())
    }
}

func TestIniToString (t *testing.T) {
    expectedIniString :=  "[Header]\n" +
                    "title = initest\n" +
                    "purpose = straightforward\n" +
                    "[Footer]\n" +
                    "copyright = mit\n" +
                    "purpose = straightforward"
    sampleIni := make(IniFile)
    sampleIni.SetProperty("Header", "title", "initest")
    sampleIni.SetProperty("Header","purpose",  "straightforward")
    sampleIni.SetProperty("Footer","copyright" , "mit")
    sampleIni.SetProperty("Footer","purpose" , "straightforward")

    if expectedIniString != sampleIni.String() {
        t.Errorf ("\nExpected:\n%s\nGot:\n%s\n", expectedIniString, sampleIni.String())
    }
    
    expectedIniString = "title = initest\n" +
                        "purpose = straightforward\n" +
                        "[Footer]\n" +
                        "copyright = mit\n" +
                        "purpose = straightforward"
    sampleIni = make(IniFile)
    sampleIni.SetProperty("", "title", "initest")
    sampleIni.SetProperty("","purpose",  "straightforward")
    sampleIni.SetProperty("Footer","copyright" , "mit")
    sampleIni.SetProperty("Footer","purpose" , "straightforward")

    if expectedIniString != sampleIni.String() {
        t.Errorf ("\nExpected:\n%s\nGot:\n%s\n", expectedIniString, sampleIni.String())
    }

}


